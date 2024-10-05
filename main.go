package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/kajikentaro/flexy-proxy/proxy"
	"github.com/kajikentaro/flexy-proxy/utils"
)

func fatalf(format string, a ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", a...)
	os.Exit(1)
}

// this will be specified like:
// go build -ldflags "-X 'main.version=1.0.0'"
var version string

func getVersion() string {
	if version == "" {
		return "unknown"
	}
	return version
}

func main() {
	var customConfigPath string
	flag.StringVar(&customConfigPath, "f", utils.DEFAULT_CONFIG_PATH, "Path to custom config file")
	var portNum int
	flag.IntVar(&portNum, "p", 8888, "Port number")
	var showVersion bool
	flag.BoolVar(&showVersion, "version", false, "Show version")
	flag.Parse()

	if showVersion {
		fmt.Println(getVersion())
		return
	}

	config, err := utils.ReadConfigYaml(customConfigPath)
	if err != nil {
		fatalf("Error parsing config: %v", err)
	}

	router, logger, proxyConfig, err := utils.ParseConfig(config)
	if err != nil {
		fatalf("%v", err)
	}

	proxy := proxy.SetupProxy(router, logger, proxyConfig)
	addr := fmt.Sprintf(":%d", portNum)
	logger.Info(fmt.Sprintf("Proxy started on %s", addr))
	fatalf("%v", http.ListenAndServe(addr, proxy))
}
