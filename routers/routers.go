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
	routes        []ActiveRoute
	httpsHostList []string
}

var regOrigin = regexp.MustCompile(`^https?://[^/]+`)
var regHttpOrHttps = regexp.MustCompile(`^https?://`)
var regHttps = regexp.MustCompile(`^https://`)

type ActiveRoute struct {
	http.RoundTripper
	// for logging and debugging
	routeConf  *models.RouteConf
	routeIndex int
	// one side is nil
	parsedUrl *url.URL
	regexUrl  *regexp.Regexp
}

func parse(routeConfList []models.RouteConf, defaultProxy *url.URL) ([]ActiveRoute, error) {
	type validRoute struct {
		// pre-set values
		*models.RouteConf
		index int

		// parsed values
		proxyUrl               *url.URL
		parsedTransformCommand *[]string

		// one side is nil
		parsedUrl *url.URL
		regexUrl  *regexp.Regexp
	}

	var validRoutes []validRoute

	for i, r := range routeConfList {
		r := r
		pos := fmt.Sprint("route.", i)
		rr := validRoute{
			RouteConf: &r,
			index:     i,
		}

		if !regHttpOrHttps.MatchString(r.Url) {
			return nil, NewValidationError(pos, "URL must start with https:// or http://", r.Url)
		}

		if r.Regex {
			regexUrl, err := regexp.Compile("^" + r.Url)
			if err != nil {
				return nil, NewValidationError(pos, "Failed to compile regex: %s", r.Url)
			}
			rr.regexUrl = regexUrl
		} else {
			parsedUrl, err := url.Parse(r.Url)
			if err != nil {
				return nil, err
			}
			if parsedUrl.Host == "" {
				return nil, NewValidationError(pos, "URL must have a host", r.Url)
			}
			rr.parsedUrl = parsedUrl
		}

		if r.Response.Transform != "" {
			parsedCommand, err := shlex.Split(r.Response.Transform)
			if err != nil {
				return nil, err
			}
			rr.parsedTransformCommand = &parsedCommand
		}

		if r.Response.Rewrite == nil {
			rr.proxyUrl = defaultProxy
		} else {
			if r.Response.Rewrite.Proxy == nil {
				rr.proxyUrl = defaultProxy
			} else if *r.Response.Rewrite.Proxy == "" {
				rr.proxyUrl = nil
			} else {
				parsedProxyUrl, err := url.ParseRequestURI(*r.Response.Rewrite.Proxy)
				if err != nil {
					return nil, err
				}
				rr.proxyUrl = parsedProxyUrl
			}
		}

		validRoutes = append(validRoutes, rr)
	}

	getMainRoundTripper := func(route *validRoute) (http.RoundTripper, error) {
		if route.Response.Content != nil {
			h := NewContentResponder(*route.Response.Content)
			return h, nil
		}

		if route.Response.Rewrite != nil {
			h := NewReverseProxyTransport(route.proxyUrl, route.Response.Rewrite)
			return h, nil
		}

		if route.Response.File != nil {
			h := NewFileResponder(*route.Response.File)
			return h, nil
		}

		// by default, return this
		h := NewReverseProxyTransport(route.proxyUrl, route.Response.Rewrite)
		return h, nil
	}

	var roundTrippers []ActiveRoute
	for _, r := range validRoutes {
		main, err := getMainRoundTripper(&r)
		if err != nil {
			return nil, err
		}

		common := middlewares.NewCommonMiddleware(
			r.Response.ContentType,
			r.Response.Status,
			r.Response.Headers,
			r.parsedTransformCommand,
		)

		route := ActiveRoute{
			RoundTripper: common.Middleware(main),
			parsedUrl:    r.parsedUrl,
			regexUrl:     r.regexUrl,
			routeConf:    r.RouteConf,
			routeIndex:   r.index,
		}
		roundTrippers = append(roundTrippers, route)
	}

	return roundTrippers, nil
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
		return "", NewValidationError(pos, "Regular expressions are not allowed in the hostname when `always_mitm` is false.", originStr)
	}
	// `originPlained` would be 'https://foo.example.com'
	originPlained, err := decodeRegexpEscape(originStr)
	if err != nil {
		return "", NewValidationError(pos, fmt.Sprintf("Failed to decode regex: %s", err), originPlained)
	}
	url, err := url.Parse(originPlained)
	if err != nil {
		return "", NewValidationError(pos, fmt.Sprintf("Failed to parse decoded regex: %s", err), originPlained)
	}
	return url.Host, nil
}

func calcHttpsHostList(routes []models.RouteConf) ([]string, error) {
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

func NewRouter(routeConfList []models.RouteConf, defaultProxy *url.URL, shouldDecryptHttps bool) (models.Router, error) {
	routes, err := parse(routeConfList, defaultProxy)
	if err != nil {
		return nil, err
	}

	var httpsHostList []string
	if !shouldDecryptHttps {
		httpsHostList, err = calcHttpsHostList(routeConfList)
		if err != nil {
			return nil, err
		}
	}

	return &router{routes: routes, httpsHostList: httpsHostList}, nil
}

func (r *router) GetMatchedRoute(url *url.URL) (models.RouteConf, error) {
	parsedRoute, err := r.getMatchedRoute(url)
	if err != nil {
		return models.RouteConf{}, err
	}
	return *parsedRoute.routeConf, nil
}

func (r *router) getMatchedRoute(url *url.URL) (*ActiveRoute, error) {
	for _, route := range r.routes {
		if !isUrlSame(url, route) {
			continue
		}
		return &route, nil
	}
	return nil, models.ErrRouteNotFound
}

func (r *router) TryRoundTrip(req *http.Request) (map[string]string, *http.Response, error) {
	route, err := r.getMatchedRoute(req.URL)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set(HEADER_ROUTE_INDEX, fmt.Sprint(route.routeConf))

	res, err := route.RoundTrip(req)
	if err != nil {
		return nil, nil, err
	}

	log := make(map[string]string)
	for headerK, logK := range HEADER_KEY_TO_LOG_KEY {
		if v := res.Header.Get(headerK); v != "" {
			log[logK] = v
		}
	}

	res.Request = req
	return log, res, nil
}

func isUrlSame(in *url.URL, route ActiveRoute) bool {
	if route.regexUrl != nil {
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
	return r.httpsHostList
}
