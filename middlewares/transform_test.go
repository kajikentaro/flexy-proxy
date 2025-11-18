package middlewares_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kajikentaro/flexy-proxy/middlewares"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type dummyRoundTripper struct {
	resBody    string
	captureReq *http.Request
}

func (d *dummyRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	d.captureReq = req
	res := httptest.NewRecorder()
	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusOK)
	res.Body.Write([]byte(d.resBody))
	_, err := io.ReadAll(req.Body)
	if err != nil {
		panic(err)
	}
	return res.Result(), nil
}

func TestTransformMiddleware(t *testing.T) {

	tests := []struct {
		name           string
		requestBody    string
		responseBody   string
		contentType    string
		command        []string
		expectedOutput string
	}{
		{
			name:           "Transform response body",
			requestBody:    "",
			responseBody:   "foo",
			contentType:    "text/plain",
			command:        []string{"sed", "-E", "s/foo/bar/g"},
			expectedOutput: "bar",
		},
		{
			name:           "Log request body",
			requestBody:    "test request body",
			responseBody:   "foo",
			contentType:    "text/plain",
			command:        []string{"bash", "-c", "echo $REQ_BODY"},
			expectedOutput: "test request body\n",
		},
		{
			name:           "$REQ_HEADER and $RES_HEADER are set",
			requestBody:    "",
			responseBody:   "",
			contentType:    "text/plain; test=req-header",
			command:        []string{"bash", "-c", "echo $REQ_HEADER; echo $RES_HEADER"},
			expectedOutput: "{\"Content-Type\":[\"text/plain; test=req-header\"]}\n{\"Content-Type\":[\"text/plain\"]}\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transform := middlewares.NewTransform(&tt.command, ".")

			req := httptest.NewRequest(http.MethodPost, "http://example.com", bytes.NewReader([]byte(tt.requestBody)))
			req.Header.Set("Content-Type", tt.contentType)

			res, err := transform.Middleware(&dummyRoundTripper{resBody: tt.responseBody}).RoundTrip(req)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, res.StatusCode)
			assert.Equal(t, "text/plain", res.Header.Get("Content-Type"))
			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedOutput, string(resBody))
		})
	}
}

func TestTransformMiddlewareErrorCase(t *testing.T) {
	command := []string{"invalid_command"}
	transform := middlewares.NewTransform(&command, ".")

	// Test for error case when the command is invalid
	req := httptest.NewRequest(http.MethodPost, "http://example.com", nil)
	req.Header.Set("Content-Type", "text/plain")

	res, err := transform.Middleware(&dummyRoundTripper{resBody: "foo"}).RoundTrip(req)
	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestTransformMiddlewareLargeBody(t *testing.T) {
	command := []string{"bash", "-c", "echo $REQ_BODY"}
	transform := middlewares.NewTransform(&command, ".")

	largeBody := bytes.Repeat([]byte("a"), 1024*1024+1)

	// Test for large body handling
	req := httptest.NewRequest(http.MethodPost, "http://example.com", bytes.NewReader(largeBody))
	req.Header.Set("Content-Type", "text/plain")

	res, err := transform.Middleware(&dummyRoundTripper{resBody: "foo"}).RoundTrip(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
	resBody, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	errorMessage := "BODY env variable is only available for requests with Content-Length less than 1MB\n"
	assert.Equal(t, errorMessage, string(resBody))
}

func TestTransformMiddlewareNonTextContent(t *testing.T) {
	command := []string{"bash", "-c", "echo $REQ_BODY"}
	transform := middlewares.NewTransform(&command, ".")

	// Test for non-text content handling
	req := httptest.NewRequest(http.MethodPost, "http://example.com", bytes.NewReader([]byte("binary data")))
	req.Header.Set("Content-Type", "application/octet-stream")

	res, err := transform.Middleware(&dummyRoundTripper{resBody: "foo"}).RoundTrip(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
	resBody, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	errorMessage := "BODY env variable is only available for text content types\n"
	assert.Equal(t, errorMessage, string(resBody))
}

func TestURLEnvironmentVariable(t *testing.T) {
	command := []string{"bash", "-c", "echo $URL"}
	transform := middlewares.NewTransform(&command, ".")

	req := httptest.NewRequest(http.MethodPost, "http://example.com", nil)
	res, err := transform.Middleware(&dummyRoundTripper{}).RoundTrip(req)
	require.NoError(t, err)

	resBody, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	assert.Equal(t, "http://example.com\n", string(resBody))
}

func TestWorkingDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	command := []string{"pwd"}
	transform := middlewares.NewTransform(&command, tmpDir)

	req := httptest.NewRequest(http.MethodPost, "http://example.com", nil)
	res, err := transform.Middleware(&dummyRoundTripper{}).RoundTrip(req)
	require.NoError(t, err)

	resBody, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	assert.Equal(t, tmpDir+"\n", string(resBody))
}
