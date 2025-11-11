package video

import (
	"encoding/json"
	"net/http"
)

func (h *Handler) getRoom(w http.ResponseWriter, r *http.Request) {
	// Extract room name from URL
	roomName := r.URL.Query().Get("roomName")
	if roomName == "" {
		http.Error(w, "Room name is required", http.StatusBadRequest)
		return
	}

	// Fetch room details using Twilio API
	room, err := h.twilioClient.VideoV1.FetchRoom(roomName)
	if err != nil {
		http.Error(w, "Failed to fetch room details", http.StatusInternalServerError)
		return
	}

	// Send JSON response with room details
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(room)
}
