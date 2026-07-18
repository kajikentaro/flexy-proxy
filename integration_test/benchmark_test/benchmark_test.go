package test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"testing"

	test_utils "github.com/kajikentaro/flexy-proxy/integration_test"
	"github.com/kajikentaro/flexy-proxy/loggers"
)

const (
	proxyHTTPAddress      = ":%d"
	proxyURLFormat        = "http://localhost:%d"
	sampleServerURLFormat = "http://localhost:%d"
)

var (
	proxyAddress        = fmt.Sprintf(proxyHTTPAddress, test_utils.PORT_NUM_BENCHMARK)
	proxyURL            = fmt.Sprintf(proxyURLFormat, test_utils.PORT_NUM_BENCHMARK)
	sampleServerAddress = fmt.Sprintf(proxyHTTPAddress, test_utils.PORT_NUM_BENCHMARK_SERVER)
	sampleServerURL     = fmt.Sprintf(sampleServerURLFormat, test_utils.PORT_NUM_BENCHMARK_SERVER)
)

func TestMain(m *testing.M) {
	{
		// setup proxy server
		ctx, cancel := context.WithCancel(context.Background())
		err := test_utils.StartProxyServer(ctx, proxyAddress, "benchmark.yaml")
		if err != nil {
			fmt.Fprintln(os.Stderr, "failed to start proxy server:", err)
			os.Exit(1)
		}
		defer cancel()
	}

	{
		// setup sample HTTP server
		ctx, cancel := context.WithCancel(context.Background())
		err := test_utils.StartSampleHttpServer(
			ctx,
			sampleServerAddress,
			loggers.GenLogger(nil),
		)
		if err != nil {
			fmt.Fprintln(os.Stderr, "failed to start sample HTTP server:", err)
			os.Exit(1)
		}
		defer cancel()
	}

	m.Run()
}

func newHTTPClient(b *testing.B, proxyAddress string) *http.Client {
	b.Helper()

	transport := http.DefaultTransport.(*http.Transport).Clone()
	if proxyAddress != "" {
		parsedProxyURL, err := url.Parse(proxyAddress)
		if err != nil {
			b.Fatalf("parse proxy URL: %v", err)
		}
		transport.Proxy = http.ProxyURL(parsedProxyURL)
	} else {
		transport.Proxy = nil
	}

	b.Cleanup(transport.CloseIdleConnections)
	return &http.Client{Transport: transport}
}

func benchmarkGET(b *testing.B, client *http.Client, targetURL, expectedBody string) {
	b.Helper()
	expected := []byte(expectedBody)

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		res, err := client.Get(targetURL)
		if err != nil {
			b.Fatalf("GET %s: %v", targetURL, err)
		}

		body, readErr := io.ReadAll(res.Body)
		closeErr := res.Body.Close()
		if readErr != nil {
			b.Fatalf("read response from %s: %v", targetURL, readErr)
		}
		if closeErr != nil {
			b.Fatalf("close response from %s: %v", targetURL, closeErr)
		}
		if res.StatusCode != http.StatusOK {
			b.Fatalf("GET %s returned status %s", targetURL, res.Status)
		}
		if !bytes.Equal(body, expected) {
			b.Fatalf("GET %s returned %q, want %q", targetURL, body, expected)
		}
	}
}

// BenchmarkHTTPGet measures direct HTTP access as the baseline.
func BenchmarkHTTPGet(b *testing.B) {
	benchmarkGET(b, newHTTPClient(b, ""), sampleServerURL, "hello,world")
}

func BenchmarkProxyHTTPGet(b *testing.B) {
	benchmarkGET(b, newHTTPClient(b, proxyURL), sampleServerURL, "hello,world")
}

func BenchmarkProxyContent(b *testing.B) {
	benchmarkGET(
		b,
		newHTTPClient(b, proxyURL),
		"http://example.com/contents",
		"benchmark content response",
	)
}

func BenchmarkProxyRewrite(b *testing.B) {
	benchmarkGET(
		b,
		newHTTPClient(b, proxyURL),
		"http://example.com/rewrite",
		"hello,world",
	)
}

func BenchmarkProxyFile(b *testing.B) {
	benchmarkGET(
		b,
		newHTTPClient(b, proxyURL),
		"http://example.com/file",
		"hello file",
	)
}

func BenchmarkProxyRewriteTransform(b *testing.B) {
	benchmarkGET(
		b,
		newHTTPClient(b, proxyURL),
		"http://example.com/rewrite-transform",
		"11\n",
	)
}
