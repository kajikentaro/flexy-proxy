package routers

import (
	"net/http"

	"github.com/kajikentaro/elastic-proxy/models"
)

func NewHandleTemplate(handler models.Handler, contentType string, statusCode int, headers map[string]string) models.Handler {
	return &HandleTemplate{
		handler:     handler,
		contentType: contentType,
		statusCode:  statusCode,
		headers:     headers,
	}
}

type HandleTemplate struct {
	handler     models.Handler
	statusCode  int
	contentType string
	headers     map[string]string
}

func (h *HandleTemplate) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.handler.ServeHTTP(w, r)

	if h.contentType != "" {
		// only if the contentType is specified, overwrite
		w.Header().Set("Content-Type", h.contentType)
	}

	if h.statusCode != 0 {
		// only if the statusCode is specified, overwrite
		w.WriteHeader(h.statusCode)
	}

	for v, k := range h.headers {
		w.Header().Set(v, k)
	}
}

func (h *HandleTemplate) GetType() string {
	return h.handler.GetType()
}

func (h *HandleTemplate) GetResponseInfo() map[string]string {
	return h.handler.GetResponseInfo()
}
