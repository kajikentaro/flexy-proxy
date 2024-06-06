package main

import (
	"flag"
	"fmt"
	"go-proxy/loggers"
	"go-proxy/proxy"
	"go-proxy/routers"
	"go-proxy/utils"
	"net/http"
	"os"
)

func fatalf(format string, a ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", a...)
	os.Exit(1)
}

func main() {
	// TODO: add tests
	var customConfigPath string
	flag.StringVar(&customConfigPath, "f", "", "Path to custom config file")
	flag.Parse()

	config, err := utils.ParseConfig(customConfigPath)
	if err != nil {
		fatalf("Error parsing config: %v", err)
	}

	logLevelStr := "INFO"
	if config.LogLevel != "" {
		logLevelStr = config.LogLevel
	}
	logLevel, err := loggers.StrToLogLevel(logLevelStr)
	if err != nil {
		fatalf("%v", err)
	}
	logger := loggers.GenLogger(&loggers.LoggerSettings{
		LogLevel: logLevel,
	})

	router, err := routers.GenRouter(config.Routes)
	if err != nil {
		fatalf("%v", err)
	}

	proxy := proxy.SetupProxy(router, logger, utils.GetProxyConfig(config))
	fatalf("%v", http.ListenAndServe(":9999", proxy))
}
