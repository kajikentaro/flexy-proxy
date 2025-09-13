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
	"github.com/kajikentaro/flexy-proxy/utils/cache"

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

	successInfo, res, err := p.router.TryRoundTrip(req)
	// if the request doesn't match any routes
	if errors.Is(err, models.ErrRouteNotFound) {
		if p.defaultRoute.DenyAccess {
			content := fmt.Sprintf("%s is out of routes", req.URL.String())
			return req, goproxy.NewResponse(req, goproxy.ContentTypeText, http.StatusForbidden, content)
		}
		return req, nil
	} else if err != nil {
		return p.handleProxyRuntimeError(req, err)
	}

	// logging
	var args []interface{}
	for k, v := range successInfo {
		args = append(args, k, v)
	}
	p.logger.Info("Request matched a route", args...)

	return req, res
}

type Proxy struct {
	defaultRoute    DefaultRoute
	alwaysMitm      bool
	certificate     *tls.Certificate
	logger          *loggers.Logger
	router          models.Router
	hostToCertCache *cache.LRUCache[string, *tls.Config]
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

// Around 100 MB (1 certificate is 8KB)
var MAX_TLS_CERT_CACHE_SIZE = 10000000

func SetupProxy(config *Config) *goproxy.ProxyHttpServer {
	p := &Proxy{
		defaultRoute:    config.DefaultRoute,
		alwaysMitm:      config.AlwaysMitm,
		certificate:     config.Certificate,
		logger:          config.Logger,
		router:          config.Router,
		hostToCertCache: cache.NewLRUCache[string, *tls.Config](MAX_TLS_CERT_CACHE_SIZE),
	}
	config.Logger.Info("Proxy has been configured")
	return p.getProxyHttpServer()
}

// TODO: if we can use *goproxy.ProxyCtx.certStore, we can simplify this code
func (p *Proxy) getTlsConfig(host string, ctx *goproxy.ProxyCtx) (*tls.Config, error) {
	caCert := p.certificate
	if caCert == nil {
		caCert = &goproxy.GoproxyCa
	}

	if cert, ok := p.hostToCertCache.Load(host); ok {
		return cert, nil
	} else {
		cert, err := goproxy.TLSConfigFromCA(caCert)(host, ctx)
		if err != nil {
			return nil, err
		}
		p.hostToCertCache.Store(host, cert)
		return cert, nil
	}
}

func (p *Proxy) eavesDropHttp(host string, ctx *goproxy.ProxyCtx) (*goproxy.ConnectAction, string) {
	return &goproxy.ConnectAction{
		Action:    goproxy.ConnectMitm,
		TLSConfig: p.getTlsConfig,
	}, host
}

func (p *Proxy) getProxyHttpServer() *goproxy.ProxyHttpServer {
	proxy := goproxy.NewProxyHttpServer()
	proxy.Logger = GenLoggerForProxy(p.logger)
	proxy.Verbose = true

	if p.alwaysMitm {
		proxy.OnRequest().HandleConnectFunc(p.eavesDropHttp)
	} else {
		hosts := p.router.GetHttpsHostList()
		proxy.OnRequest(goproxy.ReqHostIs(hosts...)).HandleConnectFunc(p.eavesDropHttp)
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
