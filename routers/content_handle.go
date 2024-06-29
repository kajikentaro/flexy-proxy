package routers

import (
	"net/http"

	"github.com/kajikentaro/elastic-proxy/models"
)

func NewContentHandler(statusCode int, contentType string, body string) models.Handler {
	return &ContentHandle{
		statusCode:  statusCode,
		contentType: contentType,
		body:        body,
	}
}

type ContentHandle struct {
	body        string
	statusCode  int
	contentType string
}

func (c *ContentHandle) Handle(w http.ResponseWriter, r *http.Request) {
	contentType := "text/plain"
	if c.contentType != "" {
		contentType = c.contentType
	}

	statusCode := 200
	if c.statusCode != 0 {
		statusCode = c.statusCode
	}

	w.Write([]byte(c.body))
	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", contentType)
}

func (c *ContentHandle) GetType() string {
	return "content"
}

func (c *ContentHandle) GetResponseInfo() map[string]string {
	return map[string]string{
		"content": c.body,
	}
}
