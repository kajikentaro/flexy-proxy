package integration_test

import (
	"fmt"
	"go-proxy"
	test_utils "go-proxy/integration_test"
	"go-proxy/models"
	default_proxy "go-proxy/to_be_remove"
	"go-proxy/utils"
	"io"
	"net/http"
	"net/url"
	"os"
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

	p := proxy.GetProxy(config)
	srv := &http.Server{Addr: PROXY_HTTP_ADDRESS, Handler: p}
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

	for _, c := range config.Routes {
		proxyUrl, err := url.Parse(PROXY_URL)
		assert.NoError(t, err)
		res, err := test_utils.Request(proxyUrl, c.Url)
		assert.NoError(t, err)
		defer res.Body.Close()
		body, err := io.ReadAll(res.Body)

		if c.Content != "" {
			assert.NoError(t, err)
			assert.Equal(t, c.Content, string(body))
			// TODO read expected headers from config file
			assert.Equal(t, res.Header.Get("Content-Type"), "text/html;charset=utf-8")
			continue
		}

		if c.File != "" {
			b, err := os.ReadFile(c.File)
			assert.NoError(t, err)
			assert.Equal(t, b, body)
			// TODO read expected headers from config file
			assert.Equal(t, res.Header.Get("Content-Type"), "text/html;charset=utf-8")
			continue
		}

		t.Error("invalid config format")
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
