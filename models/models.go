package models

type ProxyConfig struct {
	Routes       []ProxyRoute
	DefaultRoute struct {
		ProxyUrl string `yaml:"proxy_url"`
	} `yaml:"default_route"`
	LogLevel string `yaml:"log_level"`
}

type ProxyRoute struct {
	Url      string
	Response struct {
		Url     string
		Content string
		File    string

		ContentType string `yaml:"content_type"`
		Status      int
	}
}
