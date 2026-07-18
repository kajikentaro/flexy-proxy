package rewrite

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

type DummyStruct struct {
	Rewrite *Rewrite `yaml:"rewrite"`
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

func TestMarshalYaml(t *testing.T) {
	t.Run("advanced options", func(t *testing.T) {
		in := DummyStruct{
			Rewrite: &Rewrite{
				From:  "original",
				To:    "replaced",
				Regex: true,
			},
		}
		marshaled, err := yaml.Marshal(in)
		assert.NoError(t, err)
		expected :=
			`rewrite:
    from: original
    to: replaced
    regex: true
    proxy: null
`
		assert.Equal(t, expected, string(marshaled))
	})

	t.Run("empty", func(t *testing.T) {
		in := DummyStruct{
			Rewrite: &Rewrite{},
		}
		marshaled, err := yaml.Marshal(in)
		assert.NoError(t, err)
		expected :=
			`rewrite:
    from: ""
    to: ""
    regex: false
    proxy: null
`
		assert.Equal(t, expected, string(marshaled))
	})
}

func TestReplaceEmpty(t *testing.T) {
	t.Run("'from' and 'to' are empty", func(t *testing.T) {
		r := &Rewrite{}
		input, _ := url.ParseRequestURI("http://original.url")
		actual, err := r.Replace(input)
		assert.NoError(t, err)
		assert.Equal(t, input, actual)
	})

	t.Run("'to' is empty", func(t *testing.T) {
		r := &Rewrite{From: "original"}
		input, _ := url.ParseRequestURI("http://original.url")
		actual, err := r.Replace(input)
		assert.NoError(t, err)
		assert.Equal(t, input, actual)
	})

	t.Run("'from' is empty", func(t *testing.T) {
		r := &Rewrite{To: "http://target.url"}
		input, _ := url.ParseRequestURI("http://original.url")
		actual, err := r.Replace(input)
		assert.NoError(t, err)
		expected, _ := url.ParseRequestURI("http://target.url")
		assert.Equal(t, expected, actual)
	})

}
