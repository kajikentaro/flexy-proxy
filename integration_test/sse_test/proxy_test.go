package test

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	test_utils "github.com/kajikentaro/flexy-proxy/integration_test"
	"github.com/kajikentaro/flexy-proxy/loggers"

	"github.com/stretchr/testify/assert"
)

var PROXY_PORT_NUMBER = 8091
var PROXY_HTTP_ADDRESS = fmt.Sprintf(":%d", PROXY_PORT_NUMBER)
var PROXY_URL, _ = url.Parse(fmt.Sprintf("http://localhost:%d", PROXY_PORT_NUMBER))

var SAMPLE_SERVER_PORT_NUMBER = 8092
var SAMPLE_SERVER_HTTP_ADDRESS = fmt.Sprintf(":%d", SAMPLE_SERVER_PORT_NUMBER)
var SAMPLE_SERVER_URL = fmt.Sprintf("http://localhost:%d", SAMPLE_SERVER_PORT_NUMBER)

var SSE_TIMEOUT_DURATION = test_utils.SSE_DURATION + 500*time.Millisecond

func TestMain(m *testing.M) {
	// setup sample HTTP server
	{
		ctx, cancel := context.WithCancel(context.Background())
		err := test_utils.StartSampleHttpServer(ctx, SAMPLE_SERVER_HTTP_ADDRESS, loggers.GenLogger(nil))
		if err != nil {
			fmt.Fprintln(os.Stderr, "failed to start a sample http server:", err)
			os.Exit(1)
		}
		defer cancel()
	}

	// setup proxy server
	{
		ctx, cancel := context.WithCancel(context.Background())
		test_utils.StartProxyServer(ctx, PROXY_HTTP_ADDRESS, "sse_test.yaml")
		defer cancel()
	}
	m.Run()
}

func TestSSEWithoutProxy(t *testing.T) {
	start := time.Now()
	res, err := http.Get(SAMPLE_SERVER_URL + "/sse")
	assert.NoError(t, err)
	defer res.Body.Close()

	scanner := bufio.NewScanner(res.Body)
	for scanner.Scan() {
		line := scanner.Text()

		duration := time.Since(start)
		assert.Less(t, duration, SSE_TIMEOUT_DURATION)
		assert.True(t, strings.Contains(line, "data: SSE message") || line == "", "unexpected line: %s", line)

		start = time.Now()
	}
	assert.NoError(t, scanner.Err())
}

// test SSE communication with proxy but without rewrite
func TestSSEWithProxyWithoutRewrite(t *testing.T) {

	start := time.Now()
	res, err := test_utils.Request(PROXY_URL, "http://localhost:8092/sse")
	assert.NoError(t, err)
	defer res.Body.Close()

	scanner := bufio.NewScanner(res.Body)
	for scanner.Scan() {
		line := scanner.Text()

		duration := time.Since(start)
		assert.Less(t, duration, SSE_TIMEOUT_DURATION)
		assert.True(t, strings.Contains(line, "data: SSE message") || line == "", "unexpected line: %s", line)

		start = time.Now()
	}
	assert.NoError(t, scanner.Err())

}

// test SSE communication with proxy and with rewrite
func TestSSEWithProxy(t *testing.T) {
	{
		start := time.Now()
		res, err := test_utils.Request(PROXY_URL, "http://sample-sse.test/sse")
		assert.NoError(t, err)
		defer res.Body.Close()

		scanner := bufio.NewScanner(res.Body)
		for scanner.Scan() {
			line := scanner.Text()

			duration := time.Since(start)
			assert.Less(t, duration, SSE_TIMEOUT_DURATION)
			assert.True(t, strings.Contains(line, "data: SSE message") || line == "", "unexpected line: %s", line)

			start = time.Now()
		}
		assert.NoError(t, scanner.Err())
	}
}
