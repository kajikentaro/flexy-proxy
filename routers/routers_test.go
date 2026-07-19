package routers

import (
	"net/url"
	"testing"

	"github.com/kajikentaro/flexy-proxy/models"
	"github.com/kajikentaro/flexy-proxy/models/rewrite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	t.Run("Success case", func(t *testing.T) {
		config := &models.RawConfig{
			Routes: []models.RouteConf{
				{Url: "http://example.com"},
				{Url: "https://secure.com"},
			}}
		actual, err := parse(config, ".")
		require.NoError(t, err)

		assert.Nil(t, actual[0].regexUrl)
		assert.NotNil(t, actual[0].parsedUrl)

		assert.Nil(t, actual[1].regexUrl)
		assert.NotNil(t, actual[1].parsedUrl)
	})

	t.Run("Invalid scheme or empty scheme (not http or https)", func(t *testing.T) {

		config := &models.RawConfig{
			Routes: []models.RouteConf{
				{Url: "ftp://example.com"},
				{Url: "example.com"},
			},
		}
		actual, err := parse(config, ".")
		assert.ErrorContains(t, err, "URL must start with https:// or http://")
		assert.Nil(t, actual)
	})

	t.Run("Invalid URL (empty hostname)", func(t *testing.T) {
		config := &models.RawConfig{
			Routes: []models.RouteConf{
				{Url: "http://"},
				{Url: "https://"},
			}}
		actual, err := parse(config, ".")
		assert.ErrorContains(t, err, "URL must have a host")
		assert.Nil(t, actual)
	})

	t.Run("Invalid connect_to", func(t *testing.T) {
		config := &models.RawConfig{
			Routes: []models.RouteConf{
				{Url: "http://example.com"},
			},
		}
		config.Routes[0].Response.Rewrite = &rewrite.Rewrite{
			ConnectTo: "http://invalid.test:80",
		}
		actual, err := parse(config, ".")
		assert.ErrorContains(t, err, "`connect_to` must be \"[host]:[port_num]\"")
		assert.Nil(t, actual)
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

func TestGetMatchedRoute(t *testing.T) {
	config := &models.RawConfig{
		Routes: []models.RouteConf{
			{
				Url: "http://example.test",
			},
			{
				Url: "https://secure.test",
			},
			{
				Url: "http://example.test/path",
			},
			{
				// same as above. should be ignored
				Url: "http://example.test/path",
			},
		}}

	router, err := NewRouter(config, ".")
	require.NoError(t, err)

	tests := []struct {
		input    string
		expected models.RouteConf
	}{
		{"http://example.test", config.Routes[0]},
		{"https://secure.test", config.Routes[1]},
		{"http://example.test/path", config.Routes[2]},
		{"http://example.test/path", config.Routes[2]},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			parsedInput := mustParseURL(t, test.input)
			matchedRoute, err := router.GetMatchedRoute(parsedInput)
			require.NoError(t, err)
			assert.Equal(t, test.expected, matchedRoute)
		})
	}
}

func TestRegexRouteDoesNotMatchQueryParam(t *testing.T) {
	config := &models.RawConfig{
		Routes: []models.RouteConf{
			{
				Url:   "https://foo.dev",
				Regex: true,
			},
		}}
	router, err := NewRouter(config, ".")
	require.NoError(t, err)

	input := "https://example.com?callback=https://foo.dev"
	parsedInput := mustParseURL(t, input)
	_, err = router.GetMatchedRoute(parsedInput)
	assert.Error(t, err, "Should not match route with Regex against query param")
}
