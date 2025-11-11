package video

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

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /token", h.tokenHandler)
	mux.HandleFunc("GET /room", h.getRoom)
}
