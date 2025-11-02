package middlewares_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	test_utils "github.com/kajikentaro/flexy-proxy/integration_test"
	"github.com/kajikentaro/flexy-proxy/middlewares"
	"github.com/stretchr/testify/require"
)

var SSE_DURATION = time.Second

type dummyStream struct{}

func (d dummyStream) RoundTrip(_ *http.Request) (*http.Response, error) {
	pr, pw := io.Pipe()

	go func() {
		defer pw.Close()
		for i := 0; i < 5; i++ {
			fmt.Fprintf(pw, "data: SSE message %d\n\n", i)
			time.Sleep(SSE_DURATION)
		}
	}()

	res := &http.Response{
		Body:       pr,
		Header:     make(http.Header),
		Request:    nil,
		StatusCode: http.StatusOK,
	}

	res.Header.Set("Content-Type", "text/event-stream")
	return res, nil
}

func TestDummyStreamItself(t *testing.T) {
	roundTripper := dummyStream{}
	res, err := roundTripper.RoundTrip(nil)
	require.NoError(t, err)
	defer res.Body.Close()

	test_utils.TestSSEStreams(t, res.Body, false)
}

func TestTransformStream(t *testing.T) {
	command := []string{"sed", "-u", "-e", "s/data/DATA/g"}
	transform := middlewares.NewTransform(&command, ".")

	req := httptest.NewRequest(http.MethodGet, "http://example.test", nil)
	res, err := transform.Middleware(dummyStream{}).RoundTrip(req)
	require.NoError(t, err)
	defer res.Body.Close()

	test_utils.TestSSEStreams(t, res.Body, true)
}

func TestTransformStreamStdErr(t *testing.T) {
	command := []string{"bash", "-c", "sed -u -e 's/data/DATA/g' >&2"}
	transform := middlewares.NewTransform(&command, ".")

	req := httptest.NewRequest(http.MethodGet, "http://example.test", nil)
	res, err := transform.Middleware(dummyStream{}).RoundTrip(req)
	require.NoError(t, err)
	defer res.Body.Close()

	test_utils.TestSSEStreams(t, res.Body, true)
}
