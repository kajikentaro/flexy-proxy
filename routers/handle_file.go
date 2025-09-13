package routers

import (
	"mime"
	"net/http"
	"os"
	"path/filepath"
)

func NewFileResponder(filePath string) http.RoundTripper {
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
	res.Header.Set(HEADER_RESPONSE_TYPE, "file")
	res.Header.Set(HEADER_FILE_PATH, c.filePath)
	return res, nil
}
