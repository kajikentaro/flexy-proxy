package test_utils

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/kajikentaro/elastic-proxy/loggers"
	"github.com/kajikentaro/elastic-proxy/models"
	"github.com/kajikentaro/elastic-proxy/proxy"
	"github.com/kajikentaro/elastic-proxy/routers"
	"github.com/kajikentaro/elastic-proxy/utils"
)

func StartSampleHttpServer(ctx context.Context, addr string, logger *loggers.Logger) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/path/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		fmt.Fprintf(w, r.URL.Path)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/csv")
		fmt.Fprintf(w, "hello,world")
	})

	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	// wait for ending the previous server
	time.Sleep(100 * time.Millisecond)
	go func() {
		err := StartServer(srv)
		if err != nil {
			logger.Error("failed to start a server", err)
		}
	}()
	go func() {
		<-ctx.Done()
		err := StopServer(srv)
		if err != nil {
			logger.Error("failed to shutdown the server", err)
		}
	}()

	// wait for starting the server
	time.Sleep(100 * time.Millisecond)

	return nil
}

func StartProxyServer(ctx context.Context, proxyAddr string, config *models.RawConfig, logger *loggers.Logger) error {
	router, err := routers.GenRouter(config.Routes)
	if err != nil {
		return err
	}

	proxy := proxy.SetupProxy(router, logger, utils.GetProxyConfig(config))

	srv := &http.Server{Addr: proxyAddr, Handler: proxy}

	// wait for ending the previous server
	time.Sleep(100 * time.Millisecond)
	go func() {
		err := StartServer(srv)
		if err != nil {
			logger.Error("failed to start a server", err)
		}
	}()
	go func() {
		<-ctx.Done()
		err := StopServer(srv)
		if err != nil {
			logger.Error("failed to shutdown the server", err)
		}
	}()

	// wait for starting the server
	time.Sleep(100 * time.Millisecond)

	return nil
}

func StartServer(srv *http.Server) error {
	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func StopServer(srv *http.Server) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := srv.Shutdown(ctx)
	if err != nil {
		return err
	}
	return nil
}

func Request(proxyUrl *url.URL, targetUrl string) (*http.Response, error) {
	client := &http.Client{
		Transport: &http.Transport{
			Proxy:           http.ProxyURL(proxyUrl),
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	return client.Get(targetUrl)
}
