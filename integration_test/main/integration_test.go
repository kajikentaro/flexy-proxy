package integration_test

import (
	"fmt"
	"go-proxy"
	test_utils "go-proxy/integration_test"
	"go-proxy/loggers"
	"go-proxy/models"
	default_proxy "go-proxy/to_be_remove"
	"go-proxy/utils"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	yaml "gopkg.in/yaml.v2"
)

var PROXY_PORT_NUMBER = 8081
var PROXY_HTTP_ADDRESS = fmt.Sprintf(":%d", PROXY_PORT_NUMBER)
var PROXY_URL = fmt.Sprintf("http://localhost:%d", PROXY_PORT_NUMBER)

func TestMain(m *testing.M) {
	config, err := utils.ParseConfig("")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	proxy, err := proxy.SetupProxy(config, loggers.GenLogger())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	srv := &http.Server{Addr: PROXY_HTTP_ADDRESS, Handler: proxy}
	go test_utils.StartServer(srv)
	// wait for starting the server
	time.Sleep(time.Second)
	defer test_utils.StopServer(srv)
	m.Run()
}

func TestRequestOnConfigUrl(t *testing.T) {
	dir, err := os.Getwd()
	assert.NoError(t, err)
	fileName := dir + "/config.yaml"
	fileContent, err := os.ReadFile(fileName)
	assert.NoError(t, err)

	var config models.ProxyConfig
	err = yaml.Unmarshal(fileContent, &config)
	assert.NoError(t, err)

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
			} else {
				t.Error("invalid config format")
			}
		})
	}
}

func TestRequestOnOtherUrl(t *testing.T) {
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
