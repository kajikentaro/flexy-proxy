package utils

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"
)

func StartServer(srv *http.Server) {
	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		fmt.Fprintf(os.Stderr, "failed to start server: %s", err)
		os.Exit(1)
	}
}

func StopServer(srv *http.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := srv.Shutdown(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to shutdown server")
		os.Exit(1)
	}
}

func Request(proxyUrl *url.URL, targetUrl string) (*http.Response, error) {
	client := &http.Client{
		Transport: &http.Transport{
			Proxy:           http.ProxyURL(proxyUrl),
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	fmt.Println(targetUrl)
	return client.Get(targetUrl)
}
