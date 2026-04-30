package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/quiverscore/backend-go/internal/middleware"
	"github.com/quiverscore/backend-go/internal/repository"
)

// ── CreateEvent ───────────────────────────────────────────────────────

func TestCreateEvent_Success(t *testing.T) {
	userID := uuid.New().String()
	clubID := uuid.New().String()
	templateID := uuid.New().String()
	eventDate := time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339)
	mock := &mockClubRepo{
		createEventFn: func(_ context.Context, id, cID, uID, name string, desc *string, tplID string, ed time.Time, loc *string) (*repository.EventOut, error) {
			return &repository.EventOut{
				ID:           id,
				ClubID:       cID,
				Name:         name,
				TemplateID:   tplID,
				EventDate:    ed,
				CreatedBy:    uID,
				Participants: []repository.EventParticipantOut{},
				CreatedAt:    time.Now().UTC(),
			}, nil
		},
	}
	h := clubsHandler(mock)

	req := clubsAuthedBody(http.MethodPost, "/events", userID, map[string]any{
		"name":        "Weekend Shoot",
		"template_id": templateID,
		"event_date":  eventDate,
	})
	req = withChiURLParam(req, "clubID", clubID)

	rr := httptest.NewRecorder()
	h.CreateEvent(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	var result repository.EventOut
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if result.Name != "Weekend Shoot" {
		t.Errorf("expected 'Weekend Shoot', got '%s'", result.Name)
	}
}

// ── ListEvents ───────────────────────────────────────────────────────

