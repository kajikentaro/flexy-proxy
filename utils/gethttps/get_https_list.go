package gethttps

import (
	"fmt"
	"net/url"
	"regexp"

	"github.com/kajikentaro/flexy-proxy/models"
)

var regOrigin = regexp.MustCompile(`^https?://[^/]+`)
var regHttps = regexp.MustCompile(`^https://`)

func GetHttpsHostList(routes []models.RouteConf) ([]string, error) {
	var res []string
	for i, route := range routes {
		if !regHttps.MatchString(route.Url) {
			continue
		}

		pos := fmt.Sprint("route.", i)
		hostname, err := getHostname(route, pos)
		if err != nil {
			return nil, err
		}
		hostname = fmt.Sprintf("%s:443", hostname)
		res = append(res, hostname)
	}

	return res, nil
}

func getHostname(inR models.RouteConf, pos string) (string, error) {
	if !inR.Regex {
		parsedUrl, err := url.Parse(inR.Url)
		if err != nil {
			return "", err
		}
		return parsedUrl.Host, nil
	}

	// `originStr` would be 'https://foo\.example\.com'
	originStr := regOrigin.FindString(inR.Url)
	if isRegexp(originStr) {
		return "", models.NewValidationError(pos, "regular expressions are not allowed in the hostname when `always_mitm` is false", originStr)
	}
	// `originPlained` would be 'https://foo.example.com'
	originPlained, err := decodeRegexpEscape(originStr)
	if err != nil {
		return "", models.NewValidationError(pos, fmt.Sprintf("failed to decode regex: %s", err), originPlained)
	}
	url, err := url.Parse(originPlained)
	if err != nil {
		return "", models.NewValidationError(pos, fmt.Sprintf("failed to parse decoded regex: %s", err), originPlained)
	}
	return url.Host, nil
}
