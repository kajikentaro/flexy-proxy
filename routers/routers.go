package routers

import (
	"fmt"
	"net/url"
	"regexp"

	"github.com/google/shlex"
	"github.com/kajikentaro/flexy-proxy/models"
)

type router struct {
	routes []route
}

func parse(rawRoutes []models.Route, defaultProxy *url.URL) ([]route, error) {
	var routes []route

	for _, inR := range rawRoutes {
		inR := inR
		newR := route{
			Route: &inR,
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

		if inR.Response.Transform != "" {
			parsedCommand, err := shlex.Split(inR.Response.Transform)
			if err != nil {
				return nil, err
			}
			newR.parsedTransformCommand = &parsedCommand
		}

		if inR.Response.Rewrite == nil {
			newR.proxyUrl = defaultProxy
		} else {
			if inR.Response.Rewrite.Proxy == nil {
				newR.proxyUrl = defaultProxy
			} else if *inR.Response.Rewrite.Proxy == "" {
				newR.proxyUrl = nil
			} else {
				parsedProxyUrl, err := url.ParseRequestURI(*inR.Response.Rewrite.Proxy)
				if err != nil {
					return nil, err
				}
				newR.proxyUrl = parsedProxyUrl
			}
		}

		routes = append(routes, newR)
	}

	return routes, nil
}

func validate(routes []route) error {
	for _, r := range routes {
		if r.parsedUrl.Scheme != "http" && r.parsedUrl.Scheme != "https" {
			return NewValidationError(fmt.Sprintf("scheme of '%s' must be either 'http' or 'https'", r.Url))
		}
	}

	return nil
}

func GenRouter(routes []models.Route, defaultProxy *url.URL) (models.Router, error) {
	parsedRoutes, err := parse(routes, defaultProxy)
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
	proxyUrl  *url.URL
	*models.Route
	parsedTransformCommand *[]string
}

func (r *router) getMainHandler(route route, reqUrl *url.URL) (models.Handler, error) {
	if route.Response.Content != nil {
		h := NewHandleContent(*route.Response.Content)
		return h, nil
	}

	if route.Response.Rewrite != nil {
		newUrl, err := route.Response.Rewrite.Replace(reqUrl)
		if err != nil {
			return nil, err
		}
		h := NewHandleReverseProxy(newUrl, route.proxyUrl)
		return h, nil
	}

	if route.Response.File != nil {
		h := NewHandleFile(*route.Response.File)
		return h, nil
	}

	// by default, return this
	h := NewHandleReverseProxy(reqUrl, route.proxyUrl)
	return h, nil
}

func (r *router) GetHandler(reqUrl *url.URL) (models.Handler, string, error) {
	for _, route := range r.routes {
		if !isUrlSame(reqUrl, route) {
			continue
		}

		handler, err := r.getMainHandler(route, reqUrl)
		if err != nil {
			return nil, "", err
		}

		handler = NewHandleTemplate(
			handler,
			route.Response.ContentType,
			route.Response.Status,
			route.Response.Headers,
			route.parsedTransformCommand,
		)

		return handler, route.Url, nil
	}
	return nil, "", models.ErrRouteNotFound
}

func isUrlSame(in *url.URL, route route) bool {
	if route.Regex {
		return route.regexUrl.MatchString(in.String())
	}

	if in.Scheme != route.parsedUrl.Scheme {
		return false
	}
	if in.Hostname() != route.parsedUrl.Hostname() {
		return false
	}
	pathA := in.EscapedPath()
	if pathA == "" {
		pathA = "/"
	}
	pathB := route.parsedUrl.EscapedPath()
	if pathB == "" {
		pathB = "/"
	}
	if pathA != pathB {
		return false
	}
	if in.RawQuery != route.parsedUrl.RawQuery {
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

func (r *router) GetUrlList() []string {
	var res []string
	for _, route := range r.routes {
		res = append(res, route.Url)
	}
	return res
}
