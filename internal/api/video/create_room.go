package video

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/kaustavdm/awwdio/internal/api/response"
	openapi "github.com/twilio/twilio-go/rest/video/v1"
)

type createRoomRequest struct {
	Name string `json:"name"`
}

type createRoomResponse struct {
	SID  string `json:"sid"`
	Name string `json:"name"`
}

// ensureRoom creates a group room if one with the given name does not already exist.
// This is called before token generation to prevent the JS SDK from creating
// an ad-hoc room with the account's default (possibly legacy) topology.
func (h *Handler) ensureRoom(name string) error {
	roomType := "group"
	params := &openapi.CreateRoomParams{
		UniqueName: &name,
		Type:       &roomType,
	}

	if h.config.BaseURL != "" {
		statusCallback := h.config.BaseURL + "/api/webhooks/room-status"
		params.StatusCallback = &statusCallback
	}

	_, err := h.twilioClient.VideoV1.CreateRoom(params)
	if err != nil {
		// Room already exists — that's fine
		if isRoomExistsError(err) {
			return nil
		}
		return err
	}
	slog.Info("Auto-created room", "name", name)
	return nil
}

// isRoomExistsError checks if the Twilio API error indicates the room already exists (HTTP 53113 / status 400).
func isRoomExistsError(err error) bool {
	// The twilio-go SDK returns errors whose message contains the status code or error code.
	// Room-already-exists returns error code 53113.
	return strings.Contains(err.Error(), "53113")
}

func (h *Handler) createRoom(w http.ResponseWriter, r *http.Request) {
	var req createRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("Failed to decode create room request", "error", err)
		response.Err(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" {
		response.Err(w, http.StatusBadRequest, "Room name is required")
		return
	}

	roomType := "group"
	recordParticipants := true

	params := &openapi.CreateRoomParams{
		UniqueName:                  &req.Name,
		Type:                        &roomType,
		RecordParticipantsOnConnect: &recordParticipants,
	}

	if h.config.BaseURL != "" {
		statusCallback := h.config.BaseURL + "/api/webhooks/room-status"
		params.StatusCallback = &statusCallback
	}

	room, err := h.twilioClient.VideoV1.CreateRoom(params)
	if err != nil {
		slog.Error("Failed to create room", "error", err, "name", req.Name)
		response.Err(w, http.StatusInternalServerError, "Failed to create room")
		return
	}

	slog.Info("Created room", "sid", *room.Sid, "name", req.Name)

	response.JSON(w, http.StatusCreated, createRoomResponse{
		SID:  *room.Sid,
		Name: req.Name,
	})
}
