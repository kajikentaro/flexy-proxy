package routers

import (
	"net/http"

	"github.com/kajikentaro/elastic-proxy/models"
)

func NewHandleFile(filePath string) models.Handler {
	return &FileHandle{
		filePath: filePath,
	}
}

type FileHandle struct {
	filePath string
}

func (c *FileHandle) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, c.filePath)
}

func (c *FileHandle) GetType() string {
	return "file"
}

func (c *FileHandle) GetResponseInfo() map[string]string {
	return map[string]string{
		"file path": c.filePath,
	}
}
