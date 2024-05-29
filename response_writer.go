package proxy

import (
	"bytes"
	"io"
	"net/http"
)

type FileResponse struct {
	res  *http.Response
	body io.ReadWriter
}

func (f *FileResponse) Header() http.Header {
	return f.res.Header
}

func (f *FileResponse) Write(b []byte) (int, error) {
	return f.body.Write(b)
}

func (f *FileResponse) WriteHeader(statusCode int) {
	f.res.StatusCode = statusCode
}

// usage:
// fileRes := NewFileResponse(req)
// fileRes implements `http.ResponseWriter`
// fileRes.res implements `*http.Response`
func NewFileResponse(req *http.Request) *FileResponse {
	var body bytes.Buffer
	return &FileResponse{
		res: &http.Response{
			Body:    io.NopCloser(&body),
			Header:  make(http.Header),
			Request: req,
		},
		body: &body,
	}
}
