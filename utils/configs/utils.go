package configs

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
	"github.com/kajikentaro/flexy-proxy/utils/gethttps"
	"github.com/xeipuuv/gojsonschema"

	"gopkg.in/yaml.v3"
)

var DEFAULT_CONFIG_PATH = "config.yaml"

func ParseConfig(configPath string) (*proxy.Config, error) {
	rawConfig, err := ReadConfigYaml(configPath)
	if err != nil {
		return nil, err
	}

	proxyConfig, err := parseRawConfig(rawConfig)
	if err != nil {
		return nil, err
	}

	return proxyConfig, nil
}

func getConfigPath(configPath string) (string, error) {
	if configPath == "" {
		return filepath.Abs(DEFAULT_CONFIG_PATH)
	}
	return filepath.Abs(configPath)
}

var DEFAULT_CONFIG = models.RawConfig{
	AlwaysMitm: true,
	LogLevel:   "INFO",
	DefaultRoute: models.RawDefaultRoute{
		DenyAccess: false,
	},
}

func ReadConfigYaml(configPath string) (*models.RawConfig, error) {
	configPath, err := getConfigPath(configPath)
	if err != nil {
		return nil, err
	}

	fileContent, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	{
		var data map[string]interface{}
		err := yaml.Unmarshal(fileContent, &data)
		if err != nil {
			return nil, err
		}

		// validation
		schemaLoader := gojsonschema.NewStringLoader(models.ConfigSpec)
		dataLoader := gojsonschema.NewGoLoader(data)
		result, err := gojsonschema.Validate(schemaLoader, dataLoader)
		if err != nil {
			return nil, err
		}
		if !result.Valid() {
			return nil, convertToError(result.Errors())
		}
	}

	config := DEFAULT_CONFIG
	err = yaml.Unmarshal(fileContent, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func convertToError(errs []gojsonschema.ResultError) error {
	text := "\n"
	for _, err := range errs {
		text += fmt.Sprintf("- %s\n", err.String())
	}
	return errors.New(text)
}

func parseRawConfig(rawConfig *models.RawConfig) (*proxy.Config, error) {
	router, err := routers.NewRouter(rawConfig)
	if err != nil {
		return nil, err
	}

	// parse default proxy
	var defaultProxy *url.URL
	if rawConfig.DefaultRoute.Proxy != "" {
		var err error
		defaultProxy, err = url.Parse(rawConfig.DefaultRoute.Proxy)
		if err != nil {
			return nil, err
		}
	}

	// load certificates
	var cer *tls.Certificate
	if rawConfig.Certificate != "" || rawConfig.CertificateKey != "" {
		if rawConfig.Certificate == "" || rawConfig.CertificateKey == "" {
			return nil, fmt.Errorf("both 'certificate' and 'certificate_key' should be specified in the config file")
		}
		if _, err := os.Stat(rawConfig.Certificate); errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("certificate, '%s', does not exist", rawConfig.Certificate)
		}
		if _, err := os.Stat(rawConfig.CertificateKey); errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("certificate_key, '%s', does not exist", rawConfig.CertificateKey)
		}
		cer_, err := tls.LoadX509KeyPair(rawConfig.Certificate, rawConfig.CertificateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to load certificate: %w", err)
		}
		cer = &cer_
	}

	logLevelStr := "INFO"
	if rawConfig.LogLevel != "" {
		logLevelStr = rawConfig.LogLevel
	}
	logLevel, err := loggers.StrToLogLevel(logLevelStr)
	if err != nil {
		return nil, err
	}
	logger := loggers.GenLogger(&loggers.LoggerSettings{
		LogLevel: logLevel,
	})

	var httpsHostNames []string
	if !rawConfig.AlwaysMitm {
		var err error
		httpsHostNames, err = gethttps.GetHttpsHostList(rawConfig.Routes)
		if err != nil {
			return nil, err
		}
	}

	proxyConfig := &proxy.Config{
		DefaultRoute:         proxy.DefaultRoute{Proxy: defaultProxy, DenyAccess: rawConfig.DefaultRoute.DenyAccess},
		AlwaysMitm:           rawConfig.AlwaysMitm,
		Certificate:          cer,
		Logger:               logger,
		Router:               router,
		HttpsHostNames:       httpsHostNames,
		InsecureCipherSuites: rawConfig.InsecureCipherSuites,
	}

	logger.Info("Successfully parsed the config file",
		"route_length", len(rawConfig.Routes),
		"always_mitm", rawConfig.AlwaysMitm,
		"certificate_loaded", cer != nil,
		"default_route_proxy", rawConfig.DefaultRoute.Proxy,
		"default_route_deny_access", rawConfig.DefaultRoute.DenyAccess,
		"log_level", logLevelStr,
	)
	return proxyConfig, nil
}
