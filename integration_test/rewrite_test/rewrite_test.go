package test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	test_utils "github.com/kajikentaro/flexy-proxy/integration_test"
	"github.com/kajikentaro/flexy-proxy/models"
	"github.com/kajikentaro/flexy-proxy/models/rewrite"
	"github.com/kajikentaro/flexy-proxy/proxy"
	"github.com/kajikentaro/flexy-proxy/utils/configs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var PROXY_PORT_NUMBER = test_utils.PORT_NUM_REWRITE
var PROXY_HTTP_ADDRESS = fmt.Sprintf(":%d", PROXY_PORT_NUMBER)
var PROXY_URL, _ = url.Parse(fmt.Sprintf("http://localhost:%d", PROXY_PORT_NUMBER))

func TestMain(m *testing.M) {
	var server *httptest.Server

	// setup sample HTTP server
	{
		server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			message := fmt.Sprintf("[http-ok] host: %s",
				r.Host,
			)
			_, err := io.WriteString(w, message)
			if err != nil {
				panic(err)
			}
		}))
		defer server.Close()
	}

	var tlsServer *httptest.Server

	// setup sample HTTPS server
	{
		tlsServer = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			message := fmt.Sprintf("[https-ok] host: %s, tls-sni: %s",
				r.Host,
				r.TLS.ServerName,
			)
			_, err := io.WriteString(w, message)
			if err != nil {
				panic(err)
			}
		}))
		defer tlsServer.Close()
	}

	// setup proxy server
	{
		serverUrl, _ := url.Parse(server.URL)
		tlsServerUrl, _ := url.Parse(tlsServer.URL)

		rawConfig := models.RawConfig{
			Routes: []models.RouteConf{
				{
					Url:   "http://http.connect-to.test",
					Regex: true,
					Response: models.RouteResponse{
						Rewrite: &rewrite.Rewrite{
							ConnectTo: serverUrl.Host,
						},
					},
				},
				{
					Url:   "https://https.connect-to.test",
					Regex: true,
					Response: models.RouteResponse{
						Rewrite: &rewrite.Rewrite{
							ConnectTo: tlsServerUrl.Host,
						},
					},
				},
			},
			AlwaysMitm: true,
		}

		parsedConfig, err := configs.ParseRawConfig(&rawConfig, ".")
		if err != nil {
			panic(err)
		}

		srv := &http.Server{Addr: PROXY_HTTP_ADDRESS, Handler: proxy.SetupProxy(parsedConfig)}
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			err := srv.ListenAndServe()
			if err != nil {
				panic(err)
			}
			<-ctx.Done()
			err = srv.Shutdown(context.Background())
			if err != nil {
				panic(err)
			}
		}()
		// wait for starting the server
		time.Sleep(100 * time.Millisecond)

		defer cancel()
	}
	m.Run()
}

func TestConnectToOptionWithHTTP(t *testing.T) {
	res, err := test_utils.Request(PROXY_URL, "http://http.connect-to.test")
	require.NoError(t, err)

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	res.Body.Close()

	assert.Equal(t, "[http-ok] host: http.connect-to.test", string(body))
}

func TestConnectToOptionWithHTTPS(t *testing.T) {
	res, err := test_utils.Request(PROXY_URL, "https://https.connect-to.test")
	require.NoError(t, err)

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	res.Body.Close()

	assert.Equal(t, "[https-ok] host: https.connect-to.test, tls-sni: https.connect-to.test", string(body))
}
