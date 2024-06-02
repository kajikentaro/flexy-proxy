package models

import (
	"fmt"
	"net/url"
	"reflect"
	"strings"

	"gopkg.in/yaml.v3"
)

type UrlReplacement struct {
	From string
	To   string
}

func (u UrlReplacement) Replace(input *url.URL) (*url.URL, error) {
	inputStr := input.String()
	replacedStr := u.replaceStr(inputStr)

	replaced, err := url.ParseRequestURI(replacedStr)
	if err != nil {
		return nil, fmt.Errorf("failed to replace %s with %s in %s", u.From, u.To, inputStr)
	}

	return replaced, nil
}

func (u UrlReplacement) replaceStr(inputUrl string) string {
	if u.From == "" {
		return u.To
	}
	return strings.Replace(inputUrl, u.From, u.To, -1)
}

func (e *UrlReplacement) UnmarshalYAML(value *yaml.Node) error {
	var aux interface{}
	if err := value.Decode(&aux); err != nil {
		return err
	}

	switch raw := aux.(type) {
	case string:
		e.From = ""
		e.To = raw
	case map[string]interface{}:
		mapToStruct(raw, e)
	}
	return nil

}

// capitalize a initial letter
func capitalize(s string) string {
	if len(s) == 0 {
		return ""
	}
	return string(s[0]-'a'+'A') + s[1:]
}

func mapToStruct(input map[string]interface{}, output interface{}) {
	for key, value := range input {
		ckey := capitalize(key)
		if field := reflect.ValueOf(output).Elem().FieldByName(ckey); field.IsValid() {
			if field.CanSet() {
				field.SetString(value.(string))
			}
		}
	}
}
