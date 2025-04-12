package routers

import (
	"mime"
	"net/http"
	"os"
	"path/filepath"

	"github.com/kajikentaro/flexy-proxy/models"
)

func NewFileResponder(filePath string) models.RoundTripper {
	return &FileResponder{
		filePath: filePath,
	}
}

type FileResponder struct {
	filePath string
}

func (c *FileResponder) RoundTrip(r *http.Request) (*http.Response, error) {
	file, err := os.Open(c.filePath)
	if err != nil {
		return nil, err
	}
	res := &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       file,
	}
	ctype := mime.TypeByExtension(filepath.Ext(c.filePath))
	res.Header.Set("Content-Type", ctype)
	return res, nil
}

func (c *FileResponder) GetType() string {
	return "file"
}

func (c *FileResponder) GetResponseInfo() map[string]string {
	return map[string]string{
		"file_path": c.filePath,
	}
}
