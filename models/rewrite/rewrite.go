package rewrite

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

/*
There are 3 patters of input.

#1-A:

	input:
	```
	key: "single string"
	```

	behavior:
	Simply Replace the input URL to "single string"

#1-B:

	input:
	```
	key:
		to: "b"
	```

	behavior:
	Simple Replace the input URL to "b". (same as #1)

#2:

	input:
	```
	key:
		from: "a"
		to: "b"
	```

	behavior:
	Replace the "b" in the input URL with "b"

#3:

	input:
	```
	key:
		from: "[a-z]*"
		to: "regex $1"
		regex: true
	```

	behavior:
	Replace the input URL by using regex patterns
*/
type Rewrite struct {
	From  string
	To    string
	Regex bool
	// Proxy setting for this route.
	// If this is nil, default proxy is used.
	// If this is "", no proxy is used.
	Proxy     *string
	ConnectTo string `yaml:"connect_to"`
}

func (u *Rewrite) Replace(inputUrl *url.URL) (*url.URL, error) {
	// No replacement
	if u.To == "" {
		return inputUrl, nil
	}

	// pattern #1-A or #1-B
	// Use "To" without replacing
	if u.From == "" {
		newUrl, err := url.ParseRequestURI(u.To)
		if err != nil {
			return nil, newUrlRewriteError(fmt.Sprintf("invalid url in 'rewrite': %s", u.To), err)
		}
		return newUrl, nil
	}

	// pattern #2
	if !u.Regex {
		inputStr := inputUrl.String()
		newStr := strings.Replace(inputStr, u.From, u.To, -1)
		newUrl, err := url.ParseRequestURI(newStr)
		if err != nil {
			return nil, newUrlRewriteError(fmt.Sprintf("failed to replace %s with %s in %s: the replaced URL is %s", u.From, u.To, inputStr, newStr), nil)
		}
		return newUrl, nil
	}

	// pattern #3
	regex, err := regexp.Compile(u.From)
	if err != nil {
		return nil, newUrlRewriteError("failed to compile regex", err)
	}

	inputStr := inputUrl.String()
	newStr := regex.ReplaceAllString(inputStr, u.To)
	newUrl, err := url.ParseRequestURI(newStr)
	if err != nil {
		return nil, newUrlRewriteError(fmt.Sprintf("failed to replace regex, %s, with %s in %s: the replaced URL is %s", u.From, u.To, inputStr, newStr), nil)
	}
	return newUrl, nil
}

func (e *Rewrite) UnmarshalYAML(value *yaml.Node) error {
	var str string
	if err := value.Decode(&str); err == nil {
		e.To = str
		return nil
	}

	// In order to avoid infinite loop, we need to declare temporary struct which is same as Rewrite
	var tmp struct {
		From      string
		To        string
		Regex     bool
		Proxy     *string
		ConnectTo string `yaml:"connect_to"`
	}
	err := value.Decode(&tmp)
	if err == nil {
		e.From = tmp.From
		e.To = tmp.To
		e.Regex = tmp.Regex
		e.Proxy = tmp.Proxy
		e.ConnectTo = tmp.ConnectTo
		return nil
	}

	return err
}
