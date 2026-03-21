package middleware

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"log/slog"
	"net/http"
	"sort"
	"strings"

	"github.com/kaustavdm/awwdio/internal/api/response"
)

// ValidateTwilioSignature returns middleware that validates the X-Twilio-Signature
// header using HMAC-SHA1. This authenticates that requests genuinely came from Twilio.
func ValidateTwilioSignature(authToken string, baseURL string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if authToken == "" {
				slog.Warn("Twilio Auth Token not configured, skipping signature validation")
				next.ServeHTTP(w, r)
				return
			}

			signature := r.Header.Get("X-Twilio-Signature")
			if signature == "" {
				slog.Debug("Missing X-Twilio-Signature header")
				response.Err(w, http.StatusForbidden, "Missing Twilio signature")
				return
			}

			// Parse form data to access POST params
			if err := r.ParseForm(); err != nil {
				slog.Error("Failed to parse form data", "error", err)
				response.Err(w, http.StatusBadRequest, "Invalid form data")
				return
			}

			// Build the full URL that Twilio used to generate the signature.
			// We use r.RequestURI (the original, un-stripped URI) combined with
			// the configured baseURL origin (since the Host header may differ
			// behind ngrok/reverse proxies). Twilio includes the full URL with
			// query string in its signature for POST requests.
			fullURL := baseURL + r.RequestURI

			// Sort POST params and concatenate key+value pairs
			keys := make([]string, 0, len(r.PostForm))
			for k := range r.PostForm {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			var data strings.Builder
			data.WriteString(fullURL)
			for _, k := range keys {
				data.WriteString(k)
				data.WriteString(r.PostForm.Get(k))
			}

			// Compute HMAC-SHA1
			mac := hmac.New(sha1.New, []byte(authToken))
			mac.Write([]byte(data.String()))
			expected := base64.StdEncoding.EncodeToString(mac.Sum(nil))

			if !hmac.Equal([]byte(signature), []byte(expected)) {
				slog.Debug("Invalid Twilio signature", "expected", expected, "got", signature)
				response.Err(w, http.StatusForbidden, "Invalid Twilio signature")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
