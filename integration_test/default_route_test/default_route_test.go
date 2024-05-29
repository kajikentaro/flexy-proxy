package integration_test

import (
	"context"
	"fmt"
	test_utils "go-proxy/integration_test"
	"go-proxy/loggers"
	default_proxy "go-proxy/to_be_remove"
	"go-proxy/utils"
	"io"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var PROXY_PORT_NUMBER = 8083
var PROXY_HTTP_ADDRESS = fmt.Sprintf(":%d", PROXY_PORT_NUMBER)
var PROXY_URL = fmt.Sprintf("http://localhost:%d", PROXY_PORT_NUMBER)

func TestRequestOnOtherUrl(t *testing.T) {
	config, err := utils.ParseConfig("default_route_test.yaml")
	assert.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	test_utils.StartProxyServer(ctx, PROXY_HTTP_ADDRESS, config, loggers.GenLogger())
	defer cancel()

	// 2nd proxy which is used if a request url does not match urls on config file
	p := default_proxy.GetProxy()
	srv := &http.Server{Addr: ":8082", Handler: p}
	go test_utils.StartServer(srv)
	// wait for starting the server
	time.Sleep(time.Second)
	defer test_utils.StopServer(srv)

	proxyUrl, err := url.Parse(PROXY_URL)
	assert.NoError(t, err)
	res, err := test_utils.Request(proxyUrl, "https://default-proxy.jp/")
	assert.NoError(t, err)
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	assert.NoError(t, err)

	assert.Equal(t, "1,2,3", string(body))
}
