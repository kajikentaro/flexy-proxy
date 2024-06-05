package routers

import (
	"go-proxy/models"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func NewReverseProxyHandler(statusCode int, contentType string, forwardUrl *url.URL) models.ReverseProxyHandler {
	return &reverseProxyHandle{
		statusCode:  statusCode,
		contentType: contentType,
		forwardUrl:  forwardUrl,
	}
}

type reverseProxyHandle struct {
	forwardUrl  *url.URL
	statusCode  int
	contentType string
}

func (c *reverseProxyHandle) Handler(w http.ResponseWriter, r *http.Request) {
	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL = c.forwardUrl
		},
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
