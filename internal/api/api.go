package api

import (
	"net/http"

	"github.com/kaustavdm/awwdio/config"
	"github.com/kaustavdm/awwdio/internal/api/video"
)

type API struct {
	config *config.Config

	// Load sub-APIs
	videoHandler *video.Handler
}

func New(cfg *config.Config) *API {
	videoH := video.NewHandler(cfg)
	return &API{
		config:       cfg,
		videoHandler: videoH,
	}
}

func (a *API) Register(mux *http.ServeMux) {
	// Register video mux
	videoMux := http.NewServeMux()
	a.videoHandler.Register(videoMux)
	mux.Handle("/video/", http.StripPrefix("/video", videoMux))
}
