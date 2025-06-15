package middlewares

import (
	"bytes"
	"encoding/json"
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

		reqHeader, err := json.Marshal(r.Header)
		if err != nil {
			return nil, err
		}

		resHeader, err := json.Marshal(res.Header)
		if err != nil {
			return nil, err
		}

		cmd := exec.Command((*t.command)[0], (*t.command)[1:]...)
		env := append(os.Environ(),
			"REQ_BODY="+string(reqBody),
			"REQ_HEADER="+string(reqHeader),
			"RES_HEADER="+string(resHeader),
		)
		cmd.Env = env
		cmd.Stdin = res.Body
		pr, pw := io.Pipe()
		cmd.Stdout = pw
		cmd.Stderr = pw

		res.Body = pr

		if err := cmd.Start(); err != nil {
			return nil, err
		}

		go func() {
			cmd.Wait()
			pw.Close()
		}()
		return res, nil
	})
}
