package voice

import (
	"fmt"
	"log/slog"
	"net/http"
)

func (h *Handler) StatusHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		slog.Error("Failed to parse status webhook form data", "error", err)
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	callSid := r.FormValue("CallSid")
	callStatus := r.FormValue("CallStatus")

	slog.Info("Call status update received", "callSid", callSid, "callStatus", callStatus)

	w.Header().Set("Content-Type", "text/xml")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `<?xml version="1.0" encoding="UTF-8"?><Response/>`)
}
