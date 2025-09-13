package routers

import (
	"crypto/tls"
	"net/http"
	"net/url"

	"github.com/kajikentaro/flexy-proxy/models/rewrite"
)

func NewReverseProxyTransport(proxyUrl *url.URL, urlRewriter *rewrite.Rewrite) http.RoundTripper {
	return &ReverseProxyTransport{
		proxyUrl:    proxyUrl,
		urlRewriter: urlRewriter,
	}
}

type ReverseProxyTransport struct {
	proxyUrl    *url.URL
	urlRewriter *rewrite.Rewrite
}

func (c *ReverseProxyTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	forwardUrl := r.URL
	if c.urlRewriter != nil {
		var err error
		if forwardUrl, err = c.urlRewriter.Replace(r.URL); err != nil {
			return nil, err
		}
	}

	rr := r.Clone(r.Context())
	rr.URL = forwardUrl
	// NOTE:
	// we should update host manually; otherwise, the original host remains
	rr.Host = forwardUrl.Host

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

	res.Header.Set(HEADER_RESPONSE_TYPE, "rewrite")
	res.Header.Set(HEADER_REWRITE_TO, forwardUrl.String())
	return res, nil
}
