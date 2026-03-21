package webhooks

import (
	"log/slog"

	twilioSync "github.com/twilio/twilio-go/rest/sync/v1"
)

// updateSyncDocument updates (or skips) the Sync Document identified by roomName
// with the provided data map. The UniqueName of the document acts as the document
// identifier (docSid) for the update call.
//
// If TwilioSyncServiceSID is not configured the function logs a warning and returns
// without error so callers are never blocked by missing optional configuration.
func (h *Handler) updateSyncDocument(roomName string, data map[string]interface{}) {
	if h.config.TwilioSyncServiceSID == "" {
		slog.Warn("updateSyncDocument: TwilioSyncServiceSID not configured, skipping", "roomName", roomName)
		return
	}

	if roomName == "" {
		slog.Warn("updateSyncDocument: roomName is empty, skipping")
		return
	}

	params := &twilioSync.UpdateDocumentParams{}
	params.SetData(data)

	// When a document was created with a UniqueName, that UniqueName can be
	// passed as the document SID to update it directly.
	_, err := h.twilioClient.SyncV1.UpdateDocument(
		h.config.TwilioSyncServiceSID,
		roomName,
		params,
	)
	if err != nil {
		slog.Error("updateSyncDocument: failed to update Sync document",
			"roomName", roomName,
			"error", err,
		)
		return
	}

	slog.Info("Sync document updated", "roomName", roomName)
}
