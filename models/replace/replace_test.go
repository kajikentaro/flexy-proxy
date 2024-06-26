package replace_test

import (
	"net/url"
	"testing"

	"github.com/kajikentaro/elastic-proxy/models/replace"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

type DummyStruct struct {
	Url replace.Url
}

func TestSingleString(t *testing.T) {
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

func TestStringReplacement(t *testing.T) {
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

func TestRegexReplacement(t *testing.T) {
	yamlData := `
url:
  from: 'http://(.*)\.url'
  to: 'http://$1-2.net'
  regex: true
`

	var res DummyStruct
	err := yaml.Unmarshal([]byte(yamlData), &res)
	assert.NoError(t, err)

	input, _ := url.ParseRequestURI("http://original.url")
	actual, err := res.Url.Replace(input)
	assert.NoError(t, err)
	expected, _ := url.ParseRequestURI("http://original-2.net")

	assert.Equal(t, expected, actual)
}
