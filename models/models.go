package models

import (
	"go-proxy/models/replace"
	"net/http"
	"net/url"
)

type RawConfig struct {
	Routes       []Route
	DefaultRoute DefaultRoute `yaml:"default_route"`
	LogLevel     string       `yaml:"log_level"`
	AlwaysMitm   bool         `yaml:"always_mitm"`
}

type DefaultRoute struct {
	ProxyUrl   string `yaml:"proxy_url"`
	DenyAccess bool   `yaml:"deny_access"`
}

type Handler struct {
	Content      ContentHandler
	File         FileHandler
	ReverseProxy ReverseProxyHandler
}

type Router interface {
	GetHttpsHostList() []string
	GetHandler(*url.URL) (handler Handler, matchedUrl string, err error)
	GetUrlList() []string
}

type Route struct {
	Url      string
	Regex    bool
	Response struct {
		Url     *replace.Url
		Content string
		File    string

		ContentType string `yaml:"content_type"`
		Status      int
	}
}

type FileHandler interface {
	Handler(w http.ResponseWriter, r *http.Request)
	FilePath() string
}

type ContentHandler interface {
	Handler(w http.ResponseWriter, r *http.Request)
	Content() string
}

type ReverseProxyHandler interface {
	Handler(w http.ResponseWriter, r *http.Request)
	ForwardUrl() string
}
