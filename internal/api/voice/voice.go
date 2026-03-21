package voice

import (
	"net/http"

	"github.com/kaustavdm/awwdio/config"
	"github.com/twilio/twilio-go"
)

type Handler struct {
	config *config.Config

	twilioClient *twilio.RestClient
}

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

// Register registers all voice routes on the provided mux.
// Callers are expected to wrap the mux (or individual sub-muxes) with
// the appropriate middleware (JWT auth for /dial, Twilio signature
// validation for /twiml and /status) before mounting it.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /dial", h.DialHandler)
	mux.HandleFunc("POST /twiml", h.TwimlHandler)
	mux.HandleFunc("POST /status", h.StatusHandler)
}
