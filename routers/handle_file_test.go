package routers

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFileResponder_AbsoluteAndRelativePath(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "testfile-*.txt")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	content := "test file content"
	_, err = tmpFile.WriteString(content)
	require.NoError(t, err)
	tmpFile.Close()

	dir, file := filepath.Split(tmpFile.Name())

	t.Run("relative path", func(t *testing.T) {
		roundTripper, err := NewFileResponder(file, dir)
		require.NoError(t, err)
		req, err := http.NewRequest("GET", "http://file.test", nil)
		require.NoError(t, err)
		resp, err := roundTripper.RoundTrip(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Equal(t, content, string(body))
	})

	t.Run("absolute path", func(t *testing.T) {
		roundTripper, err := NewFileResponder(tmpFile.Name(), "")
		require.NoError(t, err)
		req, err := http.NewRequest("GET", "http://file.test", nil)
		require.NoError(t, err)
		resp, err := roundTripper.RoundTrip(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Equal(t, content, string(body))
	})
}
