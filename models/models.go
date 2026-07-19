package models

import (
	_ "embed"
	"net/http"
	"net/url"

	"github.com/kajikentaro/flexy-proxy/models/rewrite"
)

//go:embed config-spec.json
var ConfigSpec string

type RawConfig struct {
	Routes               []RouteConf
	DefaultRoute         RawDefaultRoute `yaml:"default_route"`
	LogLevel             string          `yaml:"log_level"`
	AlwaysMitm           bool            `yaml:"always_mitm"`
	Certificate          string          `yaml:"certificate"`
	CertificateKey       string          `yaml:"certificate_key"`
	InsecureCipherSuites bool            `yaml:"insecure_cipher_suites"`
}

type RawDefaultRoute struct {
	Proxy      string
	DenyAccess bool `yaml:"deny_access"`
}

type Router interface {
	TryRoundTrip(*http.Request) (successInfo map[string]string, res *http.Response, err error)
	GetMatchedRoute(*url.URL) (route RouteConf, err error)
}

type RouteResponse struct {
	Rewrite *rewrite.Rewrite
	Content *string
	File    *string

	ContentType string `yaml:"content_type"`
	Status      int
	Headers     map[string]string
	Transform   string
}

type RouteRequest struct {
	Headers map[string]string
}

type RouteConf struct {
	Url      string
	Regex    bool
	Response RouteResponse
	Request  RouteRequest
}
