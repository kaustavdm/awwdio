package video

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/kaustavdm/awwdio/config"
	"github.com/kaustavdm/awwdio/internal/api/middleware"
	"github.com/twilio/twilio-go/client/jwt"
)

// accessToken generates a Twilio access token with Video grant for a given user identity and room name.
func accessToken(c *config.Config, identity, roomName string) (string, error) {
	params := jwt.AccessTokenParams{
		AccountSid:    c.TwilioAccountSID,
		SigningKeySid: c.TwilioApiKey,
		Secret:        c.TwilioApiSecret,
		Identity:      identity,
	}
	token := jwt.CreateAccessToken(params)
	grant := jwt.VideoGrant{
		Room: roomName,
	}
	token.AddGrant(&grant)
	return token.ToJwt()
}

type TokenRequest struct {
	Room string `json:"room"`
}

type TokenResponse struct {
	Token string `json:"token"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// tokenHandler handles the token generation
func (h *Handler) tokenHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get authenticated user from context
	user := middleware.GetUser(r)
	if user == nil {
		slog.Error("No user in context")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Authentication required"})
		return
	}

	// Parse request body for room name
	var req TokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("Failed to decode request", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid request body"})
		return
	}

	if req.Room == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Room name is required"})
		return
	}

	// Generate token using authenticated user identity
	token, err := accessToken(h.config, user.Subject, req.Room)
	if err != nil {
		slog.Error("Failed to generate access token", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to generate token"})
		return
	}

	slog.Info("Generated video token", "identity", user.Subject, "room", req.Room)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(TokenResponse{Token: token})
}
