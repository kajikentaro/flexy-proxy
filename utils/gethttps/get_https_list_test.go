package gethttps_test

import (
	"testing"

	"github.com/kajikentaro/flexy-proxy/models"
	"github.com/kajikentaro/flexy-proxy/utils/gethttps"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalcHttpsHostList(t *testing.T) {
	t.Run("isRegexp:true && HTTP", func(t *testing.T) {
		routes := []models.RouteConf{
			{
				Url:   "http://.*example.*",
				Regex: true,
			},
		}
		actual, err := gethttps.GetHttpsHostList(routes)
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
		actual, err := gethttps.GetHttpsHostList(routes)
		assert.ErrorContains(t, err, "regular expressions are not allowed in the hostname when `always_mitm` is false")
		assert.Nil(t, actual)
	})
}
