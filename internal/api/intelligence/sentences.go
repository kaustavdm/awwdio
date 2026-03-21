package intelligence

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/kaustavdm/awwdio/internal/api/response"
)

// syncDocumentData is the shape of data stored in the Sync Document for a room.
type syncDocumentData struct {
	TranscriptSid string `json:"transcriptSid"`
}

// sentencesHandler proxies transcript sentences from Twilio Intelligence API.
// It looks up the transcriptSid for the room from the corresponding Sync Document,
// then fetches sentences from the Intelligence API and returns them as JSON.
func (h *Handler) sentencesHandler(w http.ResponseWriter, r *http.Request) {
	if h.config.TwilioSyncServiceSID == "" {
		response.Err(w, http.StatusServiceUnavailable, "Sync service not configured")
		return
	}

	if h.config.TwilioIntelligenceServiceSID == "" {
		response.Err(w, http.StatusServiceUnavailable, "Intelligence service not configured")
		return
	}

	roomName := r.PathValue("roomName")
	if roomName == "" {
		response.Err(w, http.StatusBadRequest, "Room name is required")
		return
	}

	// Fetch the Sync Document for this room to get the transcriptSid
	doc, err := h.twilioClient.SyncV1.FetchDocument(h.config.TwilioSyncServiceSID, roomName)
	if err != nil {
		slog.Error("Failed to fetch Sync document", "room", roomName, "error", err)
		response.Err(w, http.StatusNotFound, "Room data not found")
		return
	}

	// Parse document data to extract transcriptSid
	var docData syncDocumentData
	if doc.Data != nil {
		rawBytes, marshalErr := json.Marshal(doc.Data)
		if marshalErr != nil {
			slog.Error("Failed to marshal Sync document data", "room", roomName, "error", marshalErr)
			response.Err(w, http.StatusInternalServerError, "Failed to read room data")
			return
		}
		if unmarshalErr := json.Unmarshal(rawBytes, &docData); unmarshalErr != nil {
			slog.Error("Failed to unmarshal Sync document data", "room", roomName, "error", unmarshalErr)
			response.Err(w, http.StatusInternalServerError, "Failed to parse room data")
			return
		}
	}

	if docData.TranscriptSid == "" {
		response.Err(w, http.StatusNotFound, "Results not yet available")
		return
	}

	// Fetch sentences from Intelligence API
	sentences, err := h.twilioClient.IntelligenceV2.ListSentence(docData.TranscriptSid, nil)
	if err != nil {
		slog.Error("Failed to fetch transcript sentences", "transcriptSid", docData.TranscriptSid, "error", err)
		response.Err(w, http.StatusInternalServerError, "Failed to fetch transcript sentences")
		return
	}

	slog.Info("Fetched transcript sentences", "room", roomName, "transcriptSid", docData.TranscriptSid, "count", len(sentences))

	response.JSON(w, http.StatusOK, sentences)
}
