package routers

import (
	"mime"
	"net/http"
	"os"
	"path/filepath"
)

func NewFileResponder(file string, workDir string) (http.RoundTripper, error) {
	fileAbsPath := file
	if !filepath.IsAbs(fileAbsPath) {
		joined := filepath.Join(workDir, file)
		_fileAbsPath, err := filepath.Abs(joined)
		if err != nil {
			return nil, err
		}
		fileAbsPath = _fileAbsPath
	}

	return &FileResponder{
		filePath: fileAbsPath,
	}, nil
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
