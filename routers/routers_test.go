package routers

import (
	"net/url"
	"regexp"
	"testing"

	"github.com/kajikentaro/flexy-proxy/models"
	"github.com/stretchr/testify/assert"
)

func TestIsRegexp(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"hello world\n", false},
		{"\n\t   \r\n", false},
		{`\Qabc\E`, true},
		{"hello*world", true},
		{"\\bword\\b", true},
		{"^start", true},
		{"end$", true},
		{"(group)", true},
		{"abc[def]", true},
		{"", true},
	}

	for _, test := range tests {
		result := isRegexp(test.input)
		if result != test.expected {
			t.Errorf("For input %q, expected %v but got %v", test.input, test.expected, result)
		}
	}
}

func TestValidate(t *testing.T) {
	t.Run("Success case", func(t *testing.T) {
		routes := []route{
			{parsedUrl: mustParseURL(t, "http://example.com")},
			{parsedUrl: mustParseURL(t, "https://secure.com")},
		}
		err := validate(routes, true)
		assert.NoError(t, err)
	})

	t.Run("Invalid scheme (not http or https)", func(t *testing.T) {
		routes := []route{
			{parsedUrl: mustParseURL(t, "ftp://example.com"), Route: &models.Route{Url: "ftp://example.com"}},
		}
		err := validate(routes, true)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "URL Scheme must be 'http' or 'https'.")
	})

	t.Run("Invalid URL (empty hostname)", func(t *testing.T) {
		routes := []route{
			{parsedUrl: mustParseURL(t, "http:///"), Route: &models.Route{Url: "http:///"}},
		}
		err := validate(routes, true)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Invalid URL.")
	})

	t.Run("Regexp condition (shouldDecryptHttps=false && regexUrl!=nil && isRegexp=true)", func(t *testing.T) {
		routes := []route{
			{
				parsedUrl: mustParseURL(t, "http://regexp.example.com"),
				regexUrl:  regexp.MustCompile(".*example.*"),
			},
		}
		err := validate(routes, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Regular expressions are not allowed in the hostname when `always_mitm` is disabled.")
	})
}

// Helper function to parse URLs
func mustParseURL(t *testing.T, raw string) *url.URL {
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("Invalid test URL: %s", raw)
	}
	return u
}
