package main

import (
	"flag"
	"go-proxy"
	"go-proxy/loggers"
	"go-proxy/utils"
	"log"
	"net/http"
)

func main() {
	// TODO: add tests
	var customConfigPath string
	flag.StringVar(&customConfigPath, "f", "", "Path to custom config file")
	flag.Parse()

	config, err := utils.ParseConfig(customConfigPath)
	if err != nil {
		log.Fatalf("Error parsing config: %v", err)
	}

	logLevelStr := "INFO"
	if config.LogLevel != "" {
		logLevelStr = config.LogLevel
	}
	logLevel, err := loggers.StrToLogLevel(logLevelStr)
	if err != nil {
		log.Fatal(err)
	}
	logger := loggers.GenLogger(&loggers.LoggerSettings{
		LogLevel: logLevel,
	})

	proxy, err := proxy.SetupProxy(config, logger)
	if err != nil {
		log.Fatalln("failed to init proxy", err)
	}
	log.Fatal(http.ListenAndServe(":9999", proxy))
}
