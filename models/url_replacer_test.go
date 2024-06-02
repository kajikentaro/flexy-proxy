package models_test

import (
	"go-proxy/models"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

type DummyStruct struct {
	Url models.UrlReplacement
}

func TestUnmrashalYamlString(t *testing.T) {
	yamlData := `
url: "http://target.url"
`

	var res DummyStruct
	err := yaml.Unmarshal([]byte(yamlData), &res)
	assert.NoError(t, err)

	input, _ := url.ParseRequestURI("http://original.url")
	actual, err := res.Url.Replace(input)
	assert.NoError(t, err)
	expected, _ := url.ParseRequestURI("http://target.url")

	assert.Equal(t, expected, actual)
}

func TestUnmrashalYamlObject(t *testing.T) {
	yamlData := `
url:
  from: 'original'
  to: 'replaced'
`

	var res DummyStruct
	err := yaml.Unmarshal([]byte(yamlData), &res)
	assert.NoError(t, err)

	input, _ := url.ParseRequestURI("http://original.url")
	actual, err := res.Url.Replace(input)
	assert.NoError(t, err)
	expected, _ := url.ParseRequestURI("http://replaced.url")

	assert.Equal(t, expected, actual)
}
