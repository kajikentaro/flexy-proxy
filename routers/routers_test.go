package routers

import (
	"net/url"
	"testing"

	"github.com/kajikentaro/flexy-proxy/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsRegexp(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"hello world\n", false},
		{"\n\t   \r\n", false},
		{"hello*world", true},
		{"\\bword\\b", true},
		{"^start", true},
		{"end$", true},
		{"(group)", true},
		{"abc[def]", true},
		{"", true},
		// NOTE: \Q and \E are used to escape special characters in regex but go does not support it (just ignores them)
		//       once they are supported, this test should be updated
		{`\Qabc\E`, false},
	}

	for _, test := range tests {
		result := isRegexp(test.input)
		if result != test.expected {
			t.Errorf("For input %q, expected %v but got %v", test.input, test.expected, result)
		}
	}
}

func TestParse(t *testing.T) {
	t.Run("Success case", func(t *testing.T) {
		routes := []models.RouteConf{
			{Url: "http://example.com"},
			{Url: "https://secure.com"},
		}
		actual, err := parse(routes, nil)
		require.NoError(t, err)

		assert.Nil(t, actual[0].regexUrl)
		assert.NotNil(t, actual[0].parsedUrl)

		assert.Nil(t, actual[1].regexUrl)
		assert.NotNil(t, actual[1].parsedUrl)
	})

	t.Run("Invalid scheme or empty scheme (not http or https)", func(t *testing.T) {
		routes := []models.RouteConf{
			{Url: "ftp://example.com"},
			{Url: "example.com"},
		}
		actual, err := parse(routes, nil)
		assert.ErrorContains(t, err, "URL must start with https:// or http://")
		assert.Nil(t, actual)
	})

	t.Run("Invalid URL (empty hostname)", func(t *testing.T) {
		routes := []models.RouteConf{
			{Url: "http://"},
			{Url: "https://"},
		}
		actual, err := parse(routes, nil)
		assert.ErrorContains(t, err, "URL must have a host")
		assert.Nil(t, actual)
	})

}

func TestCalcHttpsHostList(t *testing.T) {
	t.Run("isRegexp:true && HTTP", func(t *testing.T) {
		routes := []models.RouteConf{
			{
				Url:   "http://.*example.*",
				Regex: true,
			},
		}
		actual, err := calcHttpsHostList(routes)
		require.NoError(t, err)
		assert.Empty(t, actual)
	})

	t.Run("isRegexp:true && HTTPS", func(t *testing.T) {
		routes := []models.RouteConf{
			{
				Url:   "https://.*example.*",
				Regex: true,
			},
		}
		actual, err := calcHttpsHostList(routes)
		assert.ErrorContains(t, err, "Regular expressions are not allowed in the hostname when `always_mitm` is false.")
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
	routes := []models.RouteConf{
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
	}

	router, err := NewRouter(routes, nil, true)
	require.NoError(t, err)

	tests := []struct {
		input    string
		expected models.RouteConf
	}{
		{"http://example.test", routes[0]},
		{"https://secure.test", routes[1]},
		{"http://example.test/path", routes[2]},
		{"http://example.test/path", routes[2]},
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

// https://github.com/kajikentaro/flexy-proxy/issues/7
func TestAlwaysMitmWithRegex(t *testing.T) {
	router, err := NewRouter([]models.RouteConf{{Url: "https://example\\.test", Regex: true}, {Url: "https://foo.test"}}, nil, true)
	require.NoError(t, err)

	hostList := router.GetHttpsHostList()
	assert.Empty(t, hostList)
}

func TestRegexRouteDoesNotMatchQueryParam(t *testing.T) {
	routes := []models.RouteConf{
		{
			Url:   "https://foo.dev",
			Regex: true,
		},
	}
	router, err := NewRouter(routes, nil, true)
	require.NoError(t, err)

	input := "https://example.com?callback=https://foo.dev"
	parsedInput := mustParseURL(t, input)
	_, err = router.GetMatchedRoute(parsedInput)
	assert.Error(t, err, "Should not match route with Regex against query param")
}
