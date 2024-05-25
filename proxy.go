package proxy

import (
	"fmt"
	"go-proxy/models"
	"log"
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

func GetProxy(config *models.ProxyConfig) *goproxy.ProxyHttpServer {
	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = true

	for _, route := range config.Routes {
		routeUrl, err := url.Parse(route.Url)
		if err != nil {
			log.Fatal(err)
		}
		if routeUrl.Scheme == "https" {
			reqHost := fmt.Sprintf("%s:443", routeUrl.Host)
			proxy.OnRequest(goproxy.ReqHostIs(reqHost)).HandleConnect(goproxy.AlwaysMitm)
			continue
		}
		if routeUrl.Scheme == "http" {
			// TODO
			continue
		}
		log.Fatalf("scheme of '%s' must be either 'http' or 'https'", route.Url)
	}

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
