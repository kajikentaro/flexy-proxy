package routers

import (
	"bytes"
	"io"
	"net/http"
)

func NewContentResponder(body string) http.RoundTripper {
	return &ContentResponder{
		body: body,
	}
}

type ContentResponder struct {
	body string
}

func (c *ContentResponder) RoundTrip(r *http.Request) (*http.Response, error) {
	res := &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewReader([]byte(c.body))),
	}
	res.Header.Set("Content-Type", "text/plain")
	res.Header.Set(HEADER_RESPONSE_TYPE, "content")
	return res, nil
}
