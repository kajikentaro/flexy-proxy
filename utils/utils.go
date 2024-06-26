package utils

import (
	"os"
	"path/filepath"

	"github.com/kajikentaro/elastic-proxy/models"
	"github.com/kajikentaro/elastic-proxy/proxy"

	"gopkg.in/yaml.v3"
)

var DEFAULT_CONFIG_PATH = "config.yaml"

func getConfigPath(customPath string) (string, error) {
	if customPath == "" {
		return filepath.Abs(DEFAULT_CONFIG_PATH)
	}
	return filepath.Abs(customPath)
}

var DEFAULT_CONFIG = models.RawConfig{
	AlwaysMitm: true,
	LogLevel:   "INFO",
	DefaultRoute: models.DefaultRoute{
		DenyAccess: false,
	},
}

func ParseConfig(customPath string) (*models.RawConfig, error) {
	configPath, err := getConfigPath(customPath)
	if err != nil {
		return nil, err
	}

	fileContent, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	config := DEFAULT_CONFIG
	err = yaml.Unmarshal(fileContent, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func GetProxyConfig(rawConfig *models.RawConfig) *proxy.Config {
	return &proxy.Config{
		DefaultRoute: &rawConfig.DefaultRoute,
		AlwaysMitm:   rawConfig.AlwaysMitm,
	}
}
