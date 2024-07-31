package routers

import (
	"net/http"

	"github.com/kajikentaro/elastic-proxy/models"
)

func NewHandleTemplate(handler models.Handler, contentType string, statusCode int) models.Handler {
	return &HandleTemplate{
		handler:     handler,
		contentType: contentType,
		statusCode:  statusCode,
	}
}

type HandleTemplate struct {
	handler     models.Handler
	statusCode  int
	contentType string
}

func (h *HandleTemplate) Handle(w http.ResponseWriter, r *http.Request) {
	h.handler.Handle(w, r)

	if h.contentType != "" {
		// only if the contentType is specified, overwrite
		w.Header().Set("Content-Type", h.contentType)
	}

	if h.statusCode != 0 {
		// only if the statusCode is specified, overwrite
		w.WriteHeader(h.statusCode)
	}
}

func (h *HandleTemplate) GetType() string {
	return h.handler.GetType()
}

func (h *HandleTemplate) GetResponseInfo() map[string]string {
	return h.handler.GetResponseInfo()
}
