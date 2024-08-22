package proxy

import (
	"bytes"
	"io"
	"net/http"
)

type ResponseWrite struct {
	*http.Response
	body *bytes.Buffer
}

func (f *ResponseWrite) Header() http.Header {
	return f.Response.Header
}

func (f *ResponseWrite) Write(b []byte) (int, error) {
	return f.body.Write(b)
}

func (f *ResponseWrite) WriteHeader(statusCode int) {
	f.Response.StatusCode = statusCode
}

// usage:
// fileRes := NewResponseWriter(req)
// fileRes implements `http.ResponseWriter`
// fileRes.res is `*http.Response`
func NewResponseWriter(req *http.Request) *ResponseWrite {
	var body bytes.Buffer
	return &ResponseWrite{
		Response: &http.Response{
			Body:       io.NopCloser(&body),
			Header:     make(http.Header),
			Request:    req,
			StatusCode: http.StatusOK,
		},
		body: &body,
	}
}
