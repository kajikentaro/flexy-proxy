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

	configFileDir := filepath.Dir(configPath)
	proxyConfig, err := parseRawConfig(rawConfig, configFileDir)
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

func loadCertificate(certPath string, certKeyPath string, baseDir string) (*tls.Certificate, error) {
	if certPath == "" && certKeyPath == "" {
		return nil, nil
	}

	if certPath == "" || certKeyPath == "" {
		return nil, fmt.Errorf("both 'certificate' and 'certificate_key' should be specified in the config file")
	}

	if !filepath.IsAbs(certPath) {
		certPath = filepath.Join(baseDir, certPath)
	}
	if !filepath.IsAbs(certKeyPath) {
		certKeyPath = filepath.Join(baseDir, certKeyPath)
	}

	if _, err := os.Stat(certPath); errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("certificate, '%s', does not exist", certPath)
	}
	if _, err := os.Stat(certKeyPath); errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("certificate_key, '%s', does not exist", certKeyPath)
	}
	cert, err := tls.LoadX509KeyPair(certPath, certKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load certificate: %w", err)
	}

	return &cert, nil
}

func parseRawConfig(rawConfig *models.RawConfig, configFileDir string) (*proxy.Config, error) {
	router, err := routers.NewRouter(rawConfig, configFileDir)
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

	cer, err := loadCertificate(rawConfig.Certificate, rawConfig.CertificateKey, configFileDir)
	if err != nil {
		return nil, err
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
		"config_dir", configFileDir,
	)
	return proxyConfig, nil
}
