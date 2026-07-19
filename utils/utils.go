package utils

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
	"reflect"
	"time"
)

func GetTransport(insecureCipherSuites bool, proxyUrl *url.URL, connectTo string) *http.Transport {
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

	if connectTo != "" {
		// same settings as http.DefaultTransport
		dialer := &net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}
		t.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialer.DialContext(ctx, network, connectTo)
		}
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

func IsNil(v any) bool {
	if v == nil {
		return true
	}

	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Chan,
		reflect.Func,
		reflect.Interface,
		reflect.Map,
		reflect.Pointer,
		reflect.Slice:
		return rv.IsNil()
	default:
		return false
	}
}
