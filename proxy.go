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

func isUrlSame(a *url.URL, b *url.URL) bool {
	if a.Scheme != b.Scheme {
		return false
	}
	if a.Hostname() != b.Hostname() {
		return false
	}
	if a.RawPath != b.RawPath {
		return false
	}
	if a.RawQuery != b.RawQuery {
		return false
	}
	if a.RawFragment != b.RawFragment {
		return false
	}
	return true
}

func handleProxyRuntimeError(req *http.Request, title string, message string) (*http.Request, *http.Response) {
	logContent := fmt.Sprintf("proxy runtime error: %s: %s", title, message)
	log.Println(logContent)
	res := goproxy.NewResponse(
		req,
		goproxy.ContentTypeText,
		http.StatusInternalServerError,
		logContent,
	)
	return req, res
}

func serveContent(req *http.Request, userStatusCode int, userContentType string, content string) (*http.Request, *http.Response) {
	contentType := goproxy.ContentTypeText
	if userContentType != "" {
		contentType = userContentType
	}
	statusCode := 200
	if userStatusCode != 0 {
		statusCode = userStatusCode
	}

	res := goproxy.NewResponse(req, contentType, statusCode, content)
	return req, res
}

func serveFile(req *http.Request, userStatusCode int, userContentType string, fileName string) (*http.Request, *http.Response) {
	fileRes := NewFileResponse(req)
	res := fileRes.res
	http.ServeFile(fileRes, req, fileName)
	if userContentType != "" {
		// if contentType is specified, overwrite it
		res.Header["Content-Type"][0] = userContentType
	}
	statusCode := 200
	if userStatusCode != 0 {
		statusCode = userStatusCode
	}
	res.StatusCode = statusCode
	return req, res
}

func GetProxy(config *models.ProxyConfig) *goproxy.ProxyHttpServer {
	// TODO: validate config file
	// i.e.: do not contain both content and file
	proxy := goproxy.NewProxyHttpServer()
	// proxy.Verbose = true

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
			for _, route := range config.Routes {
				routeUrl, _ := url.Parse(route.Url)
				if !isUrlSame(routeUrl, req.URL) {
					// unmatched
					continue
				}
				// matched

				if route.File != "" {
					// TODO use std out
					log.Printf("'%s' was routed to the file: '%s'", req.URL.String(), route.File)
					return serveFile(req, route.Status, route.ContentType, route.File)
				}

				if route.Content != "" {
					log.Printf("'%s' was routed to the content: '%s'", req.URL.String(), route.Content)
					return serveContent(req, route.Status, route.ContentType, route.Content)
				}

				return handleProxyRuntimeError(req, "File or Content are not specified", "")
			}
			return req, nil
		},
	)

	if config.DefaultRoute.ProxyUrl != "" {
		proxy.ConnectDial = proxy.NewConnectDialToProxy(config.DefaultRoute.ProxyUrl)
	}

	return proxy
}
