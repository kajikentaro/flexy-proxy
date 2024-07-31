package routers

import (
	"net/http"

	"github.com/kajikentaro/elastic-proxy/models"
)

func NewContentHandler(body string) models.Handler {
	return &ContentHandle{
		body: body,
	}
}

type ContentHandle struct {
	body string
}

func (c *ContentHandle) Handle(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(c.body))
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "text/plain")
}

func (c *ContentHandle) GetType() string {
	return "content"
}

func (c *ContentHandle) GetResponseInfo() map[string]string {
	return map[string]string{
		"content": c.body,
	}
}
