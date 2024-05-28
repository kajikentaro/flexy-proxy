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

	proxy, err := proxy.SetupProxy(config, loggers.GenLogger())
	if err != nil {
		log.Fatalln("failed to init proxy", err)
	}
	log.Fatal(http.ListenAndServe(":9999", proxy))
}
