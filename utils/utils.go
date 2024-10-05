package utils

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/kajikentaro/flexy-proxy/loggers"
	"github.com/kajikentaro/flexy-proxy/models"
	"github.com/kajikentaro/flexy-proxy/proxy"
	"github.com/kajikentaro/flexy-proxy/routers"

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
	DefaultRoute: models.RawDefaultRoute{
		DenyAccess: false,
	},
}

func ReadConfigYaml(customPath string) (*models.RawConfig, error) {
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

func ParseConfig(rawConfig *models.RawConfig) (models.Router, *loggers.Logger, *proxy.Config, error) {
	var defaultProxy *url.URL
	if rawConfig.DefaultRoute.Proxy != "" {
		var err error
		defaultProxy, err = url.Parse(rawConfig.DefaultRoute.Proxy)
		if err != nil {
			return nil, nil, nil, err
		}
	}

	router, err := routers.GenRouter(rawConfig.Routes, defaultProxy)
	if err != nil {
		return nil, nil, nil, err
	}

	// load certificates
	var cer *tls.Certificate
	if rawConfig.Certificate != "" || rawConfig.CertificateKey != "" {
		if rawConfig.Certificate == "" || rawConfig.CertificateKey == "" {
			return nil, nil, nil, fmt.Errorf("both 'certificate' and 'certificate_key' should be specified in the config file")
		}
		if _, err := os.Stat(rawConfig.Certificate); errors.Is(err, os.ErrNotExist) {
			return nil, nil, nil, fmt.Errorf("certificate, '%s', does not exist", rawConfig.Certificate)
		}
		if _, err := os.Stat(rawConfig.CertificateKey); errors.Is(err, os.ErrNotExist) {
			return nil, nil, nil, fmt.Errorf("certificate_key, '%s', does not exist", rawConfig.CertificateKey)
		}
		cer_, err := tls.LoadX509KeyPair(rawConfig.Certificate, rawConfig.CertificateKey)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to load certificate: %w", err)
		}
		cer = &cer_
	}

	proxyConfig := &proxy.Config{
		DefaultRoute: proxy.DefaultRoute{
			Proxy:      defaultProxy,
			DenyAccess: rawConfig.DefaultRoute.DenyAccess,
		},
		AlwaysMitm:  rawConfig.AlwaysMitm,
		Certificate: cer,
	}

	logLevelStr := "INFO"
	if rawConfig.LogLevel != "" {
		logLevelStr = rawConfig.LogLevel
	}
	logLevel, err := loggers.StrToLogLevel(logLevelStr)
	if err != nil {
		return nil, nil, nil, err
	}
	logger := loggers.GenLogger(&loggers.LoggerSettings{
		LogLevel: logLevel,
	})

	return router, logger, proxyConfig, nil
}
