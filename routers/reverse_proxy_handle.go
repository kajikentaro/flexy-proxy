package routers

import (
	"crypto/tls"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/kajikentaro/elastic-proxy/models"
)

func NewReverseProxyHandler(statusCode int, contentType string, forwardUrl *url.URL, proxyUrl *url.URL) models.Handler {
	return &ReverseProxyHandle{
		statusCode:  statusCode,
		contentType: contentType,
		forwardUrl:  forwardUrl,
		proxyUrl:    proxyUrl,
	}
}

type ReverseProxyHandle struct {
	forwardUrl  *url.URL
	statusCode  int
	contentType string
	proxyUrl    *url.URL
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
	proxy.ServeHTTP(w, r)

	if c.contentType != "" {
		// only if the contentType is specified, overwrite
		w.Header().Set("Content-Type", c.contentType)
	}

	if c.statusCode != 0 {
		// only if the statusCode is specified, overwrite
		w.WriteHeader(c.statusCode)
	}
}

func (c *ReverseProxyHandle) GetType() string {
	return "reverse proxy"
}

func (c *ReverseProxyHandle) GetResponseInfo() map[string]string {
	return map[string]string{
		"forward url": c.forwardUrl.String(),
	}
}
