package default_proxy

import (
	"fmt"
	"net/http"

	_ "embed"

	"github.com/elazarl/goproxy"
)

//go:embed sample3.csv
var sample string

func getResponseString(req *http.Request, ctx *goproxy.ProxyCtx) (bool, string) {
	url := req.URL.String()
	if url == "https://default-proxy.jp:443/" {
		return true, sample
	}
	return false, ""
}

func GetProxy() *goproxy.ProxyHttpServer {
	proxy := goproxy.NewProxyHttpServer()
	// proxy.Verbose = true

	proxy.OnRequest(goproxy.ReqHostIs("default-proxy.jp:443")).HandleConnect(goproxy.AlwaysMitm)

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

	return proxy
}
