package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/quiverscore/backend-go/internal/config"
	"github.com/quiverscore/backend-go/internal/middleware"
	"github.com/quiverscore/backend-go/internal/repository"
)

// ── Mock ──────────────────────────────────────────────────────────────

type mockClubRepo struct {
	createFn                 func(ctx context.Context, id, name string, description *string, ownerID string) (*repository.ClubOut, error)
	listForUserFn            func(ctx context.Context, userID string) ([]repository.ClubOut, error)
	getDetailFn              func(ctx context.Context, clubID, userID string) (*repository.ClubDetailOut, error)
	updateFn                 func(ctx context.Context, clubID, userID string, name, description *string) (*repository.ClubOut, error)
	deleteFn                 func(ctx context.Context, clubID, userID string) error
	createInviteFn           func(ctx context.Context, id, clubID, userID, code string, maxUses *int, expiresAt *time.Time, frontendURL string) (*repository.InviteOut, error)
	listInvitesFn            func(ctx context.Context, clubID, userID string, frontendURL string) ([]repository.InviteOut, error)
	deactivateInviteFn       func(ctx context.Context, clubID, inviteID, userID string) error
	previewInviteFn          func(ctx context.Context, code, userID string) (*repository.ClubOut, error)
	joinViaInviteFn          func(ctx context.Context, code, userID string) (*repository.JoinResult, error)
	promoteMemberFn          func(ctx context.Context, clubID, targetUserID, userID string) error
	demoteMemberFn           func(ctx context.Context, clubID, targetUserID, userID string) error
	removeMemberFn           func(ctx context.Context, clubID, targetUserID, userID string) error
	leaderboardFn            func(ctx context.Context, clubID, userID string, templateID *string) ([]repository.LeaderboardOut, error)
	activityFn               func(ctx context.Context, clubID, userID string, limit, offset int) ([]repository.ActivityItem, error)
	createEventFn            func(ctx context.Context, id, clubID, userID, name string, description *string, templateID string, eventDate time.Time, location *string) (*repository.EventOut, error)
	listEventsFn             func(ctx context.Context, clubID, userID string) ([]repository.EventOut, error)
	getEventFn               func(ctx context.Context, clubID, eventID, userID string) (*repository.EventOut, error)
	updateEventFn            func(ctx context.Context, clubID, eventID, userID string, name *string, description *string, descriptionSet bool, eventDate *time.Time, location *string, locationSet bool) (*repository.EventOut, error)
	deleteEventFn            func(ctx context.Context, clubID, eventID, userID string) error
	rsvpEventFn              func(ctx context.Context, clubID, eventID, userID, status string) (*repository.EventOut, error)
	createTeamFn             func(ctx context.Context, id, clubID, userID, name string, description *string, leaderID string) (*repository.TeamOut, error)
	listTeamsFn              func(ctx context.Context, clubID, userID string) ([]repository.TeamOut, error)
	getTeamDetailFn          func(ctx context.Context, clubID, teamID, userID string) (*repository.TeamDetailOut, error)
	updateTeamFn             func(ctx context.Context, clubID, teamID, userID string, name, description *string, leaderID *string) (*repository.TeamOut, error)
	deleteTeamFn             func(ctx context.Context, clubID, teamID, userID string) error
	addTeamMemberFn          func(ctx context.Context, clubID, teamID, targetUserID, userID string) error
	removeTeamMemberFn       func(ctx context.Context, clubID, teamID, targetUserID, userID string) error
	listSharedRoundsFn       func(ctx context.Context, clubID, userID string) ([]repository.ClubSharedRoundOut, error)
	removeSharedRoundFn      func(ctx context.Context, clubID, roundID, userID string) error
	createTournamentFn       func(ctx context.Context, id, clubID, userID, name string, description *string, templateID string, maxParticipants *int, registrationDeadline, startDate, endDate time.Time) (*repository.TournamentOut, error)
	listTournamentsFn        func(ctx context.Context, clubID, userID string, status *string) ([]repository.TournamentOut, error)
	getTournamentDetailFn    func(ctx context.Context, clubID, tournamentID, userID string) (*repository.TournamentDetailOut, error)
	registerForTournamentFn  func(ctx context.Context, clubID, tournamentID, userID string) error
	startTournamentFn        func(ctx context.Context, clubID, tournamentID, userID string) (*repository.TournamentOut, error)
	tournamentLeaderboardFn  func(ctx context.Context, clubID, tournamentID, userID string) ([]repository.TournamentLeaderboardEntry, error)
	completeTournamentFn     func(ctx context.Context, clubID, tournamentID, userID string) (*repository.TournamentOut, error)
	withdrawFromTournamentFn func(ctx context.Context, clubID, tournamentID, userID string) error
	submitTournamentScoreFn  func(ctx context.Context, clubID, tournamentID, userID, sessionID string) (int, int, error)
}

