package video

import (
	"net/http"

	"github.com/kaustavdm/awwdio/config"
)

type Handler struct {
	config *config.Config
}

func NewHandler(cfg *config.Config) *Handler {
	return &Handler{
		config: cfg,
	}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /token", h.tokenHandler)
}
