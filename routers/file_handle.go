package routers

import (
	"net/http"

	"github.com/kajikentaro/elastic-proxy/models"
)

func NewFileHandler(statusCode int, contentType string, filePath string) models.Handler {
	return &FileHandle{
		statusCode:  statusCode,
		contentType: contentType,
		filePath:    filePath,
	}
}

type FileHandle struct {
	filePath    string
	statusCode  int
	contentType string
}

func (c *FileHandle) Handle(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, c.filePath)

	if c.contentType != "" {
		// only if the contentType is specified, overwrite
		w.Header().Set("Content-Type", c.contentType)
	}

	if c.statusCode != 0 {
		// only if the statusCode is specified, overwrite
		w.WriteHeader(c.statusCode)
	}
}

func (c *FileHandle) GetType() string {
	return "file"
}

func (c *FileHandle) GetResponseInfo() map[string]string {
	return map[string]string{
		"file path": c.filePath,
	}
}
