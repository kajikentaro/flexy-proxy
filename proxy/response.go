package proxy

import (
	"bytes"
	"io"
	"net/http"
)

type Response struct {
	res  *http.Response
	body io.ReadWriter
}

func (f *Response) Header() http.Header {
	return f.res.Header
}

func (f *Response) Write(b []byte) (int, error) {
	return f.body.Write(b)
}

func (f *Response) WriteHeader(statusCode int) {
	f.res.StatusCode = statusCode
}

// usage:
// fileRes := NewResponseWriter(req)
// fileRes implements `http.ResponseWriter`
// fileRes.res is `*http.Response`
func NewResponseWriter(req *http.Request) *Response {
	var body bytes.Buffer
	return &Response{
		res: &http.Response{
			Body:       io.NopCloser(&body),
			Header:     make(http.Header),
			Request:    req,
			StatusCode: http.StatusOK,
		},
		body: &body,
	}
}
