package webhooks

import (
	"log/slog"
	"net/http"

	twilioSync "github.com/twilio/twilio-go/rest/sync/v1"
	twilioVideo "github.com/twilio/twilio-go/rest/video/v1"
)

// roomStatusHandler receives room lifecycle webhooks from Twilio Video.
// It triggers audio composition when a room ends and optionally records
// the initial state in a Sync Document.
func (h *Handler) roomStatusHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		slog.Error("room-status: failed to parse form", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	event := r.FormValue("StatusCallbackEvent")
	roomName := r.FormValue("RoomName")
	roomSid := r.FormValue("RoomSid")

	slog.Info("room-status webhook received", "event", event, "roomName", roomName, "roomSid", roomSid)

	if event != "room-ended" {
		w.WriteHeader(http.StatusOK)
		return
	}

	slog.Info("room ended, creating audio composition", "roomName", roomName, "roomSid", roomSid)

	// Create audio composition for the ended room.
	compositionParams := &twilioVideo.CreateCompositionParams{}
	compositionParams.SetRoomSid(roomSid)
	compositionParams.SetAudioSources([]string{"*"})
	compositionParams.SetFormat("mp4")
	compositionParams.SetStatusCallback(h.config.BaseURL + "/api/webhooks/composition")

	composition, err := h.twilioClient.VideoV1.CreateComposition(compositionParams)
	if err != nil {
		slog.Error("room-status: failed to create composition", "roomSid", roomSid, "error", err)
		// Do not return an error status — Twilio does not retry on non-2xx for status callbacks.
		w.WriteHeader(http.StatusOK)
		return
	}

	slog.Info("composition created", "compositionSid", composition.Sid, "roomSid", roomSid)

	// Record initial composing state in Sync if configured.
	if h.config.TwilioSyncServiceSID == "" {
		slog.Warn("room-status: TwilioSyncServiceSID not configured, skipping Sync update")
		w.WriteHeader(http.StatusOK)
		return
	}

	syncParams := &twilioSync.CreateDocumentParams{}
	syncParams.SetUniqueName(roomName)
	syncParams.SetData(map[string]interface{}{
		"roomSid":  roomSid,
		"roomName": roomName,
		"status":   "composing",
	})

	_, err = h.twilioClient.SyncV1.CreateDocument(h.config.TwilioSyncServiceSID, syncParams)
	if err != nil {
		slog.Error("room-status: failed to create Sync document", "roomName", roomName, "error", err)
		// Non-fatal: continue and return 200.
	} else {
		slog.Info("Sync document created", "roomName", roomName)
	}

	w.WriteHeader(http.StatusOK)
}
