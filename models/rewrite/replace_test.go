package rewrite_test

import (
	"net/url"
	"testing"

	"github.com/kajikentaro/elastic-proxy/models/rewrite"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

type DummyStruct struct {
	Rewrite rewrite.Rewrite
}

func TestSingleString(t *testing.T) {
	yamlData := `
rewrite: "http://target.url"
`

	var res DummyStruct
	err := yaml.Unmarshal([]byte(yamlData), &res)
	assert.NoError(t, err)

	input, _ := url.ParseRequestURI("http://original.url")
	actual, err := res.Rewrite.Replace(input)
	assert.NoError(t, err)
	expected, _ := url.ParseRequestURI("http://target.url")

	assert.Equal(t, expected, actual)
}

func TestStringReplacement(t *testing.T) {
	yamlData := `
rewrite:
  from: 'original'
  to: 'replaced'
`

	var res DummyStruct
	err := yaml.Unmarshal([]byte(yamlData), &res)
	assert.NoError(t, err)

	input, _ := url.ParseRequestURI("http://original.url")
	actual, err := res.Rewrite.Replace(input)
	assert.NoError(t, err)
	expected, _ := url.ParseRequestURI("http://replaced.url")

	assert.Equal(t, expected, actual)
}

func TestRegexReplacement(t *testing.T) {
	yamlData := `
rewrite:
  from: 'http://(.*)\.url'
  to: 'http://$1-2.net'
  regex: true
`

	var res DummyStruct
	err := yaml.Unmarshal([]byte(yamlData), &res)
	assert.NoError(t, err)

	input, _ := url.ParseRequestURI("http://original.url")
	actual, err := res.Rewrite.Replace(input)
	assert.NoError(t, err)
	expected, _ := url.ParseRequestURI("http://original-2.net")

	assert.Equal(t, expected, actual)
}
