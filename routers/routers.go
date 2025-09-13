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
	routes []ActiveRoute
}

type ActiveRoute struct {
	http.RoundTripper
	// for logging and debugging
	routeConf  *models.RouteConf
	routeIndex int
	// one side is nil
	parsedUrl *url.URL
	regexUrl  *regexp.Regexp
}

var regHttpOrHttps = regexp.MustCompile(`^https?://`)

func parse(config *models.RawConfig) ([]ActiveRoute, error) {
	var defaultProxy *url.URL
	if config.DefaultRoute.Proxy != "" {
		var err error
		defaultProxy, err = url.Parse(config.DefaultRoute.Proxy)
		if err != nil {
			return nil, err
		}
	}

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

	for i, r := range config.Routes {
		r := r
		pos := fmt.Sprint("route.", i)
		rr := validRoute{
			RouteConf: &r,
			index:     i,
		}

		if !regHttpOrHttps.MatchString(r.Url) {
			return nil, models.NewValidationError(pos, "URL must start with https:// or http://", r.Url)
		}

		if r.Regex {
			regexUrl, err := regexp.Compile("^" + r.Url)
			if err != nil {
				return nil, models.NewValidationError(pos, "Failed to compile regex: %s", r.Url)
			}
			rr.regexUrl = regexUrl
		} else {
			parsedUrl, err := url.Parse(r.Url)
			if err != nil {
				return nil, err
			}
			if parsedUrl.Host == "" {
				return nil, models.NewValidationError(pos, "URL must have a host", r.Url)
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
			h := NewReverseProxyTransport(route.proxyUrl, route.Response.Rewrite, config.InsecureCipherSuites)
			return h, nil
		}

		if route.Response.File != nil {
			h := NewFileResponder(*route.Response.File)
			return h, nil
		}

		// by default, return this
		h := NewReverseProxyTransport(route.proxyUrl, route.Response.Rewrite, config.InsecureCipherSuites)
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

func NewRouter(config *models.RawConfig) (models.Router, error) {
	routes, err := parse(config)
	if err != nil {
		return nil, err
	}

	return &router{routes: routes}, nil
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

	res, err := route.RoundTrip(req)
	if err != nil {
		return nil, nil, err
	}
	res.Header.Set(HEADER_ROUTE_INDEX, fmt.Sprint(route.routeIndex))

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
