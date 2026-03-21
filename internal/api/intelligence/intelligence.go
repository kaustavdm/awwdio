package intelligence

import (
	"net/http"

	"github.com/kaustavdm/awwdio/config"
	"github.com/twilio/twilio-go"
)

type Handler struct {
	config       *config.Config
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
	mux.HandleFunc("GET /sentences/{roomName}", h.sentencesHandler)
	mux.HandleFunc("GET /results/{roomName}", h.resultsHandler)
}
