package video

import (
	"encoding/json"
	"net/http"

	"github.com/kaustavdm/awwdio/config"
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

type TokenResponse struct {
	Token string `json:"token"`
}

// tokenHandler handles the token generation
func (h *Handler) tokenHandler(w http.ResponseWriter, r *http.Request) {
	token, err := accessToken(h.config, "user_1", "demo_room")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Send JSON response using TokenResponse
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := TokenResponse{
		Token: token,
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
