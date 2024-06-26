package test

import (
	"context"
	"fmt"
	"io"
	"mime"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	test_utils "github.com/kajikentaro/elastic-proxy/integration_test"
	"github.com/kajikentaro/elastic-proxy/loggers"
	"github.com/kajikentaro/elastic-proxy/utils"

	"github.com/stretchr/testify/assert"
)

var PROXY_PORT_NUMBER = 8081
var PROXY_HTTP_ADDRESS = fmt.Sprintf(":%d", PROXY_PORT_NUMBER)
var PROXY_URL = fmt.Sprintf("http://localhost:%d", PROXY_PORT_NUMBER)

var SAMPLE_SERVER_PORT_NUMBER = 8082
var SAMPLE_SERVER_HTTP_ADDRESS = fmt.Sprintf(":%d", SAMPLE_SERVER_PORT_NUMBER)

func TestRequestOnConfigUrl(t *testing.T) {
	config, err := utils.ParseConfig("route_test.yaml")
	assert.NoError(t, err)
	{
		// create a proxy server
		ctx, cancel := context.WithCancel(context.Background())
		err = test_utils.StartProxyServer(ctx, PROXY_HTTP_ADDRESS, config, loggers.GenLogger(nil))
		assert.NoError(t, err)
		defer cancel()
	}
	{
		// create a sample http server to return "hello world"
		ctx, cancel := context.WithCancel(context.Background())
		err = test_utils.StartSampleHttpServer(ctx, SAMPLE_SERVER_HTTP_ADDRESS, loggers.GenLogger(nil))
		assert.NoError(t, err)
		defer cancel()
	}

	for idx, c := range config.Routes {
		t.Run(fmt.Sprintf("index: %d, route: %s", idx, c.Url), func(t *testing.T) {
			proxyUrl, err := url.Parse(PROXY_URL)
			assert.NoError(t, err)
			res, err := test_utils.Request(proxyUrl, c.Url)
			assert.NoError(t, err)
			defer res.Body.Close()
			body, err := io.ReadAll(res.Body)

			expectedContentType := "text/plain"
			if c.Response.File != "" {
				expectedContentType = mime.TypeByExtension(filepath.Ext(c.Response.File))
			}
			if c.Response.ContentType != "" {
				expectedContentType = c.Response.ContentType
			}
			assert.Equal(t, expectedContentType, res.Header.Get("Content-Type"))

			expectedStatusCode := 200
			if c.Response.Status != 0 {
				expectedStatusCode = c.Response.Status
			}
			assert.Equal(t, expectedStatusCode, res.StatusCode)

			if c.Response.Content != "" {
				assert.NoError(t, err)
				assert.Equal(t, c.Response.Content, string(body))
			} else if c.Response.File != "" {
				b, err := os.ReadFile(c.Response.File)
				assert.NoError(t, err)
				assert.Equal(t, b, body)
			} else if c.Response.Url != nil {
				assert.Equal(t, "hello world", string(body))
			} else {
				t.Error("invalid config format")
			}
		})
	}
}

/*
func TestRequestOnOtherUrl(t *testing.T) {
	config, err := utils.ParseConfig("")
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

*/
