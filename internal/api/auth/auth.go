package auth

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/kaustavdm/awwdio/config"
	"github.com/twilio/twilio-go"
	verify "github.com/twilio/twilio-go/rest/verify/v2"
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
	mux.HandleFunc("POST /send-otp", h.sendOTPHandler)
	mux.HandleFunc("POST /verify-otp", h.verifyOTPHandler)
}

type SendOTPRequest struct {
	Channel string `json:"channel"` // "email" or "sms"
	To      string `json:"to"`      // Email address or phone number
}

type SendOTPResponse struct {
	Success bool `json:"success"`
}

type VerifyOTPRequest struct {
	Channel string `json:"channel"` // "email" or "sms"
	To      string `json:"to"`      // Email address or phone number
	OTP     string `json:"otp"`
}

type VerifyOTPResponse struct {
	Success bool   `json:"success"`
	Token   string `json:"token,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// sendOTPHandler sends an OTP via email or SMS using Twilio Verify
func (h *Handler) sendOTPHandler(w http.ResponseWriter, r *http.Request) {
	var req SendOTPRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("Failed to decode request", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid request body"})
		return
	}

	// Validate channel
	if req.Channel != "email" && req.Channel != "sms" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Channel must be 'email' or 'sms'"})
		return
	}

	// Validate contact information
	if req.To == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Contact information (to) is required"})
		return
	}

	// Create verification using Twilio Verify API
	params := &verify.CreateVerificationParams{}
	params.SetTo(req.To)
	params.SetChannel(req.Channel)

	resp, err := h.twilioClient.VerifyV2.CreateVerification(h.config.TwilioVerifyServiceSID, params)
	if err != nil {
		slog.Error("Failed to send OTP", "error", err, "channel", req.Channel)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to send OTP"})
		return
	}

	slog.Info("OTP sent", "channel", req.Channel, "to", req.To, "status", *resp.Status)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(SendOTPResponse{Success: true})
}

// verifyOTPHandler verifies the OTP code via email or SMS
func (h *Handler) verifyOTPHandler(w http.ResponseWriter, r *http.Request) {
	var req VerifyOTPRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("Failed to decode request", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid request body"})
		return
	}

	// Validate channel
	if req.Channel != "email" && req.Channel != "sms" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Channel must be 'email' or 'sms'"})
		return
	}

	// Validate required fields
	if req.To == "" || req.OTP == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Contact information and OTP are required"})
		return
	}

	// Verify the OTP using Twilio Verify API
	params := &verify.CreateVerificationCheckParams{}
	params.SetTo(req.To)
	params.SetCode(req.OTP)

	resp, err := h.twilioClient.VerifyV2.CreateVerificationCheck(h.config.TwilioVerifyServiceSID, params)
	if err != nil {
		slog.Error("Failed to verify OTP", "error", err, "channel", req.Channel)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to verify OTP"})
		return
	}

	if resp.Status == nil || *resp.Status != "approved" {
		slog.Warn("OTP verification failed", "channel", req.Channel, "to", req.To, "status", resp.Status)
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid OTP"})
		return
	}

	slog.Info("OTP verified", "channel", req.Channel, "to", req.To)

	// TODO: Generate and return a proper session token
	// For now, returning a placeholder token
	sessionToken := "session_" + req.To

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(VerifyOTPResponse{
		Success: true,
		Token:   sessionToken,
	})
}
