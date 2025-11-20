package test

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"testing"

	test_utils "github.com/kajikentaro/flexy-proxy/integration_test"
	"github.com/kajikentaro/flexy-proxy/loggers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var PROXY_PORT_NUMBER = 8087
var PROXY_HTTP_ADDRESS = fmt.Sprintf(":%d", PROXY_PORT_NUMBER)
var PROXY_URL = fmt.Sprintf("http://localhost:%d", PROXY_PORT_NUMBER)

var SAMPLE_SERVER_PORT_NUMBER = 8088
var SAMPLE_SERVER_HTTP_ADDRESS = fmt.Sprintf(":%d", SAMPLE_SERVER_PORT_NUMBER)

func fatalln(a ...any) {
	fmt.Fprintln(os.Stderr, a...)
	os.Exit(1)
}

func TestMain(m *testing.M) {
	{
		// create a proxy server
		ctx, cancel := context.WithCancel(context.Background())
		err := test_utils.StartProxyServer(ctx, PROXY_HTTP_ADDRESS, "transform_test.yaml")
		if err != nil {
			fatalln("failed to start a proxy server:", err)
		}
		defer cancel()
	}
	{
		// create a sample http server to return "hello world"
		ctx, cancel := context.WithCancel(context.Background())
		err := test_utils.StartSampleHttpServer(ctx, SAMPLE_SERVER_HTTP_ADDRESS, loggers.GenLogger(nil))
		if err != nil {
			fatalln("failed to start a http server:", err)
		}
		defer cancel()
	}
	m.Run()
}

type TestCase struct {
	title        string
	url          string
	expectedBody string
}

var testCases = []TestCase{
	{
		title:        "replace http reverse proxy response with wc command",
		url:          "http://reverse-proxy.test/",
		expectedBody: "hello world\n",
	},
}

func TestTransform(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			proxyUrl, err := url.Parse(PROXY_URL)
			require.NoError(t, err)

			res, err := test_utils.Request(proxyUrl, tc.url)
			assert.NoError(t, err)
			defer res.Body.Close()

			fmt.Println(res.Header.Get("Content-Length"))

			body, err := io.ReadAll(res.Body)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedBody, string(body))
		})
	}
}
