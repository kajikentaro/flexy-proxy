package version

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVerifyVersion(t *testing.T) {
	cases := []struct {
		Current      string
		Required     string
		ErrorMessage string
	}{
		{
			Current:  "v0.2.0",
			Required: "v0.1.11",
		},
		{
			Current:  "v0.2.0",
			Required: "v0.2.0",
		},
		{
			Current:      "v0.2.0",
			Required:     "v0.2.11",
			ErrorMessage: "Flexy Proxy doesn't support this configuration. Please upgrade Flexy Proxy to v0.2.11 or later",
		},
		{
			Current: "unknown",
		},
		{
			Current: "v0.2.0",
		},
		{
			Required: "v0.2.0",
		},
		{
			Required:     "0.2.0",
			Current:      "v0.2.0",
			ErrorMessage: "format of required version must be \"v[major].[minor].[patch]\"",
		},
		{
			Required:     "v0.2.0",
			Current:      "0.2.0",
			ErrorMessage: "format of build version must be \"v[major].[minor].[patch]\"",
		},
		{
			Required:     "v0.2",
			Current:      "v0.2.0",
			ErrorMessage: "format of required version must be \"v[major].[minor].[patch]\"",
		},
		{
			Required:     "vX.Y.Z",
			Current:      "v0.2.0",
			ErrorMessage: "invalid required version \"vX.Y.Z\". failed to parse \"X\" to int",
		},
		{
			Required:     "v0.2.0",
			Current:      "vX.Y.Z",
			ErrorMessage: "invalid build version \"vX.Y.Z\". failed to parse \"X\" to int",
		},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("current: %s, required: %s", c.Current, c.Required), func(t *testing.T) {
			actual := VerifyVersion(c.Required, c.Current)

			if c.ErrorMessage == "" {
				assert.Nil(t, actual)
			} else {
				assert.Error(t, actual)
				assert.Equal(t, c.ErrorMessage, actual.Error())
			}
		})
	}
}
