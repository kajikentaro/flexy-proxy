package models

import (
	"net/http"
	"net/url"

	"github.com/kajikentaro/elastic-proxy/models/replace"
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
		Content *string
		File    *string

		ContentType string `yaml:"content_type"`
		Status      int
		Headers     map[string]string
	}
}

type Handler interface {
	http.Handler
	GetResponseInfo() map[string]string
	GetType() string
}
