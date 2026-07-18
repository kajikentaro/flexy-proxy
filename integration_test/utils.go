package test_utils

import (
	"bufio"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/kajikentaro/flexy-proxy/loggers"
	"github.com/kajikentaro/flexy-proxy/proxy"
	"github.com/kajikentaro/flexy-proxy/utils/configs"
	"github.com/stretchr/testify/require"
)

var SSE_DURATION = 1 * time.Second

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
	mux.HandleFunc("/sse", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		for i := 0; i < 5; i++ {
			fmt.Fprintf(w, "data: SSE message %d\n\n", i)
			w.(http.Flusher).Flush()
			time.Sleep(SSE_DURATION)
		}
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

func StartProxyServer(ctx context.Context, proxyAddr string, configPath string) error {
	proxyConfig, err := configs.ParseConfigFile(configPath)
	if err != nil {
		return err
	}

	proxy := proxy.SetupProxy(proxyConfig)

	srv := &http.Server{Addr: proxyAddr, Handler: proxy}

	// wait for ending the previous server
	time.Sleep(100 * time.Millisecond)
	go func() {
		err := StartServer(srv)
		if err != nil {
			proxyConfig.Logger.Error("failed to start a server", err)
		}
	}()
	go func() {
		<-ctx.Done()
		err := StopServer(srv)
		if err != nil {
			proxyConfig.Logger.Error("failed to shutdown the server", err)
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

func TestSSEStreams(t *testing.T, actual io.ReadCloser, isCapital bool) {
	var isCapitalToExpected = map[bool][]string{
		false: {
			"data: SSE message 0\n\n",
			"data: SSE message 0\n\ndata: SSE message 1\n\n",
			"data: SSE message 0\n\ndata: SSE message 1\n\ndata: SSE message 2\n\n",
			"data: SSE message 0\n\ndata: SSE message 1\n\ndata: SSE message 2\n\ndata: SSE message 3\n\n",
			"data: SSE message 0\n\ndata: SSE message 1\n\ndata: SSE message 2\n\ndata: SSE message 3\n\ndata: SSE message 4\n\n",
		},
		true: {
			"DATA: SSE message 0\n\n",
			"DATA: SSE message 0\n\nDATA: SSE message 1\n\n",
			"DATA: SSE message 0\n\nDATA: SSE message 1\n\nDATA: SSE message 2\n\n",
			"DATA: SSE message 0\n\nDATA: SSE message 1\n\nDATA: SSE message 2\n\nDATA: SSE message 3\n\n",
			"DATA: SSE message 0\n\nDATA: SSE message 1\n\nDATA: SSE message 2\n\nDATA: SSE message 3\n\nDATA: SSE message 4\n\n",
		},
	}
	expected := isCapitalToExpected[isCapital]

	output := ""
	go func() {
		scanner := bufio.NewScanner(actual)
		for scanner.Scan() {
			line := scanner.Text()
			output += line + "\n"
		}
	}()

	time.Sleep(SSE_DURATION / 2)
	for i := 0; i < 5; i++ {
		require.Equal(t, expected[i], output)
		time.Sleep(SSE_DURATION)
	}
}
