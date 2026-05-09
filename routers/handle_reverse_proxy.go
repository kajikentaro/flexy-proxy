package routers

import (
	"net/http"
	"net/textproto"
	"net/url"

	"github.com/kajikentaro/flexy-proxy/models/rewrite"
	"github.com/kajikentaro/flexy-proxy/utils"
)

func NewReverseProxyTransport(proxyUrl *url.URL, urlRewriter *rewrite.Rewrite, insecureCipherSuites bool, reqHeaders map[string]string) http.RoundTripper {
	return &ReverseProxyTransport{
		proxyUrl:             proxyUrl,
		urlRewriter:          urlRewriter,
		insecureCipherSuites: insecureCipherSuites,
		reqHeaders:           reqHeaders,
	}
}

type ReverseProxyTransport struct {
	proxyUrl             *url.URL
	urlRewriter          *rewrite.Rewrite
	insecureCipherSuites bool
	reqHeaders           map[string]string
}

func (c *ReverseProxyTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	rr := r.Clone(r.Context())
	var replacedUrlHost string

	if c.urlRewriter != nil {
		if replacedUrl, err := c.urlRewriter.Replace(r.URL); err != nil {
			return nil, err
		} else {
			rr.URL = replacedUrl
			replacedUrlHost = replacedUrl.Host
		}
	}

	getReplacedHost := (func() string {
		// When user intentionally set the Host header, prioritize it.
		// Need to use MIMEHeader because http header is case-insensitive.
		reqHeaders := toMIMEHeader(c.reqHeaders)
		if reqHeaders.Get("Host") != "" {
			return reqHeaders.Get("Host")
		}

		// When the URL is rewritten, the Host should be rewritten to match the rewritten URL. Otherwise, the original Host will remain.
		if replacedUrlHost != "" {
			return replacedUrlHost
		}

		// Use the original Host
		return rr.Host
	})
	rr.Host = getReplacedHost()

	t := utils.GetTransport(c.insecureCipherSuites, c.proxyUrl)
	res, err := t.RoundTrip(rr)
	if err != nil {
		return nil, err
	}

	res.Header.Set(HEADER_RESPONSE_TYPE, "rewrite")
	res.Header.Set(HEADER_REWRITE_TO, rr.URL.String())
	return res, nil
}

func toMIMEHeader(headers map[string]string) textproto.MIMEHeader {
	mimeHeader := textproto.MIMEHeader{}
	for k, v := range headers {
		mimeHeader.Set(k, v)
	}
	return mimeHeader
}
