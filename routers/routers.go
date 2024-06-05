package routers

import (
	"fmt"
	"go-proxy/models"
	"net/url"
	"regexp"
)

type router struct {
	routes []route
}

func parse(rawRoutes []models.Route) ([]route, error) {
	var routes []route

	for _, inR := range rawRoutes {
		inR := inR
		newR := route{
			raw: &inR,
		}

		if inR.Regex {
			regexUrl, err := regexp.Compile(inR.Url)
			newR.regexUrl = regexUrl
			if err != nil {
				return nil, err
			}
		}

		parsedUrl, err := url.Parse(inR.Url)
		if err != nil {
			return nil, err
		}
		newR.parsedUrl = parsedUrl

		routes = append(routes, newR)
	}

	return routes, nil
}

func validate(routes []route) error {
	for _, r := range routes {
		if r.parsedUrl.Scheme != "http" && r.parsedUrl.Scheme != "https" {
			return NewValidationError(fmt.Sprintf("scheme of '%s' must be either 'http' or 'https'", r.raw.Url), nil)
		}

		response := r.raw.Response
		if response.File == "" &&
			response.Content == "" &&
			response.Url == nil {
			return NewValidationError("None of File, Content, or Url is not specified", nil)
		}
	}

	return nil
}

func GenRouter(routes []models.Route) (models.Router, error) {
	parsedRoutes, err := parse(routes)
	if err != nil {
		return nil, err
	}

	err = validate(parsedRoutes)
	if err != nil {
		return nil, err
	}

	return &router{routes: parsedRoutes}, nil
}

type route struct {
	parsedUrl *url.URL
	regexUrl  *regexp.Regexp
	raw       *models.Route
}

func (r *router) GetHandler(reqUrl *url.URL) (models.Handler, string, error) {
	for _, route := range r.routes {
		if !isUrlSame(route.parsedUrl, reqUrl) {
			continue
		}
		matchedUrl := route.raw.Url

		if route.raw.Response.Content != "" {
			h := NewContentHandler(route.raw.Response.Status, route.raw.Response.ContentType, route.raw.Response.Content)
			return models.Handler{Content: h}, matchedUrl, nil
		}

		if route.raw.Response.Url != nil {
			newUrl, err := route.raw.Response.Url.Replace(reqUrl)
			if err != nil {
				return models.Handler{}, matchedUrl, err
			}
			h := NewReverseProxyHandler(route.raw.Response.Status, route.raw.Response.ContentType, newUrl)
			return models.Handler{ReverseProxy: h}, matchedUrl, nil
		}

		if route.raw.Response.File != "" {
			h := NewFileHandler(route.raw.Response.Status, route.raw.Response.ContentType, route.raw.Response.File)
			return models.Handler{File: h}, matchedUrl, nil
		}

	}
	return models.Handler{}, "", nil
}

func isUrlSame(a *url.URL, b *url.URL) bool {
	if a.Scheme != b.Scheme {
		return false
	}
	if a.Hostname() != b.Hostname() {
		return false
	}
	pathA := a.EscapedPath()
	if pathA == "" {
		pathA = "/"
	}
	pathB := b.EscapedPath()
	if pathB == "" {
		pathB = "/"
	}
	if pathA != pathB {
		return false
	}
	if a.RawQuery != b.RawQuery {
		return false
	}
	return true
}

func (r *router) GetHttpsHostList() []string {
	var res []string
	for _, route := range r.routes {
		u := route.parsedUrl
		if u.Scheme == "https" {
			var host string
			if u.Port() == "" {
				host = fmt.Sprintf("%s:443", u.Host)
			} else {
				host = u.Host
			}
			res = append(res, host)
		}
	}
	return res
}
