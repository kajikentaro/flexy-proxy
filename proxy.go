package proxy

import (
	"fmt"
	"go-proxy/loggers"
	"go-proxy/models"
	"net/http"
	"net/url"

	_ "embed"

	"github.com/elazarl/goproxy"
)

func isUrlSame(a *url.URL, b *url.URL) bool {
	if a.Scheme != b.Scheme {
		return false
	}
	if a.Hostname() != b.Hostname() {
		return false
	}
	if a.RawPath != b.RawPath {
		return false
	}
	if a.RawQuery != b.RawQuery {
		return false
	}
	if a.RawFragment != b.RawFragment {
		return false
	}
	return true
}

func (p *ProxySeed) handleProxyRuntimeError(req *http.Request, title string, message string) (*http.Request, *http.Response) {
	logContent := fmt.Sprintf("proxy runtime error: %s: %s", title, message)
	p.logger.Error(logContent)
	res := goproxy.NewResponse(
		req,
		goproxy.ContentTypeText,
		http.StatusInternalServerError,
		logContent,
	)
	return req, res
}

func (*ProxySeed) serveContent(req *http.Request, userStatusCode int, userContentType string, content string) (*http.Request, *http.Response) {
	contentType := goproxy.ContentTypeText
	if userContentType != "" {
		contentType = userContentType
	}
	statusCode := 200
	if userStatusCode != 0 {
		statusCode = userStatusCode
	}

	res := goproxy.NewResponse(req, contentType, statusCode, content)
	return req, res
}

func (*ProxySeed) serveFile(req *http.Request, userStatusCode int, userContentType string, fileName string) (*http.Request, *http.Response) {
	fileRes := NewFileResponse(req)
	res := fileRes.res
	http.ServeFile(fileRes, req, fileName)
	if userContentType != "" {
		// if contentType is specified, overwrite it
		res.Header["Content-Type"][0] = userContentType
	}
	statusCode := 200
	if userStatusCode != 0 {
		statusCode = userStatusCode
	}
	res.StatusCode = statusCode
	return req, res
}

func (p *ProxySeed) onRequest(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
	for _, route := range p.config.Routes {
		routeUrl, _ := url.Parse(route.Url)
		if !isUrlSame(routeUrl, req.URL) {
			// unmatched
			continue
		}
		// matched

		if route.Response.File != "" {
			// TODO use std out
			p.logger.Info("routed to the file", "request URL", req.URL.String(), "file", route.Response.File)
			return p.serveFile(req, route.Response.Status, route.Response.ContentType, route.Response.File)
		}

		if route.Response.Content != "" {
			p.logger.Info("routed to the content", "request URL", req.URL.String(), "content", route.Response.Content)
			return p.serveContent(req, route.Response.Status, route.Response.ContentType, route.Response.Content)
		}

		return p.handleProxyRuntimeError(req, "File or Content are not specified", "")
	}
	return req, nil
}

type ProxySeed struct {
	config *models.ProxyConfig
	logger *loggers.Logger
}

func SetupProxy(config *models.ProxyConfig, logger *loggers.Logger) (*goproxy.ProxyHttpServer, error) {
	// TODO: validate config file
	// i.e.: do not contain both content and file
	ps := &ProxySeed{
		config: config,
		logger: logger,
	}
	return ps.getProxyHttpServer()
}

func (p *ProxySeed) getProxyHttpServer() (*goproxy.ProxyHttpServer, error) {
	proxy := goproxy.NewProxyHttpServer()
	proxy.Logger = GenLoggerForProxy(p.logger)
	// proxy.Verbose = true

	for _, route := range p.config.Routes {
		routeUrl, err := url.Parse(route.Url)
		if err != nil {
			return nil, err
		}
		if routeUrl.Scheme == "https" {
			reqHost := fmt.Sprintf("%s:443", routeUrl.Host)
			proxy.OnRequest(goproxy.ReqHostIs(reqHost)).HandleConnect(goproxy.AlwaysMitm)
			continue
		}
		if routeUrl.Scheme == "http" {
			continue
		}
		return nil, fmt.Errorf("scheme of '%s' must be either 'http' or 'https'", route.Url)
	}

	proxy.OnRequest().DoFunc(p.onRequest)

	if p.config.DefaultRoute.ProxyUrl != "" {
		proxy.ConnectDial = proxy.NewConnectDialToProxy(p.config.DefaultRoute.ProxyUrl)
	}

	return proxy, nil
}