func (m *mockClubRepo) Create(ctx context.Context, id, name string, description *string, ownerID string) (*repository.ClubOut, error) {
	return m.createFn(ctx, id, name, description, ownerID)
}
func (m *mockClubRepo) ListForUser(ctx context.Context, userID string) ([]repository.ClubOut, error) {
	return m.listForUserFn(ctx, userID)
}
func (m *mockClubRepo) GetDetail(ctx context.Context, clubID, userID string) (*repository.ClubDetailOut, error) {
	return m.getDetailFn(ctx, clubID, userID)
}
func (m *mockClubRepo) Update(ctx context.Context, clubID, userID string, name, description *string) (*repository.ClubOut, error) {
	return m.updateFn(ctx, clubID, userID, name, description)
}
func (m *mockClubRepo) Delete(ctx context.Context, clubID, userID string) error {
	return m.deleteFn(ctx, clubID, userID)
}
func (m *mockClubRepo) CreateInvite(ctx context.Context, id, clubID, userID, code string, maxUses *int, expiresAt *time.Time, frontendURL string) (*repository.InviteOut, error) {
	return m.createInviteFn(ctx, id, clubID, userID, code, maxUses, expiresAt, frontendURL)
}
func (m *mockClubRepo) ListInvites(ctx context.Context, clubID, userID string, frontendURL string) ([]repository.InviteOut, error) {
	return m.listInvitesFn(ctx, clubID, userID, frontendURL)
}
func (m *mockClubRepo) DeactivateInvite(ctx context.Context, clubID, inviteID, userID string) error {
	return m.deactivateInviteFn(ctx, clubID, inviteID, userID)
}
func (m *mockClubRepo) PreviewInvite(ctx context.Context, code, userID string) (*repository.ClubOut, error) {
	return m.previewInviteFn(ctx, code, userID)
}
func (m *mockClubRepo) JoinViaInvite(ctx context.Context, code, userID string) (*repository.JoinResult, error) {
	return m.joinViaInviteFn(ctx, code, userID)
}
func (m *mockClubRepo) PromoteMember(ctx context.Context, clubID, targetUserID, userID string) error {
	return m.promoteMemberFn(ctx, clubID, targetUserID, userID)
}
func (m *mockClubRepo) DemoteMember(ctx context.Context, clubID, targetUserID, userID string) error {
	return m.demoteMemberFn(ctx, clubID, targetUserID, userID)
}
func (m *mockClubRepo) RemoveMember(ctx context.Context, clubID, targetUserID, userID string) error {
	return m.removeMemberFn(ctx, clubID, targetUserID, userID)
}
func (m *mockClubRepo) Leaderboard(ctx context.Context, clubID, userID string, templateID *string) ([]repository.LeaderboardOut, error) {
	return m.leaderboardFn(ctx, clubID, userID, templateID)
}
func (m *mockClubRepo) Activity(ctx context.Context, clubID, userID string, limit, offset int) ([]repository.ActivityItem, error) {
	return m.activityFn(ctx, clubID, userID, limit, offset)
}
func (m *mockClubRepo) CreateEvent(ctx context.Context, id, clubID, userID, name string, description *string, templateID string, eventDate time.Time, location *string) (*repository.EventOut, error) {
	return m.createEventFn(ctx, id, clubID, userID, name, description, templateID, eventDate, location)
}
func (m *mockClubRepo) ListEvents(ctx context.Context, clubID, userID string) ([]repository.EventOut, error) {
	return m.listEventsFn(ctx, clubID, userID)
}
func (m *mockClubRepo) GetEvent(ctx context.Context, clubID, eventID, userID string) (*repository.EventOut, error) {
	return m.getEventFn(ctx, clubID, eventID, userID)
}
func (m *mockClubRepo) UpdateEvent(ctx context.Context, clubID, eventID, userID string, name *string, description *string, descriptionSet bool, eventDate *time.Time, location *string, locationSet bool) (*repository.EventOut, error) {
	return m.updateEventFn(ctx, clubID, eventID, userID, name, description, descriptionSet, eventDate, location, locationSet)
}
func (m *mockClubRepo) DeleteEvent(ctx context.Context, clubID, eventID, userID string) error {
	return m.deleteEventFn(ctx, clubID, eventID, userID)
}
func (m *mockClubRepo) RSVPEvent(ctx context.Context, clubID, eventID, userID, status string) (*repository.EventOut, error) {
	return m.rsvpEventFn(ctx, clubID, eventID, userID, status)
}
func (m *mockClubRepo) CreateTeam(ctx context.Context, id, clubID, userID, name string, description *string, leaderID string) (*repository.TeamOut, error) {
	return m.createTeamFn(ctx, id, clubID, userID, name, description, leaderID)
}
func (m *mockClubRepo) ListTeams(ctx context.Context, clubID, userID string) ([]repository.TeamOut, error) {
	return m.listTeamsFn(ctx, clubID, userID)
}
func (m *mockClubRepo) GetTeamDetail(ctx context.Context, clubID, teamID, userID string) (*repository.TeamDetailOut, error) {
	return m.getTeamDetailFn(ctx, clubID, teamID, userID)
}
func (m *mockClubRepo) UpdateTeam(ctx context.Context, clubID, teamID, userID string, name, description *string, leaderID *string) (*repository.TeamOut, error) {
	return m.updateTeamFn(ctx, clubID, teamID, userID, name, description, leaderID)
}
func (m *mockClubRepo) DeleteTeam(ctx context.Context, clubID, teamID, userID string) error {
	return m.deleteTeamFn(ctx, clubID, teamID, userID)
}
func (m *mockClubRepo) AddTeamMember(ctx context.Context, clubID, teamID, targetUserID, userID string) error {
	return m.addTeamMemberFn(ctx, clubID, teamID, targetUserID, userID)
}
func (m *mockClubRepo) RemoveTeamMember(ctx context.Context, clubID, teamID, targetUserID, userID string) error {
	return m.removeTeamMemberFn(ctx, clubID, teamID, targetUserID, userID)
}
func (m *mockClubRepo) ListSharedRounds(ctx context.Context, clubID, userID string) ([]repository.ClubSharedRoundOut, error) {
	return m.listSharedRoundsFn(ctx, clubID, userID)
}
func (m *mockClubRepo) RemoveSharedRound(ctx context.Context, clubID, roundID, userID string) error {
	return m.removeSharedRoundFn(ctx, clubID, roundID, userID)
}
func (m *mockClubRepo) CreateTournament(ctx context.Context, id, clubID, userID, name string, description *string, templateID string, maxParticipants *int, registrationDeadline, startDate, endDate time.Time) (*repository.TournamentOut, error) {
	return m.createTournamentFn(ctx, id, clubID, userID, name, description, templateID, maxParticipants, registrationDeadline, startDate, endDate)
}
func (m *mockClubRepo) ListTournaments(ctx context.Context, clubID, userID string, status *string) ([]repository.TournamentOut, error) {
	return m.listTournamentsFn(ctx, clubID, userID, status)
}
func (m *mockClubRepo) GetTournamentDetail(ctx context.Context, clubID, tournamentID, userID string) (*repository.TournamentDetailOut, error) {
	return m.getTournamentDetailFn(ctx, clubID, tournamentID, userID)
}
func (m *mockClubRepo) RegisterForTournament(ctx context.Context, clubID, tournamentID, userID string) error {
	return m.registerForTournamentFn(ctx, clubID, tournamentID, userID)
}
func (m *mockClubRepo) StartTournament(ctx context.Context, clubID, tournamentID, userID string) (*repository.TournamentOut, error) {
	return m.startTournamentFn(ctx, clubID, tournamentID, userID)
}
func (m *mockClubRepo) TournamentLeaderboard(ctx context.Context, clubID, tournamentID, userID string) ([]repository.TournamentLeaderboardEntry, error) {
	return m.tournamentLeaderboardFn(ctx, clubID, tournamentID, userID)
}
func (m *mockClubRepo) CompleteTournament(ctx context.Context, clubID, tournamentID, userID string) (*repository.TournamentOut, error) {
	return m.completeTournamentFn(ctx, clubID, tournamentID, userID)
}
func (m *mockClubRepo) WithdrawFromTournament(ctx context.Context, clubID, tournamentID, userID string) error {
	return m.withdrawFromTournamentFn(ctx, clubID, tournamentID, userID)
}
func (m *mockClubRepo) SubmitTournamentScore(ctx context.Context, clubID, tournamentID, userID, sessionID string) (int, int, error) {
	return m.submitTournamentScoreFn(ctx, clubID, tournamentID, userID, sessionID)
}

