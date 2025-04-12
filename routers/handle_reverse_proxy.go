package routers

import (
	"crypto/tls"
	"net/http"
	"net/url"

	"github.com/kajikentaro/flexy-proxy/models"
)

func NewReverseProxyTransport(forwardUrl *url.URL, proxyUrl *url.URL) models.RoundTripper {
	return &ReverseProxyTransport{
		forwardUrl: forwardUrl,
		proxyUrl:   proxyUrl,
	}
}

type ReverseProxyTransport struct {
	forwardUrl *url.URL
	proxyUrl   *url.URL
}

func (c *ReverseProxyTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	rr := r.Clone(r.Context())
	rr.URL = c.forwardUrl
	// NOTE:
	// we should update host manually; otherwise, the original host remains
	rr.Host = c.forwardUrl.Host

	t := http.DefaultTransport.(*http.Transport).Clone()
	t.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	if c.proxyUrl != nil {
		t.Proxy = func(req *http.Request) (*url.URL, error) {
			return c.proxyUrl, nil
		}
	}

	res, err := t.RoundTrip(rr)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (c *ReverseProxyTransport) GetType() string {
	return "reverse proxy"
}

func (c *ReverseProxyTransport) GetResponseInfo() map[string]string {
	proxy := "None"
	if c.proxyUrl != nil {
		proxy = c.proxyUrl.String()
	}
	return map[string]string{
		"forward_url": c.forwardUrl.String(),
		"proxy":       proxy,
	}
}
