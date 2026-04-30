package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/quiverscore/backend-go/internal/middleware"
)

// ── Events ────────────────────────────────────────────────────────────

type eventCreate struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
	TemplateID  string  `json:"template_id"`
	EventDate   string  `json:"event_date"`
	Location    *string `json:"location"`
}

func (h *ClubsHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	clubID := chi.URLParam(r, "clubID")
	userID := middleware.GetUserID(r.Context())

	var req eventCreate
	if !Decode(w, r, &req) {
		return
	}

	eventDate, err := time.Parse(time.RFC3339, req.EventDate)
	if err != nil {
		ValidationError(w, "Invalid event_date format")
		return
	}

	id := uuid.New().String()
	event, err := h.Clubs.CreateEvent(r.Context(), id, clubID, userID, req.Name, req.Description, req.TemplateID, eventDate, req.Location)
	if err != nil {
		Error(w, http.StatusUnauthorized, "Only owner or admin can create events")
		return
	}

	JSON(w, http.StatusCreated, event)
}

func (h *ClubsHandler) ListEvents(w http.ResponseWriter, r *http.Request) {
	clubID := chi.URLParam(r, "clubID")
	userID := middleware.GetUserID(r.Context())

	events, err := h.Clubs.ListEvents(r.Context(), clubID, userID)
	if err != nil {
		Error(w, http.StatusNotFound, "Club not found")
		return
	}

	JSON(w, http.StatusOK, events)
}

func (h *ClubsHandler) GetEvent(w http.ResponseWriter, r *http.Request) {
	clubID := chi.URLParam(r, "clubID")
	eventID := chi.URLParam(r, "eventID")
	userID := middleware.GetUserID(r.Context())

	event, err := h.Clubs.GetEvent(r.Context(), clubID, eventID, userID)
	if err != nil {
		Error(w, http.StatusNotFound, "Event not found")
		return
	}

	JSON(w, http.StatusOK, event)
}

type eventUpdate struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	EventDate   *string `json:"event_date"`
	Location    *string `json:"location"`
}

func (h *ClubsHandler) UpdateEvent(w http.ResponseWriter, r *http.Request) {
	clubID := chi.URLParam(r, "clubID")
	eventID := chi.URLParam(r, "eventID")
	userID := middleware.GetUserID(r.Context())

	var req eventUpdate
	if !Decode(w, r, &req) {
		return
	}

	var eventDate *time.Time
	if req.EventDate != nil {
		t, err := time.Parse(time.RFC3339, *req.EventDate)
		if err != nil {
			ValidationError(w, "Invalid event_date format")
			return
		}
		eventDate = &t
	}

	event, err := h.Clubs.UpdateEvent(r.Context(), clubID, eventID, userID,
		req.Name, req.Description, req.Description != nil,
		eventDate, req.Location, req.Location != nil,
	)
	if err != nil {
		Error(w, http.StatusNotFound, "Event not found")
		return
	}

	JSON(w, http.StatusOK, event)
}

func (h *ClubsHandler) DeleteEvent(w http.ResponseWriter, r *http.Request) {
	clubID := chi.URLParam(r, "clubID")
	eventID := chi.URLParam(r, "eventID")
	userID := middleware.GetUserID(r.Context())

	if err := h.Clubs.DeleteEvent(r.Context(), clubID, eventID, userID); err != nil {
		Error(w, http.StatusNotFound, "Event not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type eventRSVP struct {
	Status string `json:"status"`
}

func (h *ClubsHandler) RSVPEvent(w http.ResponseWriter, r *http.Request) {
	clubID := chi.URLParam(r, "clubID")
	eventID := chi.URLParam(r, "eventID")
	userID := middleware.GetUserID(r.Context())

	var req eventRSVP
	if !Decode(w, r, &req) {
		return
	}

	event, err := h.Clubs.RSVPEvent(r.Context(), clubID, eventID, userID, req.Status)
	if err != nil {
		Error(w, http.StatusNotFound, "Event not found")
		return
	}

	JSON(w, http.StatusOK, event)
}
