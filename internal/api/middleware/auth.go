package middleware

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/kaustavdm/awwdio/internal/api/auth"
)

// ContextKey is the type for context keys
type ContextKey string

// UserContextKey is the key for storing user identity in context
const UserContextKey ContextKey = "user"

// UserClaims represents the authenticated user from JWT
type UserClaims struct {
	Subject string // User identifier (email or phone)
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// RequireAuth returns middleware that validates JWT tokens
func RequireAuth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			// Get Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				slog.Debug("Missing Authorization header")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(ErrorResponse{Error: "Authorization header required"})
				return
			}

			// Parse Bearer token
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				slog.Debug("Invalid Authorization header format")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid authorization format"})
				return
			}

			token := parts[1]

			// Validate JWT
			claims, err := auth.ValidateJWT(token, jwtSecret)
			if err != nil {
				slog.Debug("JWT validation failed", "error", err)
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid or expired token"})
				return
			}

			// Store user in context
			userClaims := &UserClaims{
				Subject: claims.Sub,
			}
			ctx := context.WithValue(r.Context(), UserContextKey, userClaims)

			slog.Debug("User authenticated", "subject", claims.Sub)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUser extracts user claims from request context
func GetUser(r *http.Request) *UserClaims {
	if user, ok := r.Context().Value(UserContextKey).(*UserClaims); ok {
		return user
	}
	return nil
}
