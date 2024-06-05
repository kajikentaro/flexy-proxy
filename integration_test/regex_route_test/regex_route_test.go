package test

import (
	"context"
	"fmt"
	test_utils "go-proxy/integration_test"
	"go-proxy/loggers"
	"go-proxy/utils"
	"io"
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var PROXY_PORT_NUMBER = 8085
var PROXY_HTTP_ADDRESS = fmt.Sprintf(":%d", PROXY_PORT_NUMBER)
var PROXY_URL = fmt.Sprintf("http://localhost:%d", PROXY_PORT_NUMBER)

var SAMPLE_SERVER_PORT_NUMBER = 8086
var SAMPLE_SERVER_HTTP_ADDRESS = fmt.Sprintf(":%d", SAMPLE_SERVER_PORT_NUMBER)

func fatalln(a ...any) {
	fmt.Fprintln(os.Stderr, a...)
	os.Exit(1)
}

func TestMain(m *testing.M) {
	config, err := utils.ParseConfig("regex_route_test.yaml")
	if err != nil {
		fatalln("failed to parse config:", err)
	}

	{
		// create a proxy server
		ctx, cancel := context.WithCancel(context.Background())
		err = test_utils.StartProxyServer(ctx, PROXY_HTTP_ADDRESS, config, loggers.GenLogger(nil))
		if err != nil {
			fatalln("failed to start a proxy server:", err)
		}
		defer cancel()
	}
	{
		// create a sample http server to return "hello world"
		ctx, cancel := context.WithCancel(context.Background())
		err = test_utils.StartSampleHttpServer(ctx, SAMPLE_SERVER_HTTP_ADDRESS, loggers.GenLogger(nil))
		if err != nil {
			fatalln("failed to start a http server:", err)
		}
		defer cancel()
	}
	m.Run()
}

func TestContent(t *testing.T) {
	proxyUrl, err := url.Parse(PROXY_URL)
	assert.NoError(t, err)

	res, err := test_utils.Request(proxyUrl, "https://content.test/foo/123")
	assert.NoError(t, err)
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	assert.NoError(t, err)
	assert.Equal(t, "basic", string(body))
}

func TestContent2(t *testing.T) {
	proxyUrl, err := url.Parse(PROXY_URL)
	assert.NoError(t, err)

	res, err := test_utils.Request(proxyUrl, "https://content.test/foo/bar/123")
	assert.NoError(t, err)
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	assert.NoError(t, err)
	assert.Equal(t, "basic", string(body))
}

func TestContentFailure(t *testing.T) {
	proxyUrl, err := url.Parse(PROXY_URL)
	assert.NoError(t, err)

	res, err := test_utils.Request(proxyUrl, "https://content.test/foo/")
	assert.NoError(t, err)
	defer res.Body.Close()

	assert.Equal(t, res.StatusCode, 403)
}

func TestFile(t *testing.T) {
	proxyUrl, err := url.Parse(PROXY_URL)
	assert.NoError(t, err)

	res, err := test_utils.Request(proxyUrl, "https://file.test/foo-bar.txt")
	assert.NoError(t, err)
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	assert.NoError(t, err)
	assert.Equal(t, "sample text", string(body))
}

func TestFileFailure(t *testing.T) {
	proxyUrl, err := url.Parse(PROXY_URL)
	assert.NoError(t, err)

	res, err := test_utils.Request(proxyUrl, "https://file.test/foo.txt2")
	assert.NoError(t, err)
	defer res.Body.Close()

	assert.Equal(t, res.StatusCode, 403)
}

func TestReverseProxy(t *testing.T) {
	proxyUrl, err := url.Parse(PROXY_URL)
	assert.NoError(t, err)

	res, err := test_utils.Request(proxyUrl, "http://localhost:8086/path/v1.2-win64.zip")
	assert.NoError(t, err)
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	assert.NoError(t, err)
	assert.Equal(t, "/path/v1.2.1-win64.zip", string(body))
}
