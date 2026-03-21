package sync

import (
	"log/slog"
	"net/http"

	"github.com/kaustavdm/awwdio/internal/api/middleware"
	"github.com/kaustavdm/awwdio/internal/api/response"
	"github.com/twilio/twilio-go/client/jwt"
)

type tokenResponse struct {
	Token string `json:"token"`
}

// tokenHandler generates a Twilio Sync access token for the authenticated user.
func (h *Handler) tokenHandler(w http.ResponseWriter, r *http.Request) {
	if h.config.TwilioSyncServiceSID == "" {
		response.Err(w, http.StatusServiceUnavailable, "Sync service not configured")
		return
	}

	user := middleware.GetUser(r)
	if user == nil {
		slog.Error("No user in context")
		response.Err(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	params := jwt.AccessTokenParams{
		AccountSid:    h.config.TwilioAccountSID,
		SigningKeySid: h.config.TwilioApiKey,
		Secret:        h.config.TwilioApiSecret,
		Identity:      user.Subject,
	}
	token := jwt.CreateAccessToken(params)
	grant := jwt.SyncGrant{
		ServiceSid: h.config.TwilioSyncServiceSID,
	}
	token.AddGrant(&grant)

	tokenStr, err := token.ToJwt()
	if err != nil {
		slog.Error("Failed to generate Sync access token", "identity", user.Subject, "error", err)
		response.Err(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	slog.Info("Generated Sync token", "identity", user.Subject)

	response.JSON(w, http.StatusOK, tokenResponse{Token: tokenStr})
}
