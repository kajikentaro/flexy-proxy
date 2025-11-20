package middlewares

import (
	"io"
	"net/http"
	"os/exec"
	"strings"
)

func NewTransform(command *[]string, workDir string) *Transform {
	return &Transform{
		command: command,
		workDir: workDir,
	}
}

type Transform struct {
	command *[]string
	workDir string
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
		res, err := next.RoundTrip(r)
		if err != nil {
			return nil, err
		}

		res.ContentLength = -1
		res.Header.Del("Content-Length")

		// string "hello world" to io.ReadCloser
		// b := io.NopCloser(strings.NewReader("hello world"))

		cmd := exec.Command((*t.command)[0], (*t.command)[1:]...)
		// cmd := exec.Command("echo", "hello world")
		cmd.Stdin = res.Body
		pr, pw := io.Pipe()
		cmd.Stdout = pw
		cmd.Stderr = pw

		res.Body = pr

		cmd.Dir = t.workDir
		if err := cmd.Start(); err != nil {
			return nil, err
		}

		go func() {
			cmd.Wait()
			res.Body.Close()
			pw.Close()
		}()
		return res, nil
	})
}
