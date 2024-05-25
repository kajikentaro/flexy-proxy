package main

import (
	"flag"
	"go-proxy"
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

	p := proxy.GetProxy(config)
	log.Fatal(http.ListenAndServe(":9999", p))
}
