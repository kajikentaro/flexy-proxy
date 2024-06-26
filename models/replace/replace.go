package replace

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

/*
There are 3 patters of input.

#1:

	input:
	```
	key: "single string"
	```

	behavior:
	Simply Replace the input URL to "single string"

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
type Url struct {
	SingleUrl string
	UrlParts
}

type UrlParts struct {
	From     string
	To       string
	Regex    bool
	ProxyUrl string `yaml:"proxy_url"`
}

func (u *Url) Replace(inputUrl *url.URL) (*url.URL, error) {
	// pattern #1
	if u.SingleUrl != "" {
		newUrl, err := url.ParseRequestURI(u.SingleUrl)
		if err != nil {
			return nil, NewUrlReplaceError(fmt.Sprintf("invalid url: %s", u.SingleUrl), err)
		}
		return newUrl, nil
	}

	// pattern #2
	if !u.Regex {
		inputStr := inputUrl.String()
		newStr := strings.Replace(inputStr, u.From, u.To, -1)
		newUrl, err := url.ParseRequestURI(newStr)
		if err != nil {
			return nil, NewUrlReplaceError(fmt.Sprintf("failed to replace %s with %s in %s: the replaced URL is %s", u.From, u.To, inputStr, newStr), nil)
		}
		return newUrl, nil
	}

	// pattern #3
	regex, err := regexp.Compile(u.From)
	if err != nil {
		return nil, NewUrlReplaceError("failed to compile regex", err)
	}

	inputStr := inputUrl.String()
	newStr := regex.ReplaceAllString(inputStr, u.To)
	newUrl, err := url.ParseRequestURI(newStr)
	if err != nil {
		return nil, NewUrlReplaceError(fmt.Sprintf("failed to replace regex, %s, with %s in %s: the replaced URL is %s", u.From, u.To, inputStr, newStr), nil)
	}
	return newUrl, nil
}

func (e *Url) UnmarshalYAML(value *yaml.Node) error {
	var str string
	if err := value.Decode(&str); err == nil {
		e.SingleUrl = str
		return nil
	}

	var urlParts UrlParts
	err := value.Decode(&urlParts)
	if err == nil {
		e.UrlParts = urlParts
		return nil
	}

	return err
}
