package webhooks

import (
	"fmt"
	"log/slog"
	"net/http"

	twilioIntelligence "github.com/twilio/twilio-go/rest/intelligence/v2"
)

// compositionHandler receives composition status callbacks from Twilio Video.
// When a composition becomes available it initiates a transcript via Twilio
// Intelligence. When a composition fails it records the failure in Sync.
func (h *Handler) compositionHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		slog.Error("composition: failed to parse form", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	event := r.FormValue("StatusCallbackEvent")
	compositionSid := r.FormValue("CompositionSid")
	roomSid := r.FormValue("RoomSid")
	roomName := r.FormValue("RoomName")

	slog.Info("composition webhook received",
		"event", event,
		"compositionSid", compositionSid,
		"roomSid", roomSid,
		"roomName", roomName,
	)

	switch event {
	case "composition-available":
		h.handleCompositionAvailable(w, r, compositionSid, roomSid, roomName)
	case "composition-failed":
		h.handleCompositionFailed(w, r, compositionSid, roomSid, roomName)
	default:
		w.WriteHeader(http.StatusOK)
	}
}

func (h *Handler) handleCompositionAvailable(
	w http.ResponseWriter,
	_ *http.Request,
	compositionSid, roomSid, roomName string,
) {
	mediaURL := fmt.Sprintf("https://video.twilio.com/v1/Compositions/%s/Media", compositionSid)

	slog.Info("composition available", "compositionSid", compositionSid, "mediaURL", mediaURL)

	if h.config.TwilioIntelligenceServiceSID == "" {
		slog.Warn("composition: TwilioIntelligenceServiceSID not configured, skipping transcript creation")
		// Still update Sync to reflect that composing is done without transcription.
		h.updateSyncDocument(roomName, map[string]interface{}{
			"roomSid":        roomSid,
			"roomName":       roomName,
			"compositionSid": compositionSid,
			"mediaURL":       mediaURL,
			"status":         "composed",
		})
		w.WriteHeader(http.StatusOK)
		return
	}

	// Create transcript via Twilio Intelligence v2.
	// The Channel parameter describes the media source for transcription.
	transcriptParams := &twilioIntelligence.CreateTranscriptParams{}
	transcriptParams.SetServiceSid(h.config.TwilioIntelligenceServiceSID)
	transcriptParams.SetChannel(map[string]interface{}{
		"media_properties": map[string]interface{}{
			"source_sid":  compositionSid,
			"media_url":   mediaURL,
		},
	})

	transcript, err := h.twilioClient.IntelligenceV2.CreateTranscript(transcriptParams)
	if err != nil {
		slog.Error("composition: failed to create transcript", "compositionSid", compositionSid, "error", err)
		// Record composed state without transcript.
		h.updateSyncDocument(roomName, map[string]interface{}{
			"roomSid":        roomSid,
			"roomName":       roomName,
			"compositionSid": compositionSid,
			"mediaURL":       mediaURL,
			"status":         "composed",
		})
		w.WriteHeader(http.StatusOK)
		return
	}

	slog.Info("transcript created", "transcriptSid", transcript.Sid, "compositionSid", compositionSid)

	h.updateSyncDocument(roomName, map[string]interface{}{
		"roomSid":        roomSid,
		"roomName":       roomName,
		"compositionSid": compositionSid,
		"mediaURL":       mediaURL,
		"transcriptSid":  transcript.Sid,
		"status":         "transcribing",
	})

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) handleCompositionFailed(
	w http.ResponseWriter,
	_ *http.Request,
	compositionSid, roomSid, roomName string,
) {
	slog.Error("composition failed", "compositionSid", compositionSid, "roomSid", roomSid)

	h.updateSyncDocument(roomName, map[string]interface{}{
		"roomSid":        roomSid,
		"roomName":       roomName,
		"compositionSid": compositionSid,
		"status":         "failed",
	})

	w.WriteHeader(http.StatusOK)
}
