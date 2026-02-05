package api

import (
	"net/http"

	"github.com/kaustavdm/awwdio/config"
	"github.com/kaustavdm/awwdio/internal/api/auth"
	"github.com/kaustavdm/awwdio/internal/api/middleware"
	"github.com/kaustavdm/awwdio/internal/api/video"
)

type API struct {
	config *config.Config

	// Load sub-APIs
	authHandler  *auth.Handler
	videoHandler *video.Handler
}

func New(cfg *config.Config) *API {
	authH := auth.NewHandler(cfg)
	videoH := video.NewHandler(cfg)
	return &API{
		config:       cfg,
		authHandler:  authH,
		videoHandler: videoH,
	}
}

func (a *API) Register(mux *http.ServeMux) {
	// Register auth mux
	authMux := http.NewServeMux()
	a.authHandler.Register(authMux)
	mux.Handle("/auth/", http.StripPrefix("/auth", authMux))

	// Register video mux with auth middleware
	videoMux := http.NewServeMux()
	a.videoHandler.Register(videoMux)
	authMiddleware := middleware.RequireAuth(a.config.JWTSecret)
	mux.Handle("/video/", http.StripPrefix("/video", authMiddleware(videoMux)))
}
