package api

import (
	"net/http"

	"github.com/kaustavdm/awwdio/config"
	"github.com/kaustavdm/awwdio/internal/api/auth"
	"github.com/kaustavdm/awwdio/internal/api/intelligence"
	"github.com/kaustavdm/awwdio/internal/api/middleware"
	apiSync "github.com/kaustavdm/awwdio/internal/api/sync"
	"github.com/kaustavdm/awwdio/internal/api/video"
	"github.com/kaustavdm/awwdio/internal/api/voice"
	"github.com/kaustavdm/awwdio/internal/api/webhooks"
)

type API struct {
	config *config.Config

	// Load sub-APIs
	authHandler         *auth.Handler
	videoHandler        *video.Handler
	voiceHandler        *voice.Handler
	intelligenceHandler *intelligence.Handler
	syncHandler         *apiSync.Handler
	webhooksHandler     *webhooks.Handler
}

func New(cfg *config.Config) *API {
	authH := auth.NewHandler(cfg)
	videoH := video.NewHandler(cfg)
	voiceH := voice.NewHandler(cfg)
	intelligenceH := intelligence.NewHandler(cfg)
	syncH := apiSync.NewHandler(cfg)
	webhooksH := webhooks.NewHandler(cfg)
	return &API{
		config:              cfg,
		authHandler:         authH,
		videoHandler:        videoH,
		voiceHandler:        voiceH,
		intelligenceHandler: intelligenceH,
		syncHandler:         syncH,
		webhooksHandler:     webhooksH,
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

	// Register voice routes with per-route-group middleware:
	// POST /voice/dial   — JWT auth required
	// POST /voice/twiml  — Twilio signature validation required
	// POST /voice/status — Twilio signature validation required
	twilioMiddleware := middleware.ValidateTwilioSignature(a.config.TwilioAuthToken, a.config.BaseURL)

	voiceDialMux := http.NewServeMux()
	voiceDialMux.HandleFunc("POST /dial", a.voiceHandler.DialHandler)
	mux.Handle("/voice/dial", http.StripPrefix("/voice", authMiddleware(voiceDialMux)))

	voiceWebhookMux := http.NewServeMux()
	voiceWebhookMux.HandleFunc("POST /twiml", a.voiceHandler.TwimlHandler)
	voiceWebhookMux.HandleFunc("POST /status", a.voiceHandler.StatusHandler)
	mux.Handle("/voice/", http.StripPrefix("/voice", twilioMiddleware(voiceWebhookMux)))

	// Register intelligence mux with auth middleware
	intelligenceMux := http.NewServeMux()
	a.intelligenceHandler.Register(intelligenceMux)
	mux.Handle("/intelligence/", http.StripPrefix("/intelligence", authMiddleware(intelligenceMux)))

	// Register sync mux with auth middleware
	syncMux := http.NewServeMux()
	a.syncHandler.Register(syncMux)
	mux.Handle("/sync/", http.StripPrefix("/sync", authMiddleware(syncMux)))

	// Register webhooks mux with Twilio signature validation middleware.
	// All three webhook endpoints (room-status, composition, transcript) require
	// Twilio-signed requests — no JWT auth.
	webhooksMux := http.NewServeMux()
	a.webhooksHandler.Register(webhooksMux)
	mux.Handle("/webhooks/", http.StripPrefix("/webhooks", twilioMiddleware(webhooksMux)))
}
