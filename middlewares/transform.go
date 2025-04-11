package middlewares

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
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

func (t *Transform) Middleware(next http.RoundTripper) http.RoundTripper {
	return roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		// NOTE: if the response body is compressed, we can't use string replacement commands like 'sed'.
		r.Header.Del("Accept-Encoding")

		res, err := next.RoundTrip(r)
		if err != nil {
			return nil, err
		}

		cmd := exec.Command((*t.command)[0], (*t.command)[1:]...)
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

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}
