package utils

import (
	"go-proxy/models"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func getConfigPath(customPath string) (string, error) {
	if customPath == "" {
		// by default, use "config.yaml" in the current directory
		return filepath.Abs("config.yaml")
	}
	return filepath.Abs(customPath)
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

	var config models.RawConfig
	err = yaml.Unmarshal(fileContent, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
