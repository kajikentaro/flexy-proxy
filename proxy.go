package proxy

import (
	"fmt"
	"net/http"
	"net/url"

	_ "embed"

	"github.com/elazarl/goproxy"
)

//go:embed sample1.html
var sample string

//go:embed sample2.txt
var stg string

func getResponseString(req *http.Request, ctx *goproxy.ProxyCtx) (bool, string) {
	url := req.URL.String()
	if url == "https://sample.jp:443/" {
		return true, "matched"
	}
	if url == "https://sample.co.jp:443/foo/bar/" {
		return true, stg
	}
	if url == "https://sample.com:443/" {
		return true, sample
	}
	return false, ""
}

func GetProxy() *goproxy.ProxyHttpServer {
	proxy := goproxy.NewProxyHttpServer()
	// proxy.Verbose = true

	proxy.OnRequest(goproxy.ReqHostIs("sample.jp:443")).HandleConnect(goproxy.AlwaysMitm)
	proxy.OnRequest(goproxy.ReqHostIs("sample.co.jp:443")).HandleConnect(goproxy.AlwaysMitm)
	proxy.OnRequest(goproxy.ReqHostIs("sample.com:443")).HandleConnect(goproxy.AlwaysMitm)

	proxy.OnRequest().DoFunc(
		func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			matched, content := getResponseString(req, ctx)
			if matched {
				fmt.Println("### matched", req.URL)
				res := goproxy.NewResponse(req, goproxy.ContentTypeText, http.StatusOK, content)
				res.Header.Set("Content-Type", "text/html;charset=utf-8")
				return req, res
			}
			return req, nil
		},
	)

	proxy.Tr = &http.Transport{Proxy: func(req *http.Request) (*url.URL, error) {
		return url.Parse("http://localhost:8082")
	}}
	proxy.ConnectDial = proxy.NewConnectDialToProxy("http://localhost:8082")

	return proxy
}
