package routers

import (
	"crypto/tls"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/kajikentaro/elastic-proxy/models"
)

func NewReverseProxyHandler(statusCode int, contentType string, forwardUrl *url.URL, proxyUrl *url.URL) models.ReverseProxyHandler {
	return &reverseProxyHandle{
		statusCode:  statusCode,
		contentType: contentType,
		forwardUrl:  forwardUrl,
		proxyUrl:    proxyUrl,
	}
}

type reverseProxyHandle struct {
	forwardUrl  *url.URL
	statusCode  int
	contentType string
	proxyUrl    *url.URL
}

func (c *reverseProxyHandle) Handler(w http.ResponseWriter, r *http.Request) {
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

func (c *reverseProxyHandle) ForwardUrl() string {
	return c.forwardUrl.String()
}
