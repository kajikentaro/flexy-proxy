package routers

import (
	"bytes"
	"io"
	"net/http"

	"github.com/kajikentaro/flexy-proxy/models"
)

func NewContentResponder(body string) models.RoundTripper {
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
	return res, nil
}

func (c *ContentResponder) GetType() string {
	return "content"
}

func (c *ContentResponder) GetResponseInfo() map[string]string {
	return map[string]string{
		"content": c.body,
	}
}
