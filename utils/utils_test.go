package utils

import (
	"testing"

	"github.com/kajikentaro/flexy-proxy/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// https://github.com/kajikentaro/flexy-proxy/issues/7
func TestAlwaysMitmWithRegex(t *testing.T) {
	proxyConfig, err := parseRawConfig(
		&models.RawConfig{
			Routes:     []models.RouteConf{{Url: "https://example\\.test", Regex: true}, {Url: "https://foo.test"}},
			AlwaysMitm: true,
		})
	require.NoError(t, err)

	assert.Empty(t, proxyConfig.HttpsHostNames)
}
