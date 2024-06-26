package routers

import (
	"net/http"

	"github.com/kajikentaro/elastic-proxy/models"
)

func NewFileHandler(statusCode int, contentType string, filePath string) models.FileHandler {
	return &fileHandle{
		statusCode:  statusCode,
		contentType: contentType,
		filePath:    filePath,
	}
}

type fileHandle struct {
	filePath    string
	statusCode  int
	contentType string
}

func (c *fileHandle) Handler(w http.ResponseWriter, r *http.Request) {
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

func (c *fileHandle) FilePath() string {
	return c.filePath
}