func TestListEvents_Success(t *testing.T) {
	userID := uuid.New().String()
	clubID := uuid.New().String()
	mock := &mockClubRepo{
		listEventsFn: func(_ context.Context, cID, uID string) ([]repository.EventOut, error) {
			return []repository.EventOut{
				{
					ID:           uuid.New().String(),
					ClubID:       cID,
					Name:         "Weekly Shoot",
					TemplateID:   uuid.New().String(),
					EventDate:    time.Now().Add(24 * time.Hour).UTC(),
					CreatedBy:    uID,
					Participants: []repository.EventParticipantOut{},
					CreatedAt:    time.Now().UTC(),
				},
				{
					ID:           uuid.New().String(),
					ClubID:       cID,
					Name:         "Monthly Competition",
					TemplateID:   uuid.New().String(),
					EventDate:    time.Now().Add(72 * time.Hour).UTC(),
					CreatedBy:    uID,
					Participants: []repository.EventParticipantOut{},
					CreatedAt:    time.Now().UTC(),
				},
			}, nil
		},
	}
	h := clubsHandler(mock)

	req := httptest.NewRequest(http.MethodGet, "/events", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	req = req.WithContext(ctx)
	req = withChiURLParam(req, "clubID", clubID)

	rr := httptest.NewRecorder()
	h.ListEvents(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var result []repository.EventOut
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 events, got %d", len(result))
	}
}

func TestListEvents_NotFound(t *testing.T) {
	mock := &mockClubRepo{
		listEventsFn: func(_ context.Context, _, _ string) ([]repository.EventOut, error) {
			return nil, errors.New("not found")
		},
	}
	h := clubsHandler(mock)

	req := httptest.NewRequest(http.MethodGet, "/events", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, uuid.New().String())
	req = req.WithContext(ctx)
	req = withChiURLParam(req, "clubID", uuid.New().String())

	rr := httptest.NewRecorder()
	h.ListEvents(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

// ── GetEvent ─────────────────────────────────────────────────────────

func TestGetEvent_Success(t *testing.T) {
	userID := uuid.New().String()
	clubID := uuid.New().String()
	eventID := uuid.New().String()
	mock := &mockClubRepo{
		getEventFn: func(_ context.Context, cID, eID, uID string) (*repository.EventOut, error) {
			return &repository.EventOut{
				ID:           eID,
				ClubID:       cID,
				Name:         "Saturday Shoot",
				TemplateID:   uuid.New().String(),
				EventDate:    time.Now().Add(24 * time.Hour).UTC(),
				CreatedBy:    uID,
				Participants: []repository.EventParticipantOut{},
				CreatedAt:    time.Now().UTC(),
			}, nil
		},
	}
	h := clubsHandler(mock)

	req := httptest.NewRequest(http.MethodGet, "/events/"+eventID, nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	req = req.WithContext(ctx)
	req = clubsChiParams(req, map[string]string{"clubID": clubID, "eventID": eventID})

	rr := httptest.NewRecorder()
	h.GetEvent(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var result repository.EventOut
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if result.Name != "Saturday Shoot" {
		t.Errorf("expected 'Saturday Shoot', got '%s'", result.Name)
	}
	if result.ID != eventID {
		t.Errorf("expected event ID %s, got %s", eventID, result.ID)
	}
}

func TestGetEvent_NotFound(t *testing.T) {
	mock := &mockClubRepo{
		getEventFn: func(_ context.Context, _, _, _ string) (*repository.EventOut, error) {
			return nil, errors.New("not found")
		},
	}
	h := clubsHandler(mock)

	req := httptest.NewRequest(http.MethodGet, "/events/bad-id", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, uuid.New().String())
	req = req.WithContext(ctx)
	req = clubsChiParams(req, map[string]string{"clubID": uuid.New().String(), "eventID": "bad-id"})

	rr := httptest.NewRecorder()
	h.GetEvent(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

// ── UpdateEvent ──────────────────────────────────────────────────────

func TestUpdateEvent_Success(t *testing.T) {
	userID := uuid.New().String()
	clubID := uuid.New().String()
	eventID := uuid.New().String()
	mock := &mockClubRepo{
		updateEventFn: func(_ context.Context, cID, eID, uID string, name *string, description *string, descriptionSet bool, eventDate *time.Time, location *string, locationSet bool) (*repository.EventOut, error) {
			n := "Updated Shoot"
			if name != nil {
				n = *name
			}
			return &repository.EventOut{
				ID:           eID,
				ClubID:       cID,
				Name:         n,
				TemplateID:   uuid.New().String(),
				EventDate:    time.Now().Add(24 * time.Hour).UTC(),
				CreatedBy:    uID,
				Participants: []repository.EventParticipantOut{},
				CreatedAt:    time.Now().UTC(),
			}, nil
		},
	}
	h := clubsHandler(mock)

	req := clubsAuthedBody(http.MethodPatch, "/events/"+eventID, userID, map[string]any{
		"name": "Updated Shoot",
	})
	req = clubsChiParams(req, map[string]string{"clubID": clubID, "eventID": eventID})

	rr := httptest.NewRecorder()
	h.UpdateEvent(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var result repository.EventOut
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if result.Name != "Updated Shoot" {
		t.Errorf("expected 'Updated Shoot', got '%s'", result.Name)
	}
}

func TestUpdateEvent_NotFound(t *testing.T) {
	mock := &mockClubRepo{
		updateEventFn: func(_ context.Context, _, _, _ string, _ *string, _ *string, _ bool, _ *time.Time, _ *string, _ bool) (*repository.EventOut, error) {
			return nil, errors.New("not found")
		},
	}
	h := clubsHandler(mock)

	req := clubsAuthedBody(http.MethodPatch, "/events/bad-id", uuid.New().String(), map[string]any{
		"name": "Nope",
	})
	req = clubsChiParams(req, map[string]string{"clubID": uuid.New().String(), "eventID": "bad-id"})

	rr := httptest.NewRecorder()
	h.UpdateEvent(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

// ── DeleteEvent ──────────────────────────────────────────────────────

func TestDeleteEvent_Success(t *testing.T) {
	userID := uuid.New().String()
	clubID := uuid.New().String()
	eventID := uuid.New().String()
	mock := &mockClubRepo{
		deleteEventFn: func(_ context.Context, _, _, _ string) error {
			return nil
		},
	}
	h := clubsHandler(mock)

	req := httptest.NewRequest(http.MethodDelete, "/events/"+eventID, nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	req = req.WithContext(ctx)
	req = clubsChiParams(req, map[string]string{"clubID": clubID, "eventID": eventID})

	rr := httptest.NewRecorder()
	h.DeleteEvent(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestDeleteEvent_NotFound(t *testing.T) {
	mock := &mockClubRepo{
		deleteEventFn: func(_ context.Context, _, _, _ string) error {
			return errors.New("not found")
		},
	}
	h := clubsHandler(mock)

	req := httptest.NewRequest(http.MethodDelete, "/events/bad-id", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, uuid.New().String())
	req = req.WithContext(ctx)
	req = clubsChiParams(req, map[string]string{"clubID": uuid.New().String(), "eventID": "bad-id"})

	rr := httptest.NewRecorder()
	h.DeleteEvent(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

// ── RSVPEvent ────────────────────────────────────────────────────────

func TestRSVPEvent_Success(t *testing.T) {
	userID := uuid.New().String()
	clubID := uuid.New().String()
	eventID := uuid.New().String()
	mock := &mockClubRepo{
		rsvpEventFn: func(_ context.Context, cID, eID, uID, status string) (*repository.EventOut, error) {
			return &repository.EventOut{
				ID:        eID,
				ClubID:    cID,
				Name:      "Saturday Shoot",
				EventDate: time.Now().Add(24 * time.Hour).UTC(),
				CreatedBy: uuid.New().String(),
				Participants: []repository.EventParticipantOut{
					{UserID: uID, Username: "archer1", Status: status},
				},
				CreatedAt: time.Now().UTC(),
			}, nil
		},
	}
	h := clubsHandler(mock)

	req := clubsAuthedBody(http.MethodPost, "/events/"+eventID+"/rsvp", userID, map[string]string{
		"status": "going",
	})
	req = clubsChiParams(req, map[string]string{"clubID": clubID, "eventID": eventID})

	rr := httptest.NewRecorder()
	h.RSVPEvent(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var result repository.EventOut
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(result.Participants) != 1 {
		t.Fatalf("expected 1 participant, got %d", len(result.Participants))
	}
	if result.Participants[0].Status != "going" {
		t.Errorf("expected status 'going', got '%s'", result.Participants[0].Status)
	}
}

func TestRSVPEvent_NotFound(t *testing.T) {
	mock := &mockClubRepo{
		rsvpEventFn: func(_ context.Context, _, _, _, _ string) (*repository.EventOut, error) {
			return nil, errors.New("not found")
		},
	}
	h := clubsHandler(mock)

	req := clubsAuthedBody(http.MethodPost, "/events/bad-id/rsvp", uuid.New().String(), map[string]string{
		"status": "going",
	})
	req = clubsChiParams(req, map[string]string{"clubID": uuid.New().String(), "eventID": "bad-id"})

	rr := httptest.NewRecorder()
	h.RSVPEvent(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

// ── CreateEvent validation ───────────────────────────────────────────

func TestCreateEvent_InvalidDate(t *testing.T) {
	h := clubsHandler(&mockClubRepo{})

	req := clubsAuthedBody(http.MethodPost, "/events", uuid.New().String(), map[string]any{
		"name":        "Bad Event",
		"template_id": uuid.New().String(),
		"event_date":  "not-a-date",
	})
	req = withChiURLParam(req, "clubID", uuid.New().String())

	rr := httptest.NewRecorder()
	h.CreateEvent(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestCreateEvent_Unauthorized(t *testing.T) {
	mock := &mockClubRepo{
		createEventFn: func(_ context.Context, _, _, _, _ string, _ *string, _ string, _ time.Time, _ *string) (*repository.EventOut, error) {
			return nil, errors.New("forbidden")
		},
	}
	h := clubsHandler(mock)

	eventDate := time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339)
	req := clubsAuthedBody(http.MethodPost, "/events", uuid.New().String(), map[string]any{
		"name":        "Test",
		"template_id": uuid.New().String(),
		"event_date":  eventDate,
	})
	req = withChiURLParam(req, "clubID", uuid.New().String())

	rr := httptest.NewRecorder()
	h.CreateEvent(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

// ── UpdateEvent validation ───────────────────────────────────────────

func TestUpdateEvent_InvalidDate(t *testing.T) {
	h := clubsHandler(&mockClubRepo{})

	req := clubsAuthedBody(http.MethodPatch, "/events/e1", uuid.New().String(), map[string]any{
		"event_date": "not-a-date",
	})
	req = clubsChiParams(req, map[string]string{"clubID": uuid.New().String(), "eventID": "e1"})

	rr := httptest.NewRecorder()
	h.UpdateEvent(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d: %s", rr.Code, rr.Body.String())
	}
}
