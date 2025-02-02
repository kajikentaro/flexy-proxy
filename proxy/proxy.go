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
		if p.defaultRoute.DenyAccess {
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
	defaultRoute DefaultRoute
	alwaysMitm   bool
	certificate  *tls.Certificate
	logger       *loggers.Logger
	router       models.Router
}

type Config struct {
	DefaultRoute DefaultRoute
	AlwaysMitm   bool
	Certificate  *tls.Certificate
	Logger       *loggers.Logger
	Router       models.Router
}

type DefaultRoute struct {
	Proxy      *url.URL
	DenyAccess bool `yaml:"deny_access"`
}

func SetupProxy(config *Config) *goproxy.ProxyHttpServer {
	p := &Proxy{
		defaultRoute: config.DefaultRoute,
		alwaysMitm:   config.AlwaysMitm,
		certificate:  config.Certificate,
		logger:       config.Logger,
		router:       config.Router,
	}
	config.Logger.Info("Proxy has been configured", "route pattern length", len(config.Router.GetUrlList()))
	return p.getProxyHttpServer()
}

func (p *Proxy) eavesDropHttp() goproxy.FuncHttpsHandler {
	if p.certificate == nil {
		return goproxy.AlwaysMitm
	}

	ca := &goproxy.ConnectAction{
		Action:    goproxy.ConnectMitm,
		TLSConfig: goproxy.TLSConfigFromCA(p.certificate),
	}
	return func(host string, ctx *goproxy.ProxyCtx) (*goproxy.ConnectAction, string) {
		return ca, host
	}
}

func (p *Proxy) getProxyHttpServer() *goproxy.ProxyHttpServer {
	proxy := goproxy.NewProxyHttpServer()
	proxy.Logger = GenLoggerForProxy(p.logger)
	proxy.Verbose = true

	if p.alwaysMitm {
		proxy.OnRequest().HandleConnect(p.eavesDropHttp())
	} else {
		hosts := p.router.GetHttpsHostList()
		proxy.OnRequest(goproxy.ReqHostIs(hosts...)).HandleConnect(p.eavesDropHttp())
	}

	proxy.OnRequest().DoFunc(p.onRequest)

	if p.defaultRoute.DenyAccess {
		proxy.OnRequest().HandleConnect(goproxy.AlwaysReject)
	}

	if p.defaultRoute.Proxy != nil {
		// proxy which is used when "AlwaysMitm" hits
		proxy.Tr = &http.Transport{
			Proxy: func(req *http.Request) (*url.URL, error) {
				return p.defaultRoute.Proxy, nil
			},
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		// proxy which is used when "AlwaysMitm" doesn't hits
		proxy.ConnectDial = proxy.NewConnectDialToProxy(p.defaultRoute.Proxy.String())
	}

	return proxy
}
