package middlewares

import (
	"net/http"
)

func NewCommonMiddleware(contentType string, statusCode int, headers map[string]string, parsedTransformCommand *[]string) *CommonMiddleware {
	return &CommonMiddleware{
		contentType:            contentType,
		statusCode:             statusCode,
		headers:                headers,
		parsedTransformCommand: parsedTransformCommand,
	}
}

type CommonMiddleware struct {
	statusCode             int
	contentType            string
	headers                map[string]string
	parsedTransformCommand *[]string
}

func (h *CommonMiddleware) Middleware(next http.RoundTripper) http.RoundTripper {
	return roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		var res *http.Response
		var err error

		if h.parsedTransformCommand == nil {
			res, err = next.RoundTrip(r)
		} else {
			transform := NewTransform(h.parsedTransformCommand)
			res, err = transform.Middleware(next).RoundTrip(r)
		}

		if err != nil {
			return nil, err
		}

		if h.contentType != "" {
			// only if the contentType is specified, overwrite
			res.Header.Set("Content-Type", h.contentType)
		}

		if h.statusCode != 0 {
			// only if the statusCode is specified, overwrite
			res.StatusCode = h.statusCode
		}

		for v, k := range h.headers {
			res.Header.Set(v, k)
		}

		return res, nil
	})
}