// ── Helpers ───────────────────────────────────────────────────────────

func clubsHandler(mock *mockClubRepo) *ClubsHandler {
	return &ClubsHandler{
		Clubs: mock,
		Cfg:   &config.Config{FrontendURL: "https://example.com"},
	}
}

// clubsAuthedBody builds an authed request with a JSON body.
func clubsAuthedBody(method, path, userID string, body any) *http.Request {
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(method, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	return req.WithContext(ctx)
}

// clubsChiParams sets multiple chi URL params on a request.
func clubsChiParams(r *http.Request, params map[string]string) *http.Request {
	rctx := chi.NewRouteContext()
	for k, v := range params {
		rctx.URLParams.Add(k, v)
	}
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

// ── CreateClub ────────────────────────────────────────────────────────

func TestCreateClub_Success(t *testing.T) {
	ownerID := uuid.New().String()
	mock := &mockClubRepo{
		createFn: func(_ context.Context, id, name string, description *string, oID string) (*repository.ClubOut, error) {
			role := "owner"
			return &repository.ClubOut{
				ID:          id,
				Name:        name,
				Description: description,
				OwnerID:     oID,
				MemberCount: 1,
				MyRole:      &role,
				CreatedAt:   time.Now().UTC(),
			}, nil
		},
	}
	h := clubsHandler(mock)

	req := clubsAuthedBody(http.MethodPost, "/", ownerID, map[string]string{"name": "Archery Club"})
	rr := httptest.NewRecorder()
	h.Create(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	var result repository.ClubOut
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if result.Name != "Archery Club" {
		t.Errorf("expected name 'Archery Club', got '%s'", result.Name)
	}
	if result.OwnerID != ownerID {
		t.Errorf("expected owner_id %s, got %s", ownerID, result.OwnerID)
	}
}

func TestCreateClub_MissingName(t *testing.T) {
	h := clubsHandler(&mockClubRepo{})

	req := clubsAuthedBody(http.MethodPost, "/", uuid.New().String(), map[string]string{})
	rr := httptest.NewRecorder()
	h.Create(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d: %s", rr.Code, rr.Body.String())
	}
}

// ── ListClubs ─────────────────────────────────────────────────────────

func TestListClubs_Success(t *testing.T) {
	userID := uuid.New().String()
	role := "member"
	mock := &mockClubRepo{
		listForUserFn: func(_ context.Context, _ string) ([]repository.ClubOut, error) {
			return []repository.ClubOut{
				{ID: uuid.New().String(), Name: "Club A", OwnerID: uuid.New().String(), MemberCount: 3, MyRole: &role, CreatedAt: time.Now().UTC()},
			}, nil
		},
	}
	h := clubsHandler(mock)

	rr := httptest.NewRecorder()
	h.List(rr, authedRequest(http.MethodGet, "/", userID))

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var result []repository.ClubOut
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 club, got %d", len(result))
	}
	if result[0].Name != "Club A" {
		t.Errorf("expected 'Club A', got '%s'", result[0].Name)
	}
}

// ── GetClub ───────────────────────────────────────────────────────────

func TestGetClub_Success(t *testing.T) {
	userID := uuid.New().String()
	clubID := uuid.New().String()
	role := "owner"
	mock := &mockClubRepo{
		getDetailFn: func(_ context.Context, cID, uID string) (*repository.ClubDetailOut, error) {
			return &repository.ClubDetailOut{
				ID:          cID,
				Name:        "My Club",
				OwnerID:     uID,
				MemberCount: 2,
				MyRole:      &role,
				CreatedAt:   time.Now().UTC(),
				Members:     []repository.ClubMemberOut{},
			}, nil
		},
	}
	h := clubsHandler(mock)

	req := authedRequest(http.MethodGet, "/", userID)
	req = withChiURLParam(req, "clubID", clubID)

	rr := httptest.NewRecorder()
	h.GetDetail(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var result repository.ClubDetailOut
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if result.Name != "My Club" {
		t.Errorf("expected 'My Club', got '%s'", result.Name)
	}
}

func TestGetClub_NotFound(t *testing.T) {
	mock := &mockClubRepo{
		getDetailFn: func(_ context.Context, _, _ string) (*repository.ClubDetailOut, error) {
			return nil, errors.New("not found")
		},
	}
	h := clubsHandler(mock)

	req := authedRequest(http.MethodGet, "/", uuid.New().String())
	req = withChiURLParam(req, "clubID", uuid.New().String())

	rr := httptest.NewRecorder()
	h.GetDetail(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

// ── UpdateClub ────────────────────────────────────────────────────────

func TestUpdateClub_Success(t *testing.T) {
	userID := uuid.New().String()
	clubID := uuid.New().String()
	mock := &mockClubRepo{
		updateFn: func(_ context.Context, cID, uID string, name, description *string) (*repository.ClubOut, error) {
			role := "owner"
			return &repository.ClubOut{
				ID:          cID,
				Name:        "Updated Club",
				OwnerID:     uID,
				MemberCount: 1,
				MyRole:      &role,
				CreatedAt:   time.Now().UTC(),
			}, nil
		},
	}
	h := clubsHandler(mock)

	req := clubsAuthedBody(http.MethodPatch, "/", userID, map[string]string{"name": "Updated Club"})
	req = withChiURLParam(req, "clubID", clubID)

	rr := httptest.NewRecorder()
	h.Update(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var result repository.ClubOut
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if result.Name != "Updated Club" {
		t.Errorf("expected 'Updated Club', got '%s'", result.Name)
	}
}

func TestUpdateClub_NotOwner(t *testing.T) {
	mock := &mockClubRepo{
		updateFn: func(_ context.Context, _, _ string, _, _ *string) (*repository.ClubOut, error) {
			return nil, errors.New("forbidden")
		},
	}
	h := clubsHandler(mock)

	req := clubsAuthedBody(http.MethodPatch, "/", uuid.New().String(), map[string]string{"name": "Nope"})
	req = withChiURLParam(req, "clubID", uuid.New().String())

	rr := httptest.NewRecorder()
	h.Update(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

// ── DeleteClub ────────────────────────────────────────────────────────

func TestDeleteClub_Success(t *testing.T) {
	userID := uuid.New().String()
	clubID := uuid.New().String()
	mock := &mockClubRepo{
		deleteFn: func(_ context.Context, _, _ string) error {
			return nil
		},
	}
	h := clubsHandler(mock)

	req := authedRequest(http.MethodDelete, "/", userID)
	req = withChiURLParam(req, "clubID", clubID)

	rr := httptest.NewRecorder()
	h.Delete(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestDeleteClub_NotOwner(t *testing.T) {
	mock := &mockClubRepo{
		deleteFn: func(_ context.Context, _, _ string) error {
			return errors.New("forbidden")
		},
	}
	h := clubsHandler(mock)

	req := authedRequest(http.MethodDelete, "/", uuid.New().String())
	req = withChiURLParam(req, "clubID", uuid.New().String())

	rr := httptest.NewRecorder()
	h.Delete(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

// ── CreateInvite ──────────────────────────────────────────────────────

func TestCreateInvite_Success(t *testing.T) {
	userID := uuid.New().String()
	clubID := uuid.New().String()
	mock := &mockClubRepo{
		createInviteFn: func(_ context.Context, id, cID, uID, code string, maxUses *int, expiresAt *time.Time, frontendURL string) (*repository.InviteOut, error) {
			return &repository.InviteOut{
				ID:        id,
				Code:      code,
				URL:       frontendURL + "/clubs/join/" + code,
				Active:    true,
				CreatedAt: time.Now().UTC(),
			}, nil
		},
	}
	h := clubsHandler(mock)

	req := clubsAuthedBody(http.MethodPost, "/invites", userID, map[string]any{})
	req = withChiURLParam(req, "clubID", clubID)

	rr := httptest.NewRecorder()
	h.CreateInvite(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	var result repository.InviteOut
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !result.Active {
		t.Error("expected invite to be active")
	}
}

// ── JoinClub ──────────────────────────────────────────────────────────

func TestJoinClub_Success(t *testing.T) {
	userID := uuid.New().String()
	clubID := uuid.New().String()
	mock := &mockClubRepo{
		joinViaInviteFn: func(_ context.Context, code, uID string) (*repository.JoinResult, error) {
			return &repository.JoinResult{
				ClubID:   clubID,
				ClubName: "Test Club",
				Role:     "member",
			}, nil
		},
	}
	h := clubsHandler(mock)

	req := authedRequest(http.MethodPost, "/join/abc123", userID)
	req = withChiURLParam(req, "code", "abc123")

	rr := httptest.NewRecorder()
	h.JoinViaInvite(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var result repository.JoinResult
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if result.ClubName != "Test Club" {
		t.Errorf("expected 'Test Club', got '%s'", result.ClubName)
	}
	if result.Role != "member" {
		t.Errorf("expected role 'member', got '%s'", result.Role)
	}
}

// ── ListMembers (via GetDetail) ───────────────────────────────────────

func TestListMembers_Success(t *testing.T) {
	userID := uuid.New().String()
	clubID := uuid.New().String()
	role := "owner"
	mock := &mockClubRepo{
		getDetailFn: func(_ context.Context, cID, uID string) (*repository.ClubDetailOut, error) {
			return &repository.ClubDetailOut{
				ID:          cID,
				Name:        "Club",
				OwnerID:     uID,
				MemberCount: 2,
				MyRole:      &role,
				CreatedAt:   time.Now().UTC(),
				Members: []repository.ClubMemberOut{
					{UserID: uID, Username: "owner", Role: "owner", JoinedAt: time.Now().UTC()},
					{UserID: uuid.New().String(), Username: "member1", Role: "member", JoinedAt: time.Now().UTC()},
				},
			}, nil
		},
	}
	h := clubsHandler(mock)

	req := authedRequest(http.MethodGet, "/", userID)
	req = withChiURLParam(req, "clubID", clubID)

	rr := httptest.NewRecorder()
	h.GetDetail(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var result repository.ClubDetailOut
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(result.Members) != 2 {
		t.Fatalf("expected 2 members, got %d", len(result.Members))
	}
}

// ── RemoveMember ──────────────────────────────────────────────────────

func TestRemoveMember_Success(t *testing.T) {
	userID := uuid.New().String()
	clubID := uuid.New().String()
	targetID := uuid.New().String()
	mock := &mockClubRepo{
		removeMemberFn: func(_ context.Context, _, _, _ string) error {
			return nil
		},
	}
	h := clubsHandler(mock)

	req := authedRequest(http.MethodDelete, "/", userID)
	req = clubsChiParams(req, map[string]string{
		"clubID": clubID,
		"userID": targetID,
	})

	rr := httptest.NewRecorder()
	h.RemoveMember(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rr.Code, rr.Body.String())
	}
}

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

// ── CreateTournament ──────────────────────────────────────────────────

func TestCreateTournament_Success(t *testing.T) {
	userID := uuid.New().String()
	clubID := uuid.New().String()
	templateID := uuid.New().String()
	now := time.Now().UTC()
	regDeadline := now.Add(48 * time.Hour).Format(time.RFC3339)
	startDate := now.Add(72 * time.Hour).Format(time.RFC3339)
	endDate := now.Add(96 * time.Hour).Format(time.RFC3339)

	mock := &mockClubRepo{
		createTournamentFn: func(_ context.Context, id, cID, uID, name string, desc *string, tplID string, maxP *int, rd, sd, ed time.Time) (*repository.TournamentOut, error) {
			return &repository.TournamentOut{
				ID:                   id,
				Name:                 name,
				OrganizerID:          uID,
				TemplateID:           tplID,
				Status:               "registration",
				RegistrationDeadline: rd,
				StartDate:            sd,
				EndDate:              ed,
				ClubID:               cID,
				CreatedAt:            time.Now().UTC(),
			}, nil
		},
	}
	h := clubsHandler(mock)

	req := clubsAuthedBody(http.MethodPost, "/tournaments", userID, map[string]any{
		"name":                  "Spring Tournament",
		"template_id":           templateID,
		"registration_deadline": regDeadline,
		"start_date":            startDate,
		"end_date":              endDate,
	})
	req = withChiURLParam(req, "clubID", clubID)

	rr := httptest.NewRecorder()
	h.CreateTournament(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	var result repository.TournamentOut
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if result.Name != "Spring Tournament" {
		t.Errorf("expected 'Spring Tournament', got '%s'", result.Name)
	}
	if result.Status != "registration" {
		t.Errorf("expected status 'registration', got '%s'", result.Status)
	}
}
