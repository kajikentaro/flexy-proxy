package test

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"

	test_utils "github.com/kajikentaro/flexy-proxy/integration_test"
	"github.com/kajikentaro/flexy-proxy/loggers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var PROXY_PORT_NUMBER = test_utils.PORT_NUM_SSE
var PROXY_HTTP_ADDRESS = fmt.Sprintf(":%d", PROXY_PORT_NUMBER)
var PROXY_URL, _ = url.Parse(fmt.Sprintf("http://localhost:%d", PROXY_PORT_NUMBER))

var SAMPLE_SERVER_PORT_NUMBER = test_utils.PORT_NUM_SSE_SERVER
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
		err := test_utils.StartProxyServer(ctx, PROXY_HTTP_ADDRESS, "sse_test.yaml")
		if err != nil {
			fmt.Fprintln(os.Stderr, "failed to start a proxy server:", err)
			os.Exit(1)
		}
		defer cancel()
	}
	m.Run()
}

func TestSSEWithoutProxy(t *testing.T) {
	res, err := http.Get(SAMPLE_SERVER_URL + "/sse")
	assert.NoError(t, err)
	defer res.Body.Close()

	test_utils.TestSSEStreams(t, res.Body, false)
}

// test SSE communication with proxy but without rewrite
func TestSSEWithProxyWithoutRewrite(t *testing.T) {
	res, err := test_utils.Request(PROXY_URL, fmt.Sprintf("http://localhost:%d/sse", test_utils.PORT_NUM_SSE_SERVER))
	assert.NoError(t, err)
	defer res.Body.Close()

	test_utils.TestSSEStreams(t, res.Body, false)
}

// test SSE communication with proxy and with rewrite
func TestSSEWithProxy(t *testing.T) {
	res, err := test_utils.Request(PROXY_URL, "http://sample-sse.test/sse")
	assert.NoError(t, err)
	defer res.Body.Close()

	test_utils.TestSSEStreams(t, res.Body, false)
}

// test SSE communication with proxy and with rewrite and transform commands
func TestSSEWithTransform(t *testing.T) {
	res, err := test_utils.Request(PROXY_URL, "http://sample-sse.test/sse-with-transform")
	require.NoError(t, err)
	defer res.Body.Close()

	test_utils.TestSSEStreams(t, res.Body, true)
}
