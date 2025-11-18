package middlewares

import (
	"net/http"
)

func NewCommonMiddleware(reqHeaders map[string]string, contentType string, statusCode int, resHeaders map[string]string, parsedTransformCommand *[]string, workDir string) *CommonMiddleware {
	return &CommonMiddleware{
		contentType:            contentType,
		statusCode:             statusCode,
		reqHeaders:             reqHeaders,
		resHeaders:             resHeaders,
		parsedTransformCommand: parsedTransformCommand,
		workDir:                workDir,
	}
}

type CommonMiddleware struct {
	statusCode             int
	contentType            string
	reqHeaders             map[string]string
	resHeaders             map[string]string
	parsedTransformCommand *[]string
	workDir                string
}

func (h *CommonMiddleware) Middleware(next http.RoundTripper) http.RoundTripper {
	return roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		var res *http.Response
		var err error

		for k, v := range h.reqHeaders {
			r.Header.Set(k, v)
		}

		if h.parsedTransformCommand == nil {
			res, err = next.RoundTrip(r)
		} else {
			transform := NewTransform(h.parsedTransformCommand, h.workDir)
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

		for v, k := range h.resHeaders {
			res.Header.Set(v, k)
		}

		return res, nil
	})
}
