package routers

import (
	"crypto/tls"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/kajikentaro/elastic-proxy/models"
)

func NewHandleReverseProxy(forwardUrl *url.URL, proxyUrl *url.URL) models.Handler {
	return &ReverseProxyHandle{
		forwardUrl: forwardUrl,
		proxyUrl:   proxyUrl,
	}
}

type ReverseProxyHandle struct {
	forwardUrl *url.URL
	proxyUrl   *url.URL
}

func (c *ReverseProxyHandle) Handle(w http.ResponseWriter, r *http.Request) {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	if c.proxyUrl != nil {
		t.Proxy = func(req *http.Request) (*url.URL, error) {
			return c.proxyUrl, nil
		}
	}

	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL = c.forwardUrl
		},
		Transport: t,
	}

	rr := r.Clone(r.Context())
	// NOTE:
	// we should update host manually; otherwise, the original host remains
	rr.Host = c.forwardUrl.Host
	proxy.ServeHTTP(w, rr)
}

func (c *ReverseProxyHandle) GetType() string {
	return "reverse proxy"
}

func (c *ReverseProxyHandle) GetResponseInfo() map[string]string {
	return map[string]string{
		"forward url": c.forwardUrl.String(),
	}
}
