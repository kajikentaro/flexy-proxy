package proxy

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"

	"github.com/kajikentaro/flexy-proxy/loggers"
	"github.com/kajikentaro/flexy-proxy/models"

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

var regLast443 = regexp.MustCompile(":443$")

func removeSuffix443FromHostName(u url.URL) *url.URL {
	// remove last ":443" which is added automatically by goproxy
	u.Host = regLast443.ReplaceAllString(u.Host, "")
	return &u
}

func (p *Proxy) onRequest(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
	req.URL = removeSuffix443FromHostName(*req.URL)

	handler, matchedUrl, err := p.router.GetHandler(req.URL)
	// if the request doesn't match any routes
	if errors.Is(err, models.ErrRouteNotFound) {
		if p.config.DefaultRoute.DenyAccess {
			content := fmt.Sprintf("%s is out of routes", req.URL.String())
			return req, goproxy.NewResponse(req, goproxy.ContentTypeText, http.StatusForbidden, content)
		}
		return req, nil
	}
	if err != nil {
		p.handleProxyRuntimeError(req, err)
	}

	resWriter := NewResponseWriter(req)
	resWriter.Header().Add("flexy-proxy", fmt.Sprintf("matched URL: %s", matchedUrl))

	// logging
	args := []interface{}{
		"request URL", req.URL.String(),
		"matched URL", matchedUrl,
		"type", handler.GetType(),
	}
	for k, v := range handler.GetResponseInfo() {
		args = append(args, k, v)
	}

	p.logger.Info("request matched a route", args...)
	handler.ServeHTTP(resWriter, req)
	return req, resWriter.Response
}

type Proxy struct {
	router models.Router
	logger *loggers.Logger
	config *Config
}

type Config struct {
	DefaultRoute DefaultRoute
	AlwaysMitm   bool
}

type DefaultRoute struct {
	Proxy      *url.URL
	DenyAccess bool `yaml:"deny_access"`
}

func SetupProxy(router models.Router, logger *loggers.Logger, config *Config) *goproxy.ProxyHttpServer {
	ps := &Proxy{
		router: router,
		logger: logger,
		config: config,
	}
	logger.Info("Proxy has been configured", "route pattern length", len(router.GetUrlList()))
	return ps.getProxyHttpServer()
}

func (p *Proxy) getProxyHttpServer() *goproxy.ProxyHttpServer {
	proxy := goproxy.NewProxyHttpServer()
	proxy.Logger = GenLoggerForProxy(p.logger)
	proxy.Verbose = true

	if p.config.AlwaysMitm {
		proxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)
	} else {
		hosts := p.router.GetHttpsHostList()
		proxy.OnRequest(goproxy.ReqHostIs(hosts...)).HandleConnect(goproxy.AlwaysMitm)
	}

	proxy.OnRequest().DoFunc(p.onRequest)

	if p.config.DefaultRoute.DenyAccess {
		proxy.OnRequest().HandleConnect(goproxy.AlwaysReject)
	}

	if p.config.DefaultRoute.Proxy != nil {
		// proxy which is used when "AlwaysMitm" hits
		proxy.Tr = &http.Transport{
			Proxy: func(req *http.Request) (*url.URL, error) {
				return p.config.DefaultRoute.Proxy, nil
			},
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		// proxy which is used when "AlwaysMitm" doesn't hits
		proxy.ConnectDial = proxy.NewConnectDialToProxy(p.config.DefaultRoute.Proxy.String())
	}

	return proxy
}
