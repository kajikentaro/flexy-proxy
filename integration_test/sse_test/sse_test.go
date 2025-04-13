package test

import (
	"bufio"
	"context"
	"fmt"
	"io"
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

	testSSEStreams(t, res.Body)
}

// test SSE communication with proxy but without rewrite
func TestSSEWithProxyWithoutRewrite(t *testing.T) {
	res, err := test_utils.Request(PROXY_URL, "http://localhost:8092/sse")
	assert.NoError(t, err)
	defer res.Body.Close()

	testSSEStreams(t, res.Body)
}

// test SSE communication with proxy and with rewrite
func TestSSEWithProxy(t *testing.T) {
	res, err := test_utils.Request(PROXY_URL, "http://sample-sse.test/sse")
	assert.NoError(t, err)
	defer res.Body.Close()

	testSSEStreams(t, res.Body)
}

func testSSEStreams(t *testing.T, body io.ReadCloser) {
	expectedOutput := []string{
		"data: SSE message 0\n\n",
		"data: SSE message 0\n\ndata: SSE message 1\n\n",
		"data: SSE message 0\n\ndata: SSE message 1\n\ndata: SSE message 2\n\n",
		"data: SSE message 0\n\ndata: SSE message 1\n\ndata: SSE message 2\n\ndata: SSE message 3\n\n",
		"data: SSE message 0\n\ndata: SSE message 1\n\ndata: SSE message 2\n\ndata: SSE message 3\n\ndata: SSE message 4\n\n",
	}

	output := ""
	go func() {
		scanner := bufio.NewScanner(body)
		for scanner.Scan() {
			line := scanner.Text()
			output += line + "\n"
		}
	}()

	time.Sleep(test_utils.SSE_DURATION / 2)
	for i := 0; i < 5; i++ {
		require.Equal(t, expectedOutput[i], output)
		time.Sleep(test_utils.SSE_DURATION)
	}
}
