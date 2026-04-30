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
	submitTournamentScoreFn          func(ctx context.Context, clubID, tournamentID, userID, sessionID string) (int, int, error)
	addTournamentRoundFn             func(ctx context.Context, id, clubID, tournamentID, userID, name string, templateID *string, advancement *int) (*repository.TournamentRoundOut, error)
	listTournamentRoundsFn           func(ctx context.Context, clubID, tournamentID, userID string) ([]repository.TournamentRoundOut, error)
	startTournamentRoundFn           func(ctx context.Context, clubID, tournamentID, roundID, userID string) (*repository.TournamentRoundOut, error)
	submitTournamentRoundScoreFn     func(ctx context.Context, clubID, tournamentID, roundID, userID, sessionID string) (*repository.TournamentRoundScoreOut, error)
	getTournamentRoundLeaderboardFn  func(ctx context.Context, clubID, tournamentID, roundID, userID string) ([]repository.TournamentRoundScoreOut, error)
	completeTournamentRoundFn        func(ctx context.Context, clubID, tournamentID, roundID, userID string) (*repository.TournamentRoundOut, error)
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
func (m *mockClubRepo) AddTournamentRound(ctx context.Context, id, clubID, tournamentID, userID, name string, templateID *string, advancement *int) (*repository.TournamentRoundOut, error) {
	return m.addTournamentRoundFn(ctx, id, clubID, tournamentID, userID, name, templateID, advancement)
}
func (m *mockClubRepo) ListTournamentRounds(ctx context.Context, clubID, tournamentID, userID string) ([]repository.TournamentRoundOut, error) {
	return m.listTournamentRoundsFn(ctx, clubID, tournamentID, userID)
}
func (m *mockClubRepo) StartTournamentRound(ctx context.Context, clubID, tournamentID, roundID, userID string) (*repository.TournamentRoundOut, error) {
	return m.startTournamentRoundFn(ctx, clubID, tournamentID, roundID, userID)
}
func (m *mockClubRepo) SubmitTournamentRoundScore(ctx context.Context, clubID, tournamentID, roundID, userID, sessionID string) (*repository.TournamentRoundScoreOut, error) {
	return m.submitTournamentRoundScoreFn(ctx, clubID, tournamentID, roundID, userID, sessionID)
}
func (m *mockClubRepo) GetTournamentRoundLeaderboard(ctx context.Context, clubID, tournamentID, roundID, userID string) ([]repository.TournamentRoundScoreOut, error) {
	return m.getTournamentRoundLeaderboardFn(ctx, clubID, tournamentID, roundID, userID)
}
func (m *mockClubRepo) CompleteTournamentRound(ctx context.Context, clubID, tournamentID, roundID, userID string) (*repository.TournamentRoundOut, error) {
	return m.completeTournamentRoundFn(ctx, clubID, tournamentID, roundID, userID)
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

// ── ListInvites ──────────────────────────────────────────────────────

func TestListInvites_Success(t *testing.T) {
	userID := uuid.New().String()
	clubID := uuid.New().String()
	mock := &mockClubRepo{
		listInvitesFn: func(_ context.Context, cID, uID string, frontendURL string) ([]repository.InviteOut, error) {
			return []repository.InviteOut{
				{
					ID:        uuid.New().String(),
					Code:      "abc123",
					URL:       frontendURL + "/clubs/join/abc123",
					Active:    true,
					CreatedAt: time.Now().UTC(),
				},
			}, nil
		},
	}
	h := clubsHandler(mock)

	req := authedRequest(http.MethodGet, "/invites", userID)
	req = withChiURLParam(req, "clubID", clubID)

	rr := httptest.NewRecorder()
	h.ListInvites(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var result []repository.InviteOut
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 invite, got %d", len(result))
	}
	if !result[0].Active {
		t.Error("expected invite to be active")
	}
}

func TestListInvites_Unauthorized(t *testing.T) {
	mock := &mockClubRepo{
		listInvitesFn: func(_ context.Context, _, _ string, _ string) ([]repository.InviteOut, error) {
			return nil, errors.New("forbidden")
		},
	}
	h := clubsHandler(mock)

	req := authedRequest(http.MethodGet, "/invites", uuid.New().String())
	req = withChiURLParam(req, "clubID", uuid.New().String())

	rr := httptest.NewRecorder()
	h.ListInvites(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

// ── DeactivateInvite ─────────────────────────────────────────────────

func TestDeactivateInvite_Success(t *testing.T) {
	userID := uuid.New().String()
	clubID := uuid.New().String()
	inviteID := uuid.New().String()
	mock := &mockClubRepo{
		deactivateInviteFn: func(_ context.Context, _, _, _ string) error {
			return nil
		},
	}
	h := clubsHandler(mock)

	req := authedRequest(http.MethodDelete, "/invites/"+inviteID, userID)
	req = clubsChiParams(req, map[string]string{
		"clubID":   clubID,
		"inviteID": inviteID,
	})

	rr := httptest.NewRecorder()
	h.DeactivateInvite(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestDeactivateInvite_NotFound(t *testing.T) {
	mock := &mockClubRepo{
		deactivateInviteFn: func(_ context.Context, _, _, _ string) error {
			return errors.New("not found")
		},
	}
	h := clubsHandler(mock)

	req := authedRequest(http.MethodDelete, "/invites/bad-id", uuid.New().String())
	req = clubsChiParams(req, map[string]string{
		"clubID":   uuid.New().String(),
		"inviteID": "bad-id",
	})

	rr := httptest.NewRecorder()
	h.DeactivateInvite(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

// ── PreviewInvite ────────────────────────────────────────────────────

func TestPreviewInvite_Success(t *testing.T) {
	userID := uuid.New().String()
	role := "member"
	mock := &mockClubRepo{
		previewInviteFn: func(_ context.Context, code, uID string) (*repository.ClubOut, error) {
			return &repository.ClubOut{
				ID:          uuid.New().String(),
				Name:        "Preview Club",
				OwnerID:     uuid.New().String(),
				MemberCount: 5,
				MyRole:      &role,
				CreatedAt:   time.Now().UTC(),
			}, nil
		},
	}
	h := clubsHandler(mock)

	req := authedRequest(http.MethodGet, "/join/abc123", userID)
	req = withChiURLParam(req, "code", "abc123")

	rr := httptest.NewRecorder()
	h.PreviewInvite(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var result repository.ClubOut
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if result.Name != "Preview Club" {
		t.Errorf("expected 'Preview Club', got '%s'", result.Name)
	}
}

func TestPreviewInvite_NotFound(t *testing.T) {
	mock := &mockClubRepo{
		previewInviteFn: func(_ context.Context, _, _ string) (*repository.ClubOut, error) {
			return nil, errors.New("not found")
		},
	}
	h := clubsHandler(mock)

	req := authedRequest(http.MethodGet, "/join/bad-code", uuid.New().String())
	req = withChiURLParam(req, "code", "bad-code")

	rr := httptest.NewRecorder()
	h.PreviewInvite(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

// ── PromoteMember ────────────────────────────────────────────────────

func TestPromoteMember_Success(t *testing.T) {
	userID := uuid.New().String()
	clubID := uuid.New().String()
	targetID := uuid.New().String()
	mock := &mockClubRepo{
		promoteMemberFn: func(_ context.Context, _, _, _ string) error {
			return nil
		},
	}
	h := clubsHandler(mock)

	req := authedRequest(http.MethodPost, "/members/"+targetID+"/promote", userID)
	req = clubsChiParams(req, map[string]string{
		"clubID": clubID,
		"userID": targetID,
	})

	rr := httptest.NewRecorder()
	h.PromoteMember(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var result map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if result["detail"] != "Member promoted to admin" {
		t.Errorf("expected 'Member promoted to admin', got '%s'", result["detail"])
	}
}

func TestPromoteMember_Unauthorized(t *testing.T) {
	mock := &mockClubRepo{
		promoteMemberFn: func(_ context.Context, _, _, _ string) error {
			return errors.New("forbidden")
		},
	}
	h := clubsHandler(mock)

	req := authedRequest(http.MethodPost, "/members/someone/promote", uuid.New().String())
	req = clubsChiParams(req, map[string]string{
		"clubID": uuid.New().String(),
		"userID": "someone",
	})

	rr := httptest.NewRecorder()
	h.PromoteMember(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

// ── DemoteMember ─────────────────────────────────────────────────────

func TestDemoteMember_Success(t *testing.T) {
	userID := uuid.New().String()
	clubID := uuid.New().String()
	targetID := uuid.New().String()
	mock := &mockClubRepo{
		demoteMemberFn: func(_ context.Context, _, _, _ string) error {
			return nil
		},
	}
	h := clubsHandler(mock)

	req := authedRequest(http.MethodPost, "/members/"+targetID+"/demote", userID)
	req = clubsChiParams(req, map[string]string{
		"clubID": clubID,
		"userID": targetID,
	})

	rr := httptest.NewRecorder()
	h.DemoteMember(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var result map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if result["detail"] != "Member demoted to member" {
		t.Errorf("expected 'Member demoted to member', got '%s'", result["detail"])
	}
}

func TestDemoteMember_Unauthorized(t *testing.T) {
	mock := &mockClubRepo{
		demoteMemberFn: func(_ context.Context, _, _, _ string) error {
			return errors.New("forbidden")
		},
	}
	h := clubsHandler(mock)

	req := authedRequest(http.MethodPost, "/members/someone/demote", uuid.New().String())
	req = clubsChiParams(req, map[string]string{
		"clubID": uuid.New().String(),
		"userID": "someone",
	})

	rr := httptest.NewRecorder()
	h.DemoteMember(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

// ── Leaderboard ──────────────────────────────────────────────────────

func TestLeaderboard_Success(t *testing.T) {
	userID := uuid.New().String()
	clubID := uuid.New().String()
	mock := &mockClubRepo{
		leaderboardFn: func(_ context.Context, _, _ string, templateID *string) ([]repository.LeaderboardOut, error) {
			return []repository.LeaderboardOut{
				{
					TemplateID:   uuid.New().String(),
					TemplateName: "WA 70m",
					Entries: []repository.LeaderboardEntry{
						{UserID: userID, Username: "archer1", BestScore: 680, BestXCount: 12, SessionID: uuid.New().String(), AchievedAt: time.Now().UTC()},
					},
				},
			}, nil
		},
	}
	h := clubsHandler(mock)

	req := authedRequest(http.MethodGet, "/leaderboard", userID)
	req = withChiURLParam(req, "clubID", clubID)

	rr := httptest.NewRecorder()
	h.Leaderboard(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var result []repository.LeaderboardOut
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 leaderboard, got %d", len(result))
	}
	if result[0].TemplateName != "WA 70m" {
		t.Errorf("expected 'WA 70m', got '%s'", result[0].TemplateName)
	}
}

func TestLeaderboard_WithTemplateFilter(t *testing.T) {
	userID := uuid.New().String()
	clubID := uuid.New().String()
	templateID := uuid.New().String()
	var capturedTemplateID *string
	mock := &mockClubRepo{
		leaderboardFn: func(_ context.Context, _, _ string, tid *string) ([]repository.LeaderboardOut, error) {
			capturedTemplateID = tid
			return []repository.LeaderboardOut{}, nil
		},
	}
	h := clubsHandler(mock)

	req := authedRequest(http.MethodGet, "/leaderboard?template_id="+templateID, userID)
	req = withChiURLParam(req, "clubID", clubID)

	rr := httptest.NewRecorder()
	h.Leaderboard(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if capturedTemplateID == nil || *capturedTemplateID != templateID {
		t.Error("expected template_id to be passed to repo")
	}
}

func TestLeaderboard_NotFound(t *testing.T) {
	mock := &mockClubRepo{
		leaderboardFn: func(_ context.Context, _, _ string, _ *string) ([]repository.LeaderboardOut, error) {
			return nil, errors.New("not found")
		},
	}
	h := clubsHandler(mock)

	req := authedRequest(http.MethodGet, "/leaderboard", uuid.New().String())
	req = withChiURLParam(req, "clubID", uuid.New().String())

	rr := httptest.NewRecorder()
	h.Leaderboard(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

// ── Activity ─────────────────────────────────────────────────────────

func TestActivity_Success(t *testing.T) {
	userID := uuid.New().String()
	clubID := uuid.New().String()
	mock := &mockClubRepo{
		activityFn: func(_ context.Context, _, _ string, limit, offset int) ([]repository.ActivityItem, error) {
			return []repository.ActivityItem{
				{
					Type:         "score",
					UserID:       userID,
					Username:     "archer1",
					TemplateName: "WA 70m",
					Score:        680,
					XCount:       12,
					SessionID:    uuid.New().String(),
					OccurredAt:   time.Now().UTC(),
				},
			}, nil
		},
	}
	h := clubsHandler(mock)

	req := authedRequest(http.MethodGet, "/activity", userID)
	req = withChiURLParam(req, "clubID", clubID)

	rr := httptest.NewRecorder()
	h.Activity(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var result []repository.ActivityItem
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 activity item, got %d", len(result))
	}
	if result[0].Score != 680 {
		t.Errorf("expected score 680, got %d", result[0].Score)
	}
}

func TestActivity_CustomLimitOffset(t *testing.T) {
	userID := uuid.New().String()
	clubID := uuid.New().String()
	var capturedLimit, capturedOffset int
	mock := &mockClubRepo{
		activityFn: func(_ context.Context, _, _ string, limit, offset int) ([]repository.ActivityItem, error) {
			capturedLimit = limit
			capturedOffset = offset
			return []repository.ActivityItem{}, nil
		},
	}
	h := clubsHandler(mock)

	req := authedRequest(http.MethodGet, "/activity?limit=10&offset=5", userID)
	req = withChiURLParam(req, "clubID", clubID)

	rr := httptest.NewRecorder()
	h.Activity(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if capturedLimit != 10 {
		t.Errorf("expected limit 10, got %d", capturedLimit)
	}
	if capturedOffset != 5 {
		t.Errorf("expected offset 5, got %d", capturedOffset)
	}
}

func TestActivity_NotFound(t *testing.T) {
	mock := &mockClubRepo{
		activityFn: func(_ context.Context, _, _ string, _, _ int) ([]repository.ActivityItem, error) {
			return nil, errors.New("not found")
		},
	}
	h := clubsHandler(mock)

	req := authedRequest(http.MethodGet, "/activity", uuid.New().String())
	req = withChiURLParam(req, "clubID", uuid.New().String())

	rr := httptest.NewRecorder()
	h.Activity(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

// ── ListSharedRounds ─────────────────────────────────────────────────

func TestListSharedRounds_Success(t *testing.T) {
	userID := uuid.New().String()
	clubID := uuid.New().String()
	mock := &mockClubRepo{
		listSharedRoundsFn: func(_ context.Context, cID, uID string) ([]repository.ClubSharedRoundOut, error) {
			return []repository.ClubSharedRoundOut{
				{
					ID:               uuid.New().String(),
					ClubID:           cID,
					ClubName:         "My Club",
					TemplateID:       uuid.New().String(),
					TemplateName:     "WA 70m",
					SharedByUsername: "archer1",
					SharedAt:         time.Now().UTC(),
				},
			}, nil
		},
	}
	h := clubsHandler(mock)

	req := authedRequest(http.MethodGet, "/rounds", userID)
	req = withChiURLParam(req, "clubID", clubID)

	rr := httptest.NewRecorder()
	h.ListSharedRounds(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var result []repository.ClubSharedRoundOut
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 shared round, got %d", len(result))
	}
	if result[0].TemplateName != "WA 70m" {
		t.Errorf("expected 'WA 70m', got '%s'", result[0].TemplateName)
	}
}

func TestListSharedRounds_NotFound(t *testing.T) {
	mock := &mockClubRepo{
		listSharedRoundsFn: func(_ context.Context, _, _ string) ([]repository.ClubSharedRoundOut, error) {
			return nil, errors.New("not found")
		},
	}
	h := clubsHandler(mock)

	req := authedRequest(http.MethodGet, "/rounds", uuid.New().String())
	req = withChiURLParam(req, "clubID", uuid.New().String())

	rr := httptest.NewRecorder()
	h.ListSharedRounds(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

// ── RemoveSharedRound ────────────────────────────────────────────────

func TestRemoveSharedRound_Success(t *testing.T) {
	userID := uuid.New().String()
	clubID := uuid.New().String()
	roundID := uuid.New().String()
	mock := &mockClubRepo{
		removeSharedRoundFn: func(_ context.Context, _, _, _ string) error {
			return nil
		},
	}
	h := clubsHandler(mock)

	req := authedRequest(http.MethodDelete, "/rounds/"+roundID, userID)
	req = clubsChiParams(req, map[string]string{
		"clubID":  clubID,
		"roundID": roundID,
	})

	rr := httptest.NewRecorder()
	h.RemoveSharedRound(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestRemoveSharedRound_NotFound(t *testing.T) {
	mock := &mockClubRepo{
		removeSharedRoundFn: func(_ context.Context, _, _, _ string) error {
			return errors.New("not found")
		},
	}
	h := clubsHandler(mock)

	req := authedRequest(http.MethodDelete, "/rounds/bad-id", uuid.New().String())
	req = clubsChiParams(req, map[string]string{
		"clubID":  uuid.New().String(),
		"roundID": "bad-id",
	})

	rr := httptest.NewRecorder()
	h.RemoveSharedRound(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

// ── JoinViaInvite error paths ────────────────────────────────────────

func TestJoinClub_AlreadyMember(t *testing.T) {
	mock := &mockClubRepo{
		joinViaInviteFn: func(_ context.Context, _, _ string) (*repository.JoinResult, error) {
			return nil, repository.ErrAlreadyMember
		},
	}
	h := clubsHandler(mock)

	req := authedRequest(http.MethodPost, "/join/abc123", uuid.New().String())
	req = withChiURLParam(req, "code", "abc123")

	rr := httptest.NewRecorder()
	h.JoinViaInvite(rr, req)

	if rr.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestJoinClub_InviteNotFound(t *testing.T) {
	mock := &mockClubRepo{
		joinViaInviteFn: func(_ context.Context, _, _ string) (*repository.JoinResult, error) {
			return nil, errors.New("not found")
		},
	}
	h := clubsHandler(mock)

	req := authedRequest(http.MethodPost, "/join/bad-code", uuid.New().String())
	req = withChiURLParam(req, "code", "bad-code")

	rr := httptest.NewRecorder()
	h.JoinViaInvite(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

// ── RemoveMember error path ──────────────────────────────────────────

func TestRemoveMember_Unauthorized(t *testing.T) {
	mock := &mockClubRepo{
		removeMemberFn: func(_ context.Context, _, _, _ string) error {
			return errors.New("forbidden")
		},
	}
	h := clubsHandler(mock)

	req := authedRequest(http.MethodDelete, "/", uuid.New().String())
	req = clubsChiParams(req, map[string]string{
		"clubID": uuid.New().String(),
		"userID": uuid.New().String(),
	})

	rr := httptest.NewRecorder()
	h.RemoveMember(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

// ── CreateClub error path ────────────────────────────────────────────

func TestCreateClub_RepoError(t *testing.T) {
	mock := &mockClubRepo{
		createFn: func(_ context.Context, _, _ string, _ *string, _ string) (*repository.ClubOut, error) {
			return nil, errors.New("db error")
		},
	}
	h := clubsHandler(mock)

	req := clubsAuthedBody(http.MethodPost, "/", uuid.New().String(), map[string]string{"name": "Test"})
	rr := httptest.NewRecorder()
	h.Create(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rr.Code)
	}
}

// ── ListClubs error path ─────────────────────────────────────────────

func TestListClubs_RepoError(t *testing.T) {
	mock := &mockClubRepo{
		listForUserFn: func(_ context.Context, _ string) ([]repository.ClubOut, error) {
			return nil, errors.New("db error")
		},
	}
	h := clubsHandler(mock)

	req := authedRequest(http.MethodGet, "/", uuid.New().String())
	rr := httptest.NewRecorder()
	h.List(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rr.Code)
	}
}

// ── CreateInvite error path ──────────────────────────────────────────

func TestCreateInvite_Unauthorized(t *testing.T) {
	mock := &mockClubRepo{
		createInviteFn: func(_ context.Context, _, _, _, _ string, _ *int, _ *time.Time, _ string) (*repository.InviteOut, error) {
			return nil, errors.New("forbidden")
		},
	}
	h := clubsHandler(mock)

	req := clubsAuthedBody(http.MethodPost, "/invites", uuid.New().String(), map[string]any{})
	req = withChiURLParam(req, "clubID", uuid.New().String())

	rr := httptest.NewRecorder()
	h.CreateInvite(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestCreateInvite_WithExpiresInHours(t *testing.T) {
	userID := uuid.New().String()
	clubID := uuid.New().String()
	mock := &mockClubRepo{
		createInviteFn: func(_ context.Context, id, cID, uID, code string, maxUses *int, expiresAt *time.Time, frontendURL string) (*repository.InviteOut, error) {
			if expiresAt == nil {
				return nil, errors.New("expected expiresAt to be set")
			}
			return &repository.InviteOut{
				ID:        id,
				Code:      code,
				URL:       frontendURL + "/clubs/join/" + code,
				ExpiresAt: expiresAt,
				Active:    true,
				CreatedAt: time.Now().UTC(),
			}, nil
		},
	}
	h := clubsHandler(mock)

	req := clubsAuthedBody(http.MethodPost, "/invites", userID, map[string]any{
		"expires_in_hours": 24,
	})
	req = withChiURLParam(req, "clubID", clubID)

	rr := httptest.NewRecorder()
	h.CreateInvite(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}
}

