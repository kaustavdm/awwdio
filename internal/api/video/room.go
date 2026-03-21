package video

import (
	"log/slog"
	"net/http"

	"github.com/kaustavdm/awwdio/internal/api/response"
)

func (h *Handler) getRoom(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		response.Err(w, http.StatusBadRequest, "Room name is required")
		return
	}

	room, err := h.twilioClient.VideoV1.FetchRoom(name)
	if err != nil {
		slog.Error("Failed to fetch room", "error", err, "name", name)
		response.Err(w, http.StatusInternalServerError, "Failed to fetch room details")
		return
	}

	response.JSON(w, http.StatusOK, room)
}
