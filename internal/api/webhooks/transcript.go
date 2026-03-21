package webhooks

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
)

// transcriptWebhookPayload represents the JSON body sent by Twilio Intelligence
// for transcript status callbacks.
type transcriptWebhookPayload struct {
	TranscriptSid string `json:"TranscriptSid"`
	ServiceSid    string `json:"ServiceSid"`
	Status        string `json:"Status"`
	RoomName      string `json:"RoomName"`
}

// transcriptHandler receives transcript status callbacks from Twilio Intelligence.
// It updates the Sync Document for the room with the final transcript state.
func (h *Handler) transcriptHandler(w http.ResponseWriter, r *http.Request) {
	var transcriptSid, status, roomName string

	contentType := r.Header.Get("Content-Type")

	if strings.Contains(contentType, "application/json") {
		// Parse JSON body.
		var payload transcriptWebhookPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			slog.Error("transcript: failed to decode JSON body", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		transcriptSid = payload.TranscriptSid
		status = payload.Status
		roomName = payload.RoomName
	} else {
		// Parse form body (application/x-www-form-urlencoded).
		if err := r.ParseForm(); err != nil {
			slog.Error("transcript: failed to parse form", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		transcriptSid = r.FormValue("TranscriptSid")
		status = r.FormValue("Status")
		roomName = r.FormValue("RoomName")
	}

	slog.Info("transcript webhook received",
		"transcriptSid", transcriptSid,
		"status", status,
		"roomName", roomName,
	)

	switch status {
	case "completed":
		slog.Info("transcript completed", "transcriptSid", transcriptSid, "roomName", roomName)
		h.updateSyncDocument(roomName, map[string]interface{}{
			"roomName":      roomName,
			"transcriptSid": transcriptSid,
			"status":        "ready",
		})
	case "failed":
		slog.Error("transcript failed", "transcriptSid", transcriptSid, "roomName", roomName)
		h.updateSyncDocument(roomName, map[string]interface{}{
			"roomName":      roomName,
			"transcriptSid": transcriptSid,
			"status":        "failed",
		})
	default:
		slog.Info("transcript status update (no action)", "transcriptSid", transcriptSid, "status", status)
	}

	w.WriteHeader(http.StatusOK)
}
