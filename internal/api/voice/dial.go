package voice

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/kaustavdm/awwdio/internal/api/middleware"
	"github.com/kaustavdm/awwdio/internal/api/response"
	twilioApi "github.com/twilio/twilio-go/rest/api/v2010"
)

type dialRequest struct {
	RoomName    string `json:"roomName"`
	PhoneNumber string `json:"phoneNumber"`
}

type dialResponse struct {
	Success bool   `json:"success"`
	CallSid string `json:"callSid"`
}

func (h *Handler) DialHandler(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	if user == nil {
		slog.Error("No user in context for dial request")
		response.Err(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	var req dialRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("Failed to decode dial request body", "error", err)
		response.Err(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.RoomName == "" {
		response.Err(w, http.StatusBadRequest, "roomName is required")
		return
	}

	if req.PhoneNumber == "" {
		response.Err(w, http.StatusBadRequest, "phoneNumber is required")
		return
	}

	// Validate that the room exists
	if _, err := h.twilioClient.VideoV1.FetchRoom(req.RoomName); err != nil {
		slog.Error("Room not found", "room", req.RoomName, "error", err)
		response.Err(w, http.StatusNotFound, "Room not found")
		return
	}

	// Use the phone number as the PSTN participant identity (not the
	// authenticated user). Each video room participant must have a unique
	// identity, and the web user is already connected with user.Subject.
	twimlURL := h.config.BaseURL + "/api/voice/twiml?room=" +
		url.QueryEscape(req.RoomName) + "&identity=" + url.QueryEscape(req.PhoneNumber)

	params := &twilioApi.CreateCallParams{}
	params.SetTo(req.PhoneNumber)
	params.SetFrom(h.config.TwilioPhoneNumber)
	params.SetUrl(twimlURL)

	call, err := h.twilioClient.Api.CreateCall(params)
	if err != nil {
		slog.Error("Failed to create outbound call", "error", err)
		response.Err(w, http.StatusInternalServerError, "Failed to initiate call")
		return
	}

	callSid := ""
	if call.Sid != nil {
		callSid = *call.Sid
	}

	slog.Info("Outbound call initiated", "callSid", callSid, "room", req.RoomName, "identity", user.Subject)

	response.JSON(w, http.StatusOK, dialResponse{
		Success: true,
		CallSid: callSid,
	})
}
