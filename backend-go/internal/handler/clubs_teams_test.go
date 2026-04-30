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

// ── CreateTeam ────────────────────────────────────────────────────────

func TestCreateTeam_Success(t *testing.T) {
	userID := uuid.New().String()
	clubID := uuid.New().String()
	leaderID := uuid.New().String()
	mock := &mockClubRepo{
		createTeamFn: func(_ context.Context, id, cID, uID, name string, desc *string, lID string) (*repository.TeamOut, error) {
			return &repository.TeamOut{
				ID:          id,
				ClubID:      cID,
				Name:        name,
				Description: desc,
				Leader: repository.TeamMemberOut{
					UserID:   lID,
					Username: "leader",
					JoinedAt: time.Now().UTC(),
				},
				MemberCount: 1,
				CreatedAt:   time.Now().UTC(),
			}, nil
		},
	}
	h := clubsHandler(mock)

	req := clubsAuthedBody(http.MethodPost, "/teams", userID, map[string]string{
		"name":      "Team Alpha",
		"leader_id": leaderID,
	})
	req = withChiURLParam(req, "clubID", clubID)

	rr := httptest.NewRecorder()
	h.CreateTeam(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	var result repository.TeamOut
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if result.Name != "Team Alpha" {
		t.Errorf("expected 'Team Alpha', got '%s'", result.Name)
	}
}

func TestCreateTeam_Unauthorized(t *testing.T) {
	mock := &mockClubRepo{
		createTeamFn: func(_ context.Context, _, _, _, _ string, _ *string, _ string) (*repository.TeamOut, error) {
			return nil, errors.New("forbidden")
		},
	}
	h := clubsHandler(mock)

	req := clubsAuthedBody(http.MethodPost, "/teams", uuid.New().String(), map[string]string{
		"name":      "Team Nope",
		"leader_id": uuid.New().String(),
	})
	req = withChiURLParam(req, "clubID", uuid.New().String())

	rr := httptest.NewRecorder()
	h.CreateTeam(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

// ── ListTeams ────────────────────────────────────────────────────────

func TestListTeams_Success(t *testing.T) {
	userID := uuid.New().String()
	clubID := uuid.New().String()
	mock := &mockClubRepo{
		listTeamsFn: func(_ context.Context, cID, uID string) ([]repository.TeamOut, error) {
			return []repository.TeamOut{
				{
					ID:     uuid.New().String(),
					ClubID: cID,
					Name:   "Team Alpha",
					Leader: repository.TeamMemberOut{
						UserID:   uuid.New().String(),
						Username: "leader1",
						JoinedAt: time.Now().UTC(),
					},
					MemberCount: 3,
					CreatedAt:   time.Now().UTC(),
				},
				{
					ID:     uuid.New().String(),
					ClubID: cID,
					Name:   "Team Beta",
					Leader: repository.TeamMemberOut{
						UserID:   uuid.New().String(),
						Username: "leader2",
						JoinedAt: time.Now().UTC(),
					},
					MemberCount: 2,
					CreatedAt:   time.Now().UTC(),
				},
			}, nil
		},
	}
	h := clubsHandler(mock)

	req := httptest.NewRequest(http.MethodGet, "/teams", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	req = req.WithContext(ctx)
	req = withChiURLParam(req, "clubID", clubID)

	rr := httptest.NewRecorder()
	h.ListTeams(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var result []repository.TeamOut
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 teams, got %d", len(result))
	}
}

func TestListTeams_NotFound(t *testing.T) {
	mock := &mockClubRepo{
		listTeamsFn: func(_ context.Context, _, _ string) ([]repository.TeamOut, error) {
			return nil, errors.New("not found")
		},
	}
	h := clubsHandler(mock)

	req := httptest.NewRequest(http.MethodGet, "/teams", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, uuid.New().String())
	req = req.WithContext(ctx)
	req = withChiURLParam(req, "clubID", uuid.New().String())

	rr := httptest.NewRecorder()
	h.ListTeams(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

// ── GetTeamDetail ────────────────────────────────────────────────────

func TestGetTeamDetail_Success(t *testing.T) {
	userID := uuid.New().String()
	clubID := uuid.New().String()
	teamID := uuid.New().String()
	mock := &mockClubRepo{
		getTeamDetailFn: func(_ context.Context, cID, tID, uID string) (*repository.TeamDetailOut, error) {
			return &repository.TeamDetailOut{
				ID:     tID,
				ClubID: cID,
				Name:   "Team Alpha",
				Leader: repository.TeamMemberOut{
					UserID:   uuid.New().String(),
					Username: "leader",
					JoinedAt: time.Now().UTC(),
				},
				MemberCount: 3,
				CreatedAt:   time.Now().UTC(),
				Members: []repository.TeamMemberOut{
					{UserID: uuid.New().String(), Username: "member1", JoinedAt: time.Now().UTC()},
					{UserID: uuid.New().String(), Username: "member2", JoinedAt: time.Now().UTC()},
				},
			}, nil
		},
	}
	h := clubsHandler(mock)

	req := httptest.NewRequest(http.MethodGet, "/teams/"+teamID, nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	req = req.WithContext(ctx)
	req = clubsChiParams(req, map[string]string{"clubID": clubID, "teamID": teamID})

	rr := httptest.NewRecorder()
	h.GetTeamDetail(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var result repository.TeamDetailOut
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if result.Name != "Team Alpha" {
		t.Errorf("expected 'Team Alpha', got '%s'", result.Name)
	}
	if len(result.Members) != 2 {
		t.Errorf("expected 2 members, got %d", len(result.Members))
	}
}

func TestGetTeamDetail_NotFound(t *testing.T) {
	mock := &mockClubRepo{
		getTeamDetailFn: func(_ context.Context, _, _, _ string) (*repository.TeamDetailOut, error) {
			return nil, errors.New("not found")
		},
	}
	h := clubsHandler(mock)

	req := httptest.NewRequest(http.MethodGet, "/teams/bad-id", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, uuid.New().String())
	req = req.WithContext(ctx)
	req = clubsChiParams(req, map[string]string{"clubID": uuid.New().String(), "teamID": "bad-id"})

	rr := httptest.NewRecorder()
	h.GetTeamDetail(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

// ── UpdateTeam ───────────────────────────────────────────────────────

func TestUpdateTeam_Success(t *testing.T) {
	userID := uuid.New().String()
	clubID := uuid.New().String()
	teamID := uuid.New().String()
	mock := &mockClubRepo{
		updateTeamFn: func(_ context.Context, cID, tID, uID string, name, description *string, leaderID *string) (*repository.TeamOut, error) {
			n := "Updated Team"
			if name != nil {
				n = *name
			}
			return &repository.TeamOut{
				ID:     tID,
				ClubID: cID,
				Name:   n,
				Leader: repository.TeamMemberOut{
					UserID:   uuid.New().String(),
					Username: "leader",
					JoinedAt: time.Now().UTC(),
				},
				MemberCount: 2,
				CreatedAt:   time.Now().UTC(),
			}, nil
		},
	}
	h := clubsHandler(mock)

	req := clubsAuthedBody(http.MethodPatch, "/teams/"+teamID, userID, map[string]any{
		"name": "Updated Team",
	})
	req = clubsChiParams(req, map[string]string{"clubID": clubID, "teamID": teamID})

	rr := httptest.NewRecorder()
	h.UpdateTeam(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var result repository.TeamOut
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if result.Name != "Updated Team" {
		t.Errorf("expected 'Updated Team', got '%s'", result.Name)
	}
}

func TestUpdateTeam_NotFound(t *testing.T) {
	mock := &mockClubRepo{
		updateTeamFn: func(_ context.Context, _, _, _ string, _, _ *string, _ *string) (*repository.TeamOut, error) {
			return nil, errors.New("not found")
		},
	}
	h := clubsHandler(mock)

	req := clubsAuthedBody(http.MethodPatch, "/teams/bad-id", uuid.New().String(), map[string]any{
		"name": "Nope",
	})
	req = clubsChiParams(req, map[string]string{"clubID": uuid.New().String(), "teamID": "bad-id"})

	rr := httptest.NewRecorder()
	h.UpdateTeam(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

// ── DeleteTeam ───────────────────────────────────────────────────────

func TestDeleteTeam_Success(t *testing.T) {
	userID := uuid.New().String()
	clubID := uuid.New().String()
	teamID := uuid.New().String()
	mock := &mockClubRepo{
		deleteTeamFn: func(_ context.Context, _, _, _ string) error {
			return nil
		},
	}
	h := clubsHandler(mock)

	req := httptest.NewRequest(http.MethodDelete, "/teams/"+teamID, nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	req = req.WithContext(ctx)
	req = clubsChiParams(req, map[string]string{"clubID": clubID, "teamID": teamID})

	rr := httptest.NewRecorder()
	h.DeleteTeam(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestDeleteTeam_NotFound(t *testing.T) {
	mock := &mockClubRepo{
		deleteTeamFn: func(_ context.Context, _, _, _ string) error {
			return errors.New("not found")
		},
	}
	h := clubsHandler(mock)

	req := httptest.NewRequest(http.MethodDelete, "/teams/bad-id", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, uuid.New().String())
	req = req.WithContext(ctx)
	req = clubsChiParams(req, map[string]string{"clubID": uuid.New().String(), "teamID": "bad-id"})

	rr := httptest.NewRecorder()
	h.DeleteTeam(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

// ── AddTeamMember ────────────────────────────────────────────────────

func TestAddTeamMember_Success(t *testing.T) {
	userID := uuid.New().String()
	clubID := uuid.New().String()
	teamID := uuid.New().String()
	memberUserID := uuid.New().String()
	mock := &mockClubRepo{
		addTeamMemberFn: func(_ context.Context, _, _, _, _ string) error {
			return nil
		},
	}
	h := clubsHandler(mock)

	req := httptest.NewRequest(http.MethodPost, "/teams/"+teamID+"/members/"+memberUserID, nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	req = req.WithContext(ctx)
	req = clubsChiParams(req, map[string]string{"clubID": clubID, "teamID": teamID, "memberUserID": memberUserID})

	rr := httptest.NewRecorder()
	h.AddTeamMember(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	var result map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if result["detail"] != "Member added to team" {
		t.Errorf("expected 'Member added to team', got '%s'", result["detail"])
	}
}

func TestAddTeamMember_AlreadyMember(t *testing.T) {
	mock := &mockClubRepo{
		addTeamMemberFn: func(_ context.Context, _, _, _, _ string) error {
			return repository.ErrAlreadyMember
		},
	}
	h := clubsHandler(mock)

	req := httptest.NewRequest(http.MethodPost, "/teams/t1/members/u1", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, uuid.New().String())
	req = req.WithContext(ctx)
	req = clubsChiParams(req, map[string]string{"clubID": "c1", "teamID": "t1", "memberUserID": "u1"})

	rr := httptest.NewRecorder()
	h.AddTeamMember(rr, req)

	if rr.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rr.Code)
	}
}

func TestAddTeamMember_Unauthorized(t *testing.T) {
	mock := &mockClubRepo{
		addTeamMemberFn: func(_ context.Context, _, _, _, _ string) error {
			return errors.New("forbidden")
		},
	}
	h := clubsHandler(mock)

	req := httptest.NewRequest(http.MethodPost, "/teams/t1/members/u1", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, uuid.New().String())
	req = req.WithContext(ctx)
	req = clubsChiParams(req, map[string]string{"clubID": "c1", "teamID": "t1", "memberUserID": "u1"})

	rr := httptest.NewRecorder()
	h.AddTeamMember(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

// ── RemoveTeamMember ─────────────────────────────────────────────────

func TestRemoveTeamMember_Success(t *testing.T) {
	userID := uuid.New().String()
	clubID := uuid.New().String()
	teamID := uuid.New().String()
	memberUserID := uuid.New().String()
	mock := &mockClubRepo{
		removeTeamMemberFn: func(_ context.Context, _, _, _, _ string) error {
			return nil
		},
	}
	h := clubsHandler(mock)

	req := httptest.NewRequest(http.MethodDelete, "/teams/"+teamID+"/members/"+memberUserID, nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	req = req.WithContext(ctx)
	req = clubsChiParams(req, map[string]string{"clubID": clubID, "teamID": teamID, "memberUserID": memberUserID})

	rr := httptest.NewRecorder()
	h.RemoveTeamMember(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestRemoveTeamMember_NotFound(t *testing.T) {
	mock := &mockClubRepo{
		removeTeamMemberFn: func(_ context.Context, _, _, _, _ string) error {
			return errors.New("not found")
		},
	}
	h := clubsHandler(mock)

	req := httptest.NewRequest(http.MethodDelete, "/teams/t1/members/u1", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, uuid.New().String())
	req = req.WithContext(ctx)
	req = clubsChiParams(req, map[string]string{"clubID": "c1", "teamID": "t1", "memberUserID": "u1"})

	rr := httptest.NewRecorder()
	h.RemoveTeamMember(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}
