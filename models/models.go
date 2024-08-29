package models

import (
	"net/http"
	"net/url"

	"github.com/kajikentaro/elastic-proxy/models/rewrite"
)

type RawConfig struct {
	Routes       []Route
	DefaultRoute DefaultRoute `yaml:"default_route"`
	LogLevel     string       `yaml:"log_level"`
	AlwaysMitm   bool         `yaml:"always_mitm"`
}

type DefaultRoute struct {
	ProxyUrl   string `yaml:"proxy"`
	DenyAccess bool   `yaml:"deny_access"`
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
		Rewrite *rewrite.Rewrite
		Content *string
		File    *string

		ContentType string `yaml:"content_type"`
		Status      int
		Headers     map[string]string
		Transform   string
	}
}

type Handler interface {
	http.Handler
	GetResponseInfo() map[string]string
	GetType() string
}
