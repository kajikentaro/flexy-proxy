package routers

import (
	"go-proxy/models"
	"net/http"
)

func NewContentHandler(statusCode int, contentType string, body string) models.ContentHandler {
	return &contentHandle{
		statusCode:  statusCode,
		contentType: contentType,
		body:        body,
	}
}

type contentHandle struct {
	body        string
	statusCode  int
	contentType string
}

func (c *contentHandle) Handler(w http.ResponseWriter, r *http.Request) {
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

func (c *contentHandle) Content() string {
	return c.body
}
