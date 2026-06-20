package test

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"testing"

	test_utils "github.com/kajikentaro/flexy-proxy/integration_test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var PROXY_PORT_NUMBER = test_utils.PORT_NUM_ALWAYS_MITM
var PROXY_HTTP_ADDRESS = fmt.Sprintf(":%d", PROXY_PORT_NUMBER)
var PROXY_URL, _ = url.Parse(
	fmt.Sprintf("http://localhost:%d", PROXY_PORT_NUMBER),
)

func TestMain(m *testing.M) {
	// setup proxy server
	{
		ctx, cancel := context.WithCancel(context.Background())
		err := test_utils.StartProxyServer(ctx, PROXY_HTTP_ADDRESS, "always_mitm_test.yaml")
		if err != nil {
			fmt.Fprintln(os.Stderr, "failed to start a proxy server:", err)
			os.Exit(1)
		}
		defer cancel()
	}
	m.Run()
}

func TestRegexUrl(t *testing.T) {
	res, err := test_utils.Request(PROXY_URL, "https://with-regex.test/foo")
	require.NoError(t, err)
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, "with-regex.test", res.Request.URL.Host)
	assert.Equal(t, "with regex", string(body))
}

func TestNotRegexUrl(t *testing.T) {
	res, err := test_utils.Request(PROXY_URL, "https://not-regex.test")
	require.NoError(t, err)
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, "not-regex.test", res.Request.URL.Host)
	assert.Equal(t, "not regex", string(body))
}

func TestOutOfRoute(t *testing.T) {
	res, err := test_utils.Request(PROXY_URL, "https://out-of-route.test")
	assert.Error(t, err)
	assert.Nil(t, res)
}
