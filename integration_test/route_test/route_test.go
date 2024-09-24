package test

import (
	"context"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	test_utils "github.com/kajikentaro/flexy-proxy/integration_test"
	"github.com/kajikentaro/flexy-proxy/loggers"
	"github.com/kajikentaro/flexy-proxy/models"
	"github.com/kajikentaro/flexy-proxy/utils"

	"github.com/stretchr/testify/assert"
)

var PROXY_PORT_NUMBER = 8081
var PROXY_HTTP_ADDRESS = fmt.Sprintf(":%d", PROXY_PORT_NUMBER)
var PROXY_URL, _ = url.Parse(
	fmt.Sprintf("http://localhost:%d", PROXY_PORT_NUMBER),
)

var SAMPLE_SERVER_PORT_NUMBER = 8082
var SAMPLE_SERVER_HTTP_ADDRESS = fmt.Sprintf(":%d", SAMPLE_SERVER_PORT_NUMBER)

func TestRequestOnConfigUrl(t *testing.T) {
	config, err := utils.ReadConfigYaml("route_test.yaml")
	assert.NoError(t, err)
	{
		// create a proxy server
		ctx, cancel := context.WithCancel(context.Background())
		err = test_utils.StartProxyServer(ctx, PROXY_HTTP_ADDRESS, config)
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
			res, err := test_utils.Request(PROXY_URL, c.Url)
			assert.NoError(t, err)
			defer res.Body.Close()

			body, err := io.ReadAll(res.Body)
			assert.NoError(t, err)

			assertCommon(t, c, res)
			if c.Response.Content != nil {
				assertContent(t, c, res, body)
				return
			}
			if c.Response.File != nil {
				assertFile(t, c, res, body)
				return
			}
			if c.Response.Rewrite != nil {
				assertRewrite(t, c, res, body)
				return
			}

			assertRewrite(t, c, res, body)
		})
	}
}

func assertCommon(t *testing.T, conf models.Route, res *http.Response) {
	// check content type
	if conf.Response.ContentType != "" {
		expectedContentType := conf.Response.ContentType
		assert.Equal(t, expectedContentType, res.Header.Get("Content-Type"))
	}

	// check status code
	if conf.Response.Status != 0 {
		expectedStatusCode := conf.Response.Status
		assert.Equal(t, expectedStatusCode, res.StatusCode)
	}

	// check custom header
	for key, expected := range conf.Response.Headers {
		actual := res.Header.Get(key)
		assert.Equal(t, expected, actual)
	}
}

func assertFile(t *testing.T, conf models.Route, res *http.Response, body []byte) {
	// check content type set by the handler
	if conf.Response.ContentType == "" {
		expectedContentType := mime.TypeByExtension(filepath.Ext(*conf.Response.File))
		assert.Equal(t, expectedContentType, res.Header.Get("Content-Type"))
	}

	// check status code set by the handler
	if conf.Response.Status == 0 {
		assert.Equal(t, 200, res.StatusCode)
	}

	b, err := os.ReadFile(*conf.Response.File)
	assert.NoError(t, err)
	assert.Equal(t, b, body)
}

func assertContent(t *testing.T, conf models.Route, res *http.Response, body []byte) {
	// check content type set by the handler
	if conf.Response.ContentType == "" {
		assert.Equal(t, "text/plain", res.Header.Get("Content-Type"))
	}

	// check status code set by the handler
	if conf.Response.Status == 0 {
		assert.Equal(t, 200, res.StatusCode)
	}

	assert.Equal(t, *conf.Response.Content, string(body))
}

func assertRewrite(t *testing.T, conf models.Route, res *http.Response, body []byte) {
	// check content type set by the handler
	if conf.Response.ContentType == "" {
		assert.Equal(t, "text/csv", res.Header.Get("Content-Type"))
	}

	// check status code set by the handler
	if conf.Response.Status == 0 {
		assert.Equal(t, 200, res.StatusCode)
	}

	assert.Equal(t, "hello,world", string(body))
}
