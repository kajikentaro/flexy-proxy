package middlewares_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kajikentaro/flexy-proxy/middlewares"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommonMiddleware_RequestHeaders(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	req.Header.Add("User-Agent", "httptest")
	req.Header.Add("Server", "test-server")

	tripper := &dummyRoundTripper{}
	reqHeaderOverwrite := map[string]string{"User-Agent": "test-middleware"}
	middleware := middlewares.NewCommonMiddleware(reqHeaderOverwrite, "", 0, nil, nil, ".")

	res, err := middleware.Middleware(tripper).RoundTrip(req)
	require.NoError(t, err)

	assert.Equal(t, "test-middleware", tripper.captureReq.Header.Get("User-Agent"))
	assert.Equal(t, "test-server", tripper.captureReq.Header.Get("Server"))

	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestCommonMiddleware_ResponseHeaders(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)

	tripper := &dummyRoundTripper{}
	resHeaderOverwrite := map[string]string{"User-Agent": "test-middleware"}
	middleware := middlewares.NewCommonMiddleware(nil, "", 0, resHeaderOverwrite, nil, ".")

	res, err := middleware.Middleware(tripper).RoundTrip(req)
	require.NoError(t, err)

	assert.Equal(t, "test-middleware", res.Header.Get("User-Agent"))
	assert.Equal(t, "text/plain", res.Header.Get("Content-Type"))
	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestCommonMiddleware_ContentType(t *testing.T) {
	middleware := middlewares.NewCommonMiddleware(nil, "application/json", 0, nil, nil, ".")
	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)

	res, err := middleware.Middleware(&dummyRoundTripper{resBody: "test"}).RoundTrip(req)
	require.NoError(t, err)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
}

func TestCommonMiddleware_StatusCode(t *testing.T) {
	tests := []struct {
		name               string
		statusCode         int
		expectedStatusCode int
	}{
		{
			name:               "Zero status code does not override",
			statusCode:         0,
			expectedStatusCode: 200, // dummyRoundTripper's default
		},
		{
			name:               "Override to 404 Not Found",
			statusCode:         404,
			expectedStatusCode: 404,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := middlewares.NewCommonMiddleware(nil, "", tt.statusCode, nil, nil, ".")
			req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)

			res, err := middleware.Middleware(&dummyRoundTripper{resBody: "test"}).RoundTrip(req)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatusCode, res.StatusCode)
		})
	}
}
