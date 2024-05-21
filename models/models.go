package models

type ProxyConfig struct {
	Routes []struct {
		Url     string
		Content string
		File    string
	}
	DefaultRoute struct {
		ProxyUrl string `yaml:"proxy_url"`
	} `yaml:"default_route"`
}
