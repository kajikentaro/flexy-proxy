package proxy

import (
	"fmt"
	"go-proxy/loggers"
	"go-proxy/models"
	"net/http"

	"github.com/elazarl/goproxy"
)

func (p *Proxy) handleProxyRuntimeError(req *http.Request, err error) (*http.Request, *http.Response) {
	logContent := fmt.Sprintf("proxy runtime error: %s", err.Error())
	p.logger.Error(logContent)
	res := goproxy.NewResponse(
		req,
		goproxy.ContentTypeText,
		http.StatusInternalServerError,
		logContent,
	)
	return req, res
}

func (p *Proxy) onRequest(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
	handler, matchedUrl, err := p.router.GetHandler(req.URL)
	if err != nil {
		p.handleProxyRuntimeError(req, err)
	}

	resWriter := NewResponseWriter(req)
	resWriter.Header().Add("Elastic-Proxy", fmt.Sprintf("matched URL: %s", matchedUrl))

	if h := handler.ReverseProxy; h != nil {
		p.logger.Info("routed to the URL", "request URL", req.URL.String(), "forward URL", h.ForwardUrl())
		h.Handler(resWriter, req)
		return req, resWriter.res
	}

	if h := handler.File; h != nil {
		p.logger.Info("routed to the file", "request URL", req.URL.String(), "file", h.FilePath())
		h.Handler(resWriter, req)
		return req, resWriter.res
	}

	if h := handler.Content; h != nil {
		p.logger.Info("routed to the content", "request URL", req.URL.String(), "content", h.Content())
		h.Handler(resWriter, req)
		return req, resWriter.res
	}

	// if the request doesn't match any routes

	if p.defaultRoute.DenyAccess {
		content := fmt.Sprintf("%s is out of routes", req.URL.String())
		return req, goproxy.NewResponse(req, goproxy.ContentTypeText, http.StatusForbidden, content)
	}
	return req, nil
}

type Proxy struct {
	router       models.Router
	logger       *loggers.Logger
	defaultRoute *models.DefaultRoute
}

func SetupProxy(router models.Router, logger *loggers.Logger, defaultRoute *models.DefaultRoute) *goproxy.ProxyHttpServer {
	ps := &Proxy{
		router:       router,
		logger:       logger,
		defaultRoute: defaultRoute,
	}
	return ps.getProxyHttpServer()
}

func (p *Proxy) getProxyHttpServer() *goproxy.ProxyHttpServer {
	proxy := goproxy.NewProxyHttpServer()
	proxy.Logger = GenLoggerForProxy(p.logger)
	proxy.Verbose = true

	hosts := p.router.GetHttpsHostList()
	proxy.OnRequest(goproxy.ReqHostIs(hosts...)).HandleConnect(goproxy.AlwaysMitm)

	proxy.OnRequest().DoFunc(p.onRequest)

	if p.defaultRoute.DenyAccess {
		proxy.OnRequest().HandleConnect(goproxy.AlwaysReject)
	}

	if proxyUrl := p.defaultRoute.ProxyUrl; proxyUrl != "" {
		proxy.ConnectDial = proxy.NewConnectDialToProxy(proxyUrl)
	}

	return proxy
}
