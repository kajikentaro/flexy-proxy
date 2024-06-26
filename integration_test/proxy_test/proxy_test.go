package test

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"testing"

	test_utils "github.com/kajikentaro/elastic-proxy/integration_test"
	"github.com/kajikentaro/elastic-proxy/loggers"
	"github.com/kajikentaro/elastic-proxy/utils"

	"github.com/stretchr/testify/assert"
)

var PROXY_PORT_NUMBER_1 = 8083
var PROXY_HTTP_ADDRESS_1 = fmt.Sprintf(":%d", PROXY_PORT_NUMBER_1)
var PROXY_URL_1 = fmt.Sprintf("http://localhost:%d", PROXY_PORT_NUMBER_1)

var PROXY_PORT_NUMBER_2 = 8084
var PROXY_HTTP_ADDRESS_2 = fmt.Sprintf(":%d", PROXY_PORT_NUMBER_2)

func TestRequestOnOtherUrl(t *testing.T) {
	// setup 1st proxy
	// if a request url does not match urls on config file, it goes 2nd proxy
	{
		config, err := utils.ParseConfig("1st_proxy.yaml")
		assert.NoError(t, err)

		ctx, cancel := context.WithCancel(context.Background())
		test_utils.StartProxyServer(ctx, PROXY_HTTP_ADDRESS_1, config, loggers.GenLogger(nil))
		defer cancel()
	}

	// setup 2nd proxy
	{
		config, err := utils.ParseConfig("2nd_proxy.yaml")
		assert.NoError(t, err)

		ctx, cancel := context.WithCancel(context.Background())
		test_utils.StartProxyServer(ctx, PROXY_HTTP_ADDRESS_2, config, loggers.GenLogger(nil))
		defer cancel()
	}

	proxyUrl, err := url.Parse(PROXY_URL_1)
	assert.NoError(t, err)
	res, err := test_utils.Request(proxyUrl, "https://default-proxy.jp/")
	assert.NoError(t, err)
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	assert.NoError(t, err)

	assert.Equal(t, "1,2,3", string(body))
}

func TestRequestDenial(t *testing.T) {
	// setup 1st proxy
	// if a request url does not match urls on config file, it goes 2nd proxy
	{
		config, err := utils.ParseConfig("1st_proxy_deny.yaml")
		assert.NoError(t, err)

		ctx, cancel := context.WithCancel(context.Background())
		test_utils.StartProxyServer(ctx, PROXY_HTTP_ADDRESS_1, config, loggers.GenLogger(nil))
		defer cancel()
	}

	// setup 2nd proxy
	{
		config, err := utils.ParseConfig("2nd_proxy.yaml")
		assert.NoError(t, err)

		ctx, cancel := context.WithCancel(context.Background())
		test_utils.StartProxyServer(ctx, PROXY_HTTP_ADDRESS_2, config, loggers.GenLogger(nil))
		defer cancel()
	}

	// test:
	// if a request url is https, the proxy returns ERR_EMPTY_RESPONSE
	{
		proxyUrl, err := url.Parse(PROXY_URL_1)
		assert.NoError(t, err)
		res, err := test_utils.Request(proxyUrl, "https://out-of-route-url.test")

		var urlError *url.Error
		assert.ErrorAs(t, err, &urlError)
		assert.Nil(t, res)
	}

	// test:
	// if a request url is http, the proxy returns ERR_EMPTY_RESPONSE
	{
		proxyUrl, err := url.Parse(PROXY_URL_1)
		assert.NoError(t, err)
		res, err := test_utils.Request(proxyUrl, "http://out-of-route-url.test")

		assert.NoError(t, err)
		assert.Equal(t, 403, res.StatusCode)

		body, err := io.ReadAll(res.Body)
		assert.NoError(t, err)
		assert.Equal(t, "http://out-of-route-url.test/ is out of routes", string(body))
	}
}

func TestProxyUrlOnRoutes(t *testing.T) {
	// setup 1st proxy
	// if a request url match, it goes 2nd proxy
	{
		config, err := utils.ParseConfig("1st_proxy_on_routes.yaml")
		assert.NoError(t, err)

		ctx, cancel := context.WithCancel(context.Background())
		test_utils.StartProxyServer(ctx, PROXY_HTTP_ADDRESS_1, config, loggers.GenLogger(nil))
		defer cancel()
	}

	// setup 2nd proxy
	{
		config, err := utils.ParseConfig("2nd_proxy.yaml")
		assert.NoError(t, err)

		ctx, cancel := context.WithCancel(context.Background())
		test_utils.StartProxyServer(ctx, PROXY_HTTP_ADDRESS_2, config, loggers.GenLogger(nil))
		defer cancel()
	}

	{
		proxyUrl, err := url.Parse(PROXY_URL_1)
		assert.NoError(t, err)
		res, err := test_utils.Request(proxyUrl, "https://default-proxy.jp/")
		assert.NoError(t, err)
		defer res.Body.Close()
		body, err := io.ReadAll(res.Body)
		assert.NoError(t, err)

		assert.Equal(t, "1,2,3", string(body))
	}
	{
		proxyUrl, err := url.Parse(PROXY_URL_1)
		assert.NoError(t, err)
		res, err := test_utils.Request(proxyUrl, "https://go-proxy.test/")
		assert.NoError(t, err)
		defer res.Body.Close()
		body, err := io.ReadAll(res.Body)
		assert.NoError(t, err)

		assert.Equal(t, "1,2,3", string(body))
	}
}
