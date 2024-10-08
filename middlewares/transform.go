package middlewares

import (
	"bytes"
	"fmt"
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

func (t *Transform) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextResponse := &responseWriter{ResponseWriter: w}

		// NOTE: if the response body is compressed, we can't use string replacement commands like 'sed'.
		r.Header.Del("Accept-Encoding")

		next.ServeHTTP(nextResponse, r)

		cmd := exec.Command((*t.command)[0], (*t.command)[1:]...)
		cmd.Stdin = &nextResponse.body

		var stdout bytes.Buffer
		cmd.Stdout = &stdout
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			http.Error(w, fmt.Sprintf("Failed to execute command: '%s'\nError Log: \n%s", strings.Join((*t.command), " "), stderr.String()), http.StatusInternalServerError)
			return
		}

		_, err := w.Write(stdout.Bytes())
		if err != nil {
			http.Error(w, "Failed to write response", http.StatusInternalServerError)
		}
	})
}

type responseWriter struct {
	http.ResponseWriter
	body bytes.Buffer
}

func (w *responseWriter) Write(b []byte) (int, error) {
	return w.body.Write(b)
}
