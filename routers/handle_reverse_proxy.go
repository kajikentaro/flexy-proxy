package routers

import (
	"net/http"
	"net/url"

	"github.com/kajikentaro/flexy-proxy/models/rewrite"
	"github.com/kajikentaro/flexy-proxy/utils"
)

func NewReverseProxyTransport(proxyUrl *url.URL, urlRewriter *rewrite.Rewrite, insecureCipherSuites bool) http.RoundTripper {
	return &ReverseProxyTransport{
		proxyUrl:             proxyUrl,
		urlRewriter:          urlRewriter,
		insecureCipherSuites: insecureCipherSuites,
	}
}

type ReverseProxyTransport struct {
	proxyUrl             *url.URL
	urlRewriter          *rewrite.Rewrite
	insecureCipherSuites bool
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

	t := utils.GetTransport(c.insecureCipherSuites, c.proxyUrl)
	res, err := t.RoundTrip(rr)
	if err != nil {
		return nil, err
	}

	res.Header.Set(HEADER_RESPONSE_TYPE, "rewrite")
	res.Header.Set(HEADER_REWRITE_TO, forwardUrl.String())
	return res, nil
}
