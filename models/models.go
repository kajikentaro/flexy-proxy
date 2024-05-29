package models

type ProxyConfig struct {
	Routes []struct {
		Url      string
		Response struct {
			Content string
			File    string

			ContentType string `yaml:"content_type"`
			Status      int
		}
	}
	DefaultRoute struct {
		ProxyUrl string `yaml:"proxy_url"`
	} `yaml:"default_route"`
}
