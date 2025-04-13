package middlewares

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

func NewTransform(command *[]string) *Transform {
	return &Transform{
		command: command,
	}
}

type Transform struct {
	command *[]string
}

func isProbablyText(contentType string) bool {
	if contentType == "" ||
		strings.HasPrefix(contentType, "text/") ||
		contentType == "application/json" ||
		contentType == "application/xml" ||
		contentType == "application/x-www-form-urlencoded" {
		return true
	}
	return false
}

func (t *Transform) Middleware(next http.RoundTripper) http.RoundTripper {
	return roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		// NOTE: if the response body is compressed, we can't use string replacement commands like 'sed'.
		r.Header.Del("Accept-Encoding")

		res, err := next.RoundTrip(r)
		if err != nil {
			return nil, err
		}

		var reqBody []byte
		if r.ContentLength > 1024*1024 {
			reqBody = []byte("BODY env variable is only available for requests with Content-Length less than 1MB")
		} else if !isProbablyText(r.Header.Get("Content-Type")) {
			reqBody = []byte("BODY env variable is only available for text content types")
		} else {
			var err error
			reqBody, err = io.ReadAll(r.Body)
			if err != nil {
				return nil, err
			}
			r.Body = io.NopCloser(bytes.NewReader(reqBody))
		}

		cmd := exec.Command((*t.command)[0], (*t.command)[1:]...)
		cmd.Env = append(os.Environ(), "REQ_BODY="+string(reqBody))
		cmd.Stdin = res.Body

		var stdout bytes.Buffer
		cmd.Stdout = &stdout
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			return nil, fmt.Errorf("failed to execute command: '%s'\nError Log: \n%s", strings.Join((*t.command), " "), stderr.String())
		}

		res.Body = io.NopCloser(&stdout)
		return res, nil
	})
}
