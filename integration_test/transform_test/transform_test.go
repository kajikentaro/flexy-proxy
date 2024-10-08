package test

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"testing"

	test_utils "github.com/kajikentaro/flexy-proxy/integration_test"
	"github.com/kajikentaro/flexy-proxy/loggers"
	"github.com/kajikentaro/flexy-proxy/utils"

	"github.com/stretchr/testify/assert"
)

var PROXY_PORT_NUMBER = 8087
var PROXY_HTTP_ADDRESS = fmt.Sprintf(":%d", PROXY_PORT_NUMBER)
var PROXY_URL = fmt.Sprintf("http://localhost:%d", PROXY_PORT_NUMBER)

var SAMPLE_SERVER_PORT_NUMBER = 8088
var SAMPLE_SERVER_HTTP_ADDRESS = fmt.Sprintf(":%d", SAMPLE_SERVER_PORT_NUMBER)

func fatalln(a ...any) {
	fmt.Fprintln(os.Stderr, a...)
	os.Exit(1)
}

func TestMain(m *testing.M) {
	config, err := utils.ReadConfigYaml("transform_test.yaml")
	if err != nil {
		fatalln("failed to parse config:", err)
	}

	{
		// create a proxy server
		ctx, cancel := context.WithCancel(context.Background())
		err = test_utils.StartProxyServer(ctx, PROXY_HTTP_ADDRESS, config)
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

	res, err := test_utils.Request(proxyUrl, "https://content.test/")
	assert.NoError(t, err)
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	assert.NoError(t, err)
	assert.Equal(t, "bar", string(body))
}

func TestFile(t *testing.T) {
	proxyUrl, err := url.Parse(PROXY_URL)
	assert.NoError(t, err)

	res, err := test_utils.Request(proxyUrl, "https://file.test/")
	assert.NoError(t, err)
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	assert.NoError(t, err)
	assert.Equal(t, "baz sample text", string(body))
}

func TestRewrite(t *testing.T) {
	proxyUrl, err := url.Parse(PROXY_URL)
	assert.NoError(t, err)

	res, err := test_utils.Request(proxyUrl, "https://url.test/")
	assert.NoError(t, err)
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	assert.NoError(t, err)
	assert.Equal(t, "11\n", string(body))
}
