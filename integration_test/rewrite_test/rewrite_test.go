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
				{
					Url:   "http://http.basic-rewrite.test",
					Regex: true,
					Response: models.RouteResponse{
						Rewrite: &rewrite.Rewrite{
							// Use localhost:[port] instead of [IP]:[localhost] in order to check Hostname
							To: "http://localhost:" + serverUrl.Port(),
						},
					},
				},
				{
					Url:   "https://https.basic-rewrite.test",
					Regex: true,
					Response: models.RouteResponse{
						Rewrite: &rewrite.Rewrite{
							// Use localhost:[port] instead of [IP]:[localhost] in order to check Hostname
							To: "https://localhost:" + tlsServerUrl.Port(),
						},
					},
				},
				{
					Url:   "http://http.overwrite-host.test",
					Regex: true,
					Response: models.RouteResponse{
						Rewrite: &rewrite.Rewrite{
							// Use localhost:[port] instead of [IP]:[localhost] in order to check Hostname
							To: "http://localhost:" + serverUrl.Port(),
						},
					},
					Request: models.RouteRequest{
						Headers: map[string]string{"Host": "replaced-by-header"},
					},
				},
				{
					Url:   "https://https.overwrite-host.test",
					Regex: true,
					Response: models.RouteResponse{
						Rewrite: &rewrite.Rewrite{
							// Use localhost:[port] instead of [IP]:[localhost] in order to check Hostname
							To: "https://localhost:" + tlsServerUrl.Port(),
						},
					},
					Request: models.RouteRequest{
						Headers: map[string]string{"Host": "replaced-by-header"},
					},
				},
				{
					Url:   "http://http.all.test",
					Regex: true,
					Response: models.RouteResponse{
						Rewrite: &rewrite.Rewrite{
							ConnectTo: serverUrl.Host,
							To:        "http://specified-by-rewrite.test:" + serverUrl.Port(),
						},
					},
					Request: models.RouteRequest{
						Headers: map[string]string{"Host": "replaced-by-header"},
					},
				},
				{
					Url:   "https://https.all.test",
					Regex: true,
					Response: models.RouteResponse{
						Rewrite: &rewrite.Rewrite{
							ConnectTo: tlsServerUrl.Host,
							To:        "https://specified-by-rewrite.test:" + tlsServerUrl.Port(),
						},
					},
					Request: models.RouteRequest{
						Headers: map[string]string{"Host": "replaced-by-header"},
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

	assert.Equal(t, "[http-ok] host: http.connect-to.test", string(body), "hostname must be same as request")
}

func TestConnectToOptionWithHTTPS(t *testing.T) {
	res, err := test_utils.Request(PROXY_URL, "https://https.connect-to.test")
	require.NoError(t, err)

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	res.Body.Close()

	assert.Equal(t, "[https-ok] host: https.connect-to.test, tls-sni: https.connect-to.test", string(body), "hostname and tls-sni must be same as request")
}

func TestBasicRewriteWithHTTP(t *testing.T) {
	res, err := test_utils.Request(PROXY_URL, "http://http.basic-rewrite.test")
	require.NoError(t, err)

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	res.Body.Close()

	assert.Regexp(t, "^\\[http-ok\\] host: localhost:[0-9]+$", string(body), "hostname must be same as \"to\"")
}

func TestBasicRewriteWithHTTPS(t *testing.T) {
	res, err := test_utils.Request(PROXY_URL, "https://https.basic-rewrite.test")
	require.NoError(t, err)

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	res.Body.Close()

	assert.Regexp(t, "^\\[https-ok\\] host: localhost:[0-9]+, tls-sni: localhost$", string(body), "hostname and tls-sni must be same as \"to\"")
}

func TestOverwriteHostWithHTTP(t *testing.T) {
	res, err := test_utils.Request(PROXY_URL, "http://http.overwrite-host.test")
	require.NoError(t, err)

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	res.Body.Close()

	assert.Equal(t, "[http-ok] host: replaced-by-header", string(body), "hostname must be replaced")
}

func TestOverwriteHostWithHTTPS(t *testing.T) {
	res, err := test_utils.Request(PROXY_URL, "https://https.overwrite-host.test")
	require.NoError(t, err)

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	res.Body.Close()

	assert.Equal(t, "[https-ok] host: replaced-by-header, tls-sni: localhost", string(body), "only hostname must be replaced")
}

func TestAllAreSpecifiedWithHTTP(t *testing.T) {
	res, err := test_utils.Request(PROXY_URL, "http://http.all.test")
	require.NoError(t, err)

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	res.Body.Close()

	assert.Equal(t, "[http-ok] host: replaced-by-header", string(body), "hostname must be replaced to header's one")
}

func TestAllAreSpecifiedWithHTTPS(t *testing.T) {
	res, err := test_utils.Request(PROXY_URL, "https://https.all.test")
	require.NoError(t, err)

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	res.Body.Close()

	// Reason
	// 1. Flexy receive a request with "https://https.all.test".
	// 2. Flexy replace the URL to "https://specified-by-rewrite.test:xxxx". This is SNI.
	// 3. Flexy connect to the server "https://127.0.0.1:xxxx" which is specified on "connect_to".
	// 4. Flexy send hostname "replaced-by-header" which is specified on "headers".
	// As a result, we can get a response from https://127.0.0.1:xxxx but SNI and hostname are different.
	assert.Equal(t, "[https-ok] host: replaced-by-header, tls-sni: specified-by-rewrite.test", string(body), "hostname must be replaced by header and tls-sni must be replaced by rewrite")
}
