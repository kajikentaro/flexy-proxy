package routers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockUrlReplacer struct {
	toUrl string
}

func (m *mockUrlReplacer) Replace(u *url.URL) (*url.URL, error) {
	return url.Parse(m.toUrl)
}

func sendRequest(t *testing.T, roundTripper http.RoundTripper, url string) *http.Response {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	require.NoError(t, err)

	resp, err := roundTripper.RoundTrip(req)
	require.NoError(t, err)

	return resp
}

func TestReverseProxyTransportHostPriority(t *testing.T) {
	var gotHost string
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHost = r.Host
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "ok")
	}))
	defer target.Close()
	rewriter := &mockUrlReplacer{toUrl: target.URL}

	t.Run("configured Host header has highest priority", func(t *testing.T) {
		roundTripper := NewReverseProxyTransport(nil, rewriter, false, map[string]string{
			"host": "forced.test",
		})

		resp := sendRequest(t, roundTripper, "http://original.test/path")
		defer resp.Body.Close()

		assert.Equal(t, "forced.test", gotHost)
		assert.Equal(t, "rewrite", resp.Header.Get(HEADER_RESPONSE_TYPE))
		assert.Equal(t, target.URL, resp.Header.Get(HEADER_REWRITE_TO))
	})

	t.Run("rewritten URL host is used when Host header is not configured", func(t *testing.T) {
		roundTripper := NewReverseProxyTransport(nil, rewriter, false, nil)

		resp := sendRequest(t, roundTripper, "http://original.test/path")
		defer resp.Body.Close()

		assert.Equal(t, mustParseURL(t, target.URL).Host, gotHost)
		assert.Equal(t, "rewrite", resp.Header.Get(HEADER_RESPONSE_TYPE))
		assert.Equal(t, target.URL, resp.Header.Get(HEADER_REWRITE_TO))
	})

	t.Run("original Host is used when there is no configured Host and no rewrite", func(t *testing.T) {
		roundTripper := NewReverseProxyTransport(nil, nil, false, nil)

		url := target.URL + "/test"
		resp := sendRequest(t, roundTripper, url)
		defer resp.Body.Close()

		assert.Equal(t, mustParseURL(t, url).Host, gotHost)
		assert.Equal(t, url, resp.Header.Get(HEADER_REWRITE_TO))
	})
}
