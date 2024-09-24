package test

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"testing"

	test_utils "github.com/kajikentaro/flexy-proxy/integration_test"
	"github.com/kajikentaro/flexy-proxy/loggers"
	"github.com/kajikentaro/flexy-proxy/utils"

	"github.com/stretchr/testify/assert"
)

var PROXY_PORT_NUMBER_1 = 8083
var PROXY_HTTP_ADDRESS_1 = fmt.Sprintf(":%d", PROXY_PORT_NUMBER_1)
var PROXY_URL_1 = fmt.Sprintf("http://localhost:%d", PROXY_PORT_NUMBER_1)

var PROXY_PORT_NUMBER_2 = 8084
var PROXY_HTTP_ADDRESS_2 = fmt.Sprintf(":%d", PROXY_PORT_NUMBER_2)

var SAMPLE_SERVER_PORT_NUMBER = 8089
var SAMPLE_SERVER_HTTP_ADDRESS = fmt.Sprintf(":%d", SAMPLE_SERVER_PORT_NUMBER)

func TestDefaultRoute(t *testing.T) {
	// setup 1st proxy
	// if a request url does not match urls on config file, it goes 2nd proxy
	{
		config, err := utils.ReadConfigYaml("1st_proxy_default.yaml")
		assert.NoError(t, err)

		ctx, cancel := context.WithCancel(context.Background())
		test_utils.StartProxyServer(ctx, PROXY_HTTP_ADDRESS_1, config)
		defer cancel()
	}

	// setup 2nd proxy
	{
		config, err := utils.ReadConfigYaml("2nd_proxy.yaml")
		assert.NoError(t, err)

		ctx, cancel := context.WithCancel(context.Background())
		test_utils.StartProxyServer(ctx, PROXY_HTTP_ADDRESS_2, config)
		defer cancel()
	}

	{
		// should use default proxy if the request is out of routes
		proxyUrl, err := url.Parse(PROXY_URL_1)
		assert.NoError(t, err)
		res, err := test_utils.Request(proxyUrl, "https://out-of-route.test/")
		assert.NoError(t, err)
		defer res.Body.Close()
		body, err := io.ReadAll(res.Body)
		assert.NoError(t, err)

		assert.Equal(t, "1,2,3", string(body))
	}

	{
		// should use default proxy if no additional proxy is specified
		proxyUrl, err := url.Parse(PROXY_URL_1)
		assert.NoError(t, err)
		res, err := test_utils.Request(proxyUrl, "https://on-route.test")
		assert.NoError(t, err)
		defer res.Body.Close()
		body, err := io.ReadAll(res.Body)
		assert.NoError(t, err)

		assert.Equal(t, "1,2,3", string(body))
	}
	{
		// should use default proxy if only use transform
		proxyUrl, err := url.Parse(PROXY_URL_1)
		assert.NoError(t, err)
		res, err := test_utils.Request(proxyUrl, "https://only-transform.test")
		assert.NoError(t, err)
		defer res.Body.Close()
		body, err := io.ReadAll(res.Body)
		assert.NoError(t, err)

		assert.Equal(t, "replaced", string(body))
	}
}

func TestRequestDenial(t *testing.T) {
	// setup 1st proxy
	// if a request url does not match urls on config file, it goes 2nd proxy
	{
		config, err := utils.ReadConfigYaml("1st_proxy_deny.yaml")
		assert.NoError(t, err)

		ctx, cancel := context.WithCancel(context.Background())
		test_utils.StartProxyServer(ctx, PROXY_HTTP_ADDRESS_1, config)
		defer cancel()
	}

	// setup 2nd proxy
	{
		config, err := utils.ReadConfigYaml("2nd_proxy.yaml")
		assert.NoError(t, err)

		ctx, cancel := context.WithCancel(context.Background())
		test_utils.StartProxyServer(ctx, PROXY_HTTP_ADDRESS_2, config)
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

func TestProxyOnEachRoutes(t *testing.T) {
	// setup 1st proxy
	// if a request url match, it goes 2nd proxy
	{
		config, err := utils.ReadConfigYaml("1st_proxy_on_routes.yaml")
		assert.NoError(t, err)

		ctx, cancel := context.WithCancel(context.Background())
		test_utils.StartProxyServer(ctx, PROXY_HTTP_ADDRESS_1, config)
		defer cancel()
	}

	// setup 2nd proxy
	{
		config, err := utils.ReadConfigYaml("2nd_proxy.yaml")
		assert.NoError(t, err)

		ctx, cancel := context.WithCancel(context.Background())
		test_utils.StartProxyServer(ctx, PROXY_HTTP_ADDRESS_2, config)
		defer cancel()
	}

	{
		proxyUrl, err := url.Parse(PROXY_URL_1)
		assert.NoError(t, err)
		res, err := test_utils.Request(proxyUrl, "https://2nd-proxy.test/")
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

func TestOverwriteProxy(t *testing.T) {
	// setup 1st proxy
	{
		config, err := utils.ReadConfigYaml("1st_proxy_overwrite_default.yaml")
		assert.NoError(t, err)

		ctx, cancel := context.WithCancel(context.Background())
		test_utils.StartProxyServer(ctx, PROXY_HTTP_ADDRESS_1, config)
		defer cancel()
	}

	// setup 2nd proxy
	{
		config, err := utils.ReadConfigYaml("2nd_proxy.yaml")
		assert.NoError(t, err)

		ctx, cancel := context.WithCancel(context.Background())
		test_utils.StartProxyServer(ctx, PROXY_HTTP_ADDRESS_2, config)
		defer cancel()
	}

	{
		// sample http server
		ctx, cancel := context.WithCancel(context.Background())
		err := test_utils.StartSampleHttpServer(ctx, SAMPLE_SERVER_HTTP_ADDRESS, loggers.GenLogger(nil))
		assert.NoError(t, err)
		defer cancel()
	}

	{
		// should overwrite the default proxy with another proxy
		proxyUrl, err := url.Parse(PROXY_URL_1)
		assert.NoError(t, err)
		res, err := test_utils.Request(proxyUrl, "https://overwrite-proxy.test/")
		assert.NoError(t, err)
		defer res.Body.Close()
		body, err := io.ReadAll(res.Body)
		assert.NoError(t, err)

		assert.Equal(t, "1,2,3", string(body))
	}

	{
		// should not use default proxy and access the internet directly
		proxyUrl, err := url.Parse(PROXY_URL_1)
		assert.NoError(t, err)
		res, err := test_utils.Request(proxyUrl, "https://remove-proxy.test/")
		assert.NoError(t, err)
		defer res.Body.Close()
		body, err := io.ReadAll(res.Body)
		assert.NoError(t, err)

		assert.Equal(t, "hello,world", string(body))
	}
}
