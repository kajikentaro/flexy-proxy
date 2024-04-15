package models

type ProxyConfig struct {
	Configs []struct {
		Url     string
		Content string
		File    string
	}
}
