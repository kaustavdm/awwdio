package webhooks

import (
	"net/http"

	"github.com/kaustavdm/awwdio/config"
	"github.com/twilio/twilio-go"
)

// Handler holds the configuration and Twilio client for webhook handlers.
type Handler struct {
	config       *config.Config
	twilioClient *twilio.RestClient
}

// NewHandler creates a new webhooks Handler with a Twilio REST client.
func NewHandler(cfg *config.Config) *Handler {
	return &Handler{
		config: cfg,
		twilioClient: twilio.NewRestClientWithParams(twilio.ClientParams{
			Username:   cfg.TwilioApiKey,
			Password:   cfg.TwilioApiSecret,
			AccountSid: cfg.TwilioAccountSID,
		}),
	}
}

// Register registers the webhook routes on the given mux.
// Twilio signature validation is applied externally via middleware.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /room-status", h.roomStatusHandler)
	mux.HandleFunc("POST /composition", h.compositionHandler)
	mux.HandleFunc("POST /transcript", h.transcriptHandler)
}
