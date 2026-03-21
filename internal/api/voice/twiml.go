package voice

import (
	"fmt"
	"log/slog"
	"net/http"
)

func (h *Handler) TwimlHandler(w http.ResponseWriter, r *http.Request) {
	room := r.URL.Query().Get("room")
	identity := r.URL.Query().Get("identity")

	if room == "" {
		slog.Error("Missing room parameter in TwiML request")
		http.Error(w, "room parameter is required", http.StatusBadRequest)
		return
	}

	if identity == "" {
		slog.Error("Missing identity parameter in TwiML request")
		http.Error(w, "identity parameter is required", http.StatusBadRequest)
		return
	}

	slog.Info("Serving TwiML for voice call", "room", room, "identity", identity)

	w.Header().Set("Content-Type", "text/xml")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `<?xml version="1.0" encoding="UTF-8"?>
<Response>
  <Connect>
    <Room participantIdentity="%s">%s</Room>
  </Connect>
</Response>`, identity, room)
}
