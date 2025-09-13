package utils

import (
	"crypto/tls"
	"net/http"
	"net/url"
)

func GetTransport(insecureCipherSuites bool, proxyUrl *url.URL) *http.Transport {
	// if cipherSuites is nil, Go uses a default list
	var cipherSuites []uint16

	if insecureCipherSuites {
		for _, cs := range tls.CipherSuites() {
			cipherSuites = append(cipherSuites, cs.ID)
		}
		for _, cs := range tls.InsecureCipherSuites() {
			cipherSuites = append(cipherSuites, cs.ID)
		}
	}

	t := http.DefaultTransport.(*http.Transport).Clone()
	t.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true,
		CipherSuites:       cipherSuites,
	}
	if proxyUrl != nil {
		t.Proxy = func(req *http.Request) (*url.URL, error) {
			return proxyUrl, nil
		}
	} else {
		t.Proxy = http.ProxyFromEnvironment
	}
	return t
}
