package routers

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"

	"github.com/google/shlex"
	"github.com/kajikentaro/flexy-proxy/middlewares"
	"github.com/kajikentaro/flexy-proxy/models"
)

type router struct {
	routes []parsedRoute
}

var regOrigin = regexp.MustCompile(`^https?://[^/]+`)
var regHttp = regexp.MustCompile(`^https?://`)

func parse(rawRoutes []models.Route, defaultProxy *url.URL, shouldDecryptHttps bool) ([]parsedRoute, error) {
	var routes []parsedRoute

	for i, inR := range rawRoutes {
		pos := fmt.Sprint("route.", i)
		inR := inR
		newR := parsedRoute{
			Route: &inR,
		}

		if !regHttp.MatchString(inR.Url) {
			return nil, NewValidationError(pos, "URL must start with https:// or http://", inR.Url)
		}

		if inR.Regex {
			regexUrl, err := regexp.Compile(inR.Url)
			if err != nil {
				return nil, NewValidationError(pos, "Failed to compile regex: %s", inR.Url)
			}
			newR.regexUrl = regexUrl
		} else {
			parsedUrl, err := url.Parse(inR.Url)
			if err != nil {
				return nil, err
			}
			newR.parsedUrl = parsedUrl
		}

		getOrigin := func() (*origin, error) {
			if shouldDecryptHttps {
				return nil, nil
			}

			if !inR.Regex {
				parsedUrl, err := url.Parse(inR.Url)
				if err != nil {
					return nil, err
				}
				if parsedUrl.Host == "" {
					return nil, NewValidationError(pos, "Invalid URL.", inR.Url)
				}
				return &origin{scheme: parsedUrl.Scheme, host: parsedUrl.Host}, nil
			}

			// `originStr` would be 'https://foo\.example\.com'
			originStr := regOrigin.FindString(inR.Url)
			if isRegexp(originStr) {
				return nil, NewValidationError(pos, "Regular expressions are not allowed in the hostname when `always_mitm` is false.", originStr)
			}
			// `originPlained` would be 'https://foo.example.com'
			originPlained, err := decodeRegexpEscape(originStr)
			if err != nil {
				return nil, NewValidationError(pos, fmt.Sprintf("Failed to decode regex: %s", err), originPlained)
			}
			url, err := url.Parse(originPlained)
			if err != nil {
				return nil, NewValidationError(pos, fmt.Sprintf("Failed to parse decoded regex: %s", err), originPlained)
			}
			return &origin{scheme: url.Scheme, host: url.Host}, err
		}

		origin, err := getOrigin()
		if err != nil {
			return nil, err
		}
		newR.origin = origin

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

func GenRouter(routes []models.Route, defaultProxy *url.URL, shouldDecryptHttps bool) (models.Router, error) {
	parsedRoutes, err := parse(routes, defaultProxy, shouldDecryptHttps)
	if err != nil {
		return nil, err
	}

	return &router{routes: parsedRoutes}, nil
}

type origin struct {
	scheme string
	host   string
}

type parsedRoute struct {
	proxyUrl *url.URL
	*models.Route
	parsedTransformCommand *[]string

	// one side is nil
	parsedUrl *url.URL
	regexUrl  *regexp.Regexp

	// if `always_mitm` is true (default), this will be nil
	origin *origin
}

func (r *router) getMainRoundTripper(route *parsedRoute, reqUrl *url.URL) (models.RoundTripper, error) {
	if route.Response.Content != nil {
		h := NewContentResponder(*route.Response.Content)
		return h, nil
	}

	if route.Response.Rewrite != nil {
		newUrl, err := route.Response.Rewrite.Replace(reqUrl)
		if err != nil {
			return nil, err
		}
		h := NewReverseProxyTransport(newUrl, route.proxyUrl)
		return h, nil
	}

	if route.Response.File != nil {
		h := NewFileResponder(*route.Response.File)
		return h, nil
	}

	// by default, return this
	h := NewReverseProxyTransport(reqUrl, route.proxyUrl)
	return h, nil
}

func (r *router) GetMatchedRoute(url *url.URL) (models.Route, error) {
	parsedRoute, err := r.getMatchedParsedRoute(url)
	if err != nil {
		return models.Route{}, err
	}
	return *parsedRoute.Route, nil
}

func (r *router) getMatchedParsedRoute(url *url.URL) (*parsedRoute, error) {
	for _, route := range r.routes {
		if !isUrlSame(url, route) {
			continue
		}
		return &route, nil
	}
	return nil, models.ErrRouteNotFound
}

func (r *router) TryRoundTrip(req *http.Request) (map[string]string, *http.Response, error) {
	reqUrl := req.URL
	route, err := r.getMatchedParsedRoute(reqUrl)
	if err != nil {
		return nil, nil, err
	}

	common := middlewares.NewCommonMiddleware(
		route.Response.ContentType,
		route.Response.Status,
		route.Response.Headers,
		route.parsedTransformCommand,
	)

	main, err := r.getMainRoundTripper(route, reqUrl)
	if err != nil {
		return nil, nil, err
	}

	res, err := common.Middleware(main).RoundTrip(req)
	if err != nil {
		return nil, nil, err
	}

	info := main.GetResponseInfo()
	info["matched_url"] = reqUrl.String()
	info["type"] = main.GetType()

	res.Header.Add("flexy-proxy", fmt.Sprintf("matched route: %s", route.Url))
	res.Request = req

	return info, res, nil
}

func isUrlSame(in *url.URL, route parsedRoute) bool {
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
		host := fmt.Sprintf("%s:443", route.origin.host)
		res = append(res, host)
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
