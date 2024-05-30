package integration_test

import (
	"context"
	"fmt"
	test_utils "go-proxy/integration_test"
	"go-proxy/loggers"
	"go-proxy/utils"
	"io"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

var PROXY_PORT_NUMBER_1 = 8083
var PROXY_HTTP_ADDRESS_1 = fmt.Sprintf(":%d", PROXY_PORT_NUMBER_1)
var PROXY_URL_1 = fmt.Sprintf("http://localhost:%d", PROXY_PORT_NUMBER_1)

var PROXY_PORT_NUMBER_2 = 8084
var PROXY_HTTP_ADDRESS_2 = fmt.Sprintf(":%d", PROXY_PORT_NUMBER_2)

func TestRequestOnOtherUrl(t *testing.T) {
	// setup 1st proxy
	{
		config, err := utils.ParseConfig("1st_proxy.yaml")
		assert.NoError(t, err)

		ctx, cancel := context.WithCancel(context.Background())
		test_utils.StartProxyServer(ctx, PROXY_HTTP_ADDRESS_1, config, loggers.GenLogger(nil))
		defer cancel()
	}

	// setup 2nd proxy which is used if a request url does not match urls on config file
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
