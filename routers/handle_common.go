package routers

import (
	"net/http"

	"github.com/kajikentaro/flexy-proxy/middlewares"
	"github.com/kajikentaro/flexy-proxy/models"
)

func NewHandleCommon(handler models.RoundTripper, contentType string, statusCode int, headers map[string]string, parsedTransformCommand *[]string) models.RoundTripper {
	return &CommonWrapper{
		handler:                handler,
		contentType:            contentType,
		statusCode:             statusCode,
		headers:                headers,
		parsedTransformCommand: parsedTransformCommand,
	}
}

type CommonWrapper struct {
	handler                models.RoundTripper
	statusCode             int
	contentType            string
	headers                map[string]string
	parsedTransformCommand *[]string
}

func (h *CommonWrapper) RoundTrip(r *http.Request) (*http.Response, error) {
	var res *http.Response
	var err error

	if h.parsedTransformCommand == nil {
		res, err = h.handler.RoundTrip(r)
	} else {
		transform := middlewares.NewTransform(h.parsedTransformCommand)
		res, err = transform.Middleware(h.handler).RoundTrip(r)
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
}

func (h *CommonWrapper) GetType() string {
	return h.handler.GetType()
}

func (h *CommonWrapper) GetResponseInfo() map[string]string {
	return h.handler.GetResponseInfo()
}
