package main

import (
	"go-proxy"
	"log"
	"net/http"
)

func main() {
	p := proxy.GetProxy()
	log.Fatal(http.ListenAndServe(":9999", p))
}
