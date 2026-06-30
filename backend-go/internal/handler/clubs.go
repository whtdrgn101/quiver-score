package handler

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/quiverscore/backend-go/internal/config"
	"github.com/quiverscore/backend-go/internal/middleware"
	"github.com/quiverscore/backend-go/internal/repository"
)

type ClubRepository interface {
	Create(ctx context.Context, id, name string, description *string, ownerID string) (*repository.ClubOut, error)
	ListForUser(ctx context.Context, userID string) ([]repository.ClubOut, error)
	GetDetail(ctx context.Context, clubID, userID string) (*repository.ClubDetailOut, error)
	Update(ctx context.Context, clubID, userID string, name, description *string) (*repository.ClubOut, error)
	Delete(ctx context.Context, clubID, userID string) error
	CreateInvite(ctx context.Context, id, clubID, userID, code string, maxUses *int, expiresAt *time.Time, frontendURL string) (*repository.InviteOut, error)
	ListInvites(ctx context.Context, clubID, userID string, frontendURL string) ([]repository.InviteOut, error)
	DeactivateInvite(ctx context.Context, clubID, inviteID, userID string) error
	PreviewInvite(ctx context.Context, code, userID string) (*repository.ClubOut, error)
	JoinViaInvite(ctx context.Context, code, userID string) (*repository.JoinResult, error)
	PromoteMember(ctx context.Context, clubID, targetUserID, userID string) error
	DemoteMember(ctx context.Context, clubID, targetUserID, userID string) error
	RemoveMember(ctx context.Context, clubID, targetUserID, userID string) error
	Leaderboard(ctx context.Context, clubID, userID string, templateID *string) ([]repository.LeaderboardOut, error)
	Activity(ctx context.Context, clubID, userID string, limit, offset int) ([]repository.ActivityItem, error)
	CreateEvent(ctx context.Context, id, clubID, userID, name string, description *string, templateID string, eventDate time.Time, location *string) (*repository.EventOut, error)
	ListEvents(ctx context.Context, clubID, userID string) ([]repository.EventOut, error)
	GetEvent(ctx context.Context, clubID, eventID, userID string) (*repository.EventOut, error)
	UpdateEvent(ctx context.Context, clubID, eventID, userID string, name *string, description *string, descriptionSet bool, eventDate *time.Time, location *string, locationSet bool) (*repository.EventOut, error)
	DeleteEvent(ctx context.Context, clubID, eventID, userID string) error
	RSVPEvent(ctx context.Context, clubID, eventID, userID, status string) (*repository.EventOut, error)
	CreateTeam(ctx context.Context, id, clubID, userID, name string, description *string, leaderID string) (*repository.TeamOut, error)
	ListTeams(ctx context.Context, clubID, userID string) ([]repository.TeamOut, error)
	GetTeamDetail(ctx context.Context, clubID, teamID, userID string) (*repository.TeamDetailOut, error)
	UpdateTeam(ctx context.Context, clubID, teamID, userID string, name, description *string, leaderID *string) (*repository.TeamOut, error)
	DeleteTeam(ctx context.Context, clubID, teamID, userID string) error
	AddTeamMember(ctx context.Context, clubID, teamID, targetUserID, userID string) error
	RemoveTeamMember(ctx context.Context, clubID, teamID, targetUserID, userID string) error
	ListSharedRounds(ctx context.Context, clubID, userID string) ([]repository.ClubSharedRoundOut, error)
	RemoveSharedRound(ctx context.Context, clubID, roundID, userID string) error
	CreateTournament(ctx context.Context, id, clubID, userID, name string, description *string, templateID string, maxParticipants *int, registrationDeadline, startDate, endDate time.Time) (*repository.TournamentOut, error)
	ListTournaments(ctx context.Context, clubID, userID string, status *string) ([]repository.TournamentOut, error)
	GetTournamentDetail(ctx context.Context, clubID, tournamentID, userID string) (*repository.TournamentDetailOut, error)
	RegisterForTournament(ctx context.Context, clubID, tournamentID, userID string) error
	StartTournament(ctx context.Context, clubID, tournamentID, userID string) (*repository.TournamentOut, error)
	TournamentLeaderboard(ctx context.Context, clubID, tournamentID, userID string) ([]repository.TournamentLeaderboardEntry, error)
	CompleteTournament(ctx context.Context, clubID, tournamentID, userID string) (*repository.TournamentOut, error)
	WithdrawFromTournament(ctx context.Context, clubID, tournamentID, userID string) error
	SubmitTournamentScore(ctx context.Context, clubID, tournamentID, userID, sessionID string) (int, int, error)
	AddTournamentRound(ctx context.Context, id, clubID, tournamentID, userID, name string, templateID *string, advancement *int, roundType string) (*repository.TournamentRoundOut, error)
	ListTournamentRounds(ctx context.Context, clubID, tournamentID, userID string) ([]repository.TournamentRoundOut, error)
	StartTournamentRound(ctx context.Context, clubID, tournamentID, roundID, userID string) (*repository.TournamentRoundOut, error)
	SubmitTournamentRoundScore(ctx context.Context, clubID, tournamentID, roundID, userID, sessionID string) (*repository.TournamentRoundScoreOut, error)
	GetTournamentRoundLeaderboard(ctx context.Context, clubID, tournamentID, roundID, userID string) ([]repository.TournamentRoundScoreOut, error)
	CompleteTournamentRound(ctx context.Context, clubID, tournamentID, roundID, userID string) (*repository.TournamentRoundOut, error)
	GetTournamentMatchups(ctx context.Context, roundID string) ([]repository.TournamentMatchupOut, error)
	SubmitMatchupScore(ctx context.Context, clubID, tournamentID, roundID, matchupID, userID, sessionID string) (*repository.TournamentMatchupOut, error)
	UpdateMatchup(ctx context.Context, clubID, tournamentID, roundID, matchupID, userID string, scoreA, scoreB *int, winnerID *string) (*repository.TournamentMatchupOut, error)
}

type ClubsHandler struct {
	Clubs ClubRepository
	Cfg   *config.Config
}

func (h *ClubsHandler) Routes(r chi.Router) {
	r.Use(middleware.RequireAuth(h.Cfg.SecretKey))

	r.Post("/", h.Create)
	r.Get("/", h.List)

	// Invite join (before /{club_id} to avoid conflict)
	r.Get("/join/{code}", h.PreviewInvite)
	r.Post("/join/{code}", h.JoinViaInvite)

	r.Route("/{clubID}", func(cr chi.Router) {
		cr.Get("/", h.GetDetail)
		cr.Patch("/", h.Update)
		cr.Delete("/", h.Delete)

		// Invites
		cr.Post("/invites", h.CreateInvite)
		cr.Get("/invites", h.ListInvites)
		cr.Delete("/invites/{inviteID}", h.DeactivateInvite)

		// Members
		cr.Post("/members/{userID}/promote", h.PromoteMember)
		cr.Post("/members/{userID}/demote", h.DemoteMember)
		cr.Delete("/members/{userID}", h.RemoveMember)

		// Leaderboard & Activity
		cr.Get("/leaderboard", h.Leaderboard)
		cr.Get("/activity", h.Activity)

		// Events
		cr.Post("/events", h.CreateEvent)
		cr.Get("/events", h.ListEvents)
		cr.Get("/events/{eventID}", h.GetEvent)
		cr.Patch("/events/{eventID}", h.UpdateEvent)
		cr.Delete("/events/{eventID}", h.DeleteEvent)
		cr.Post("/events/{eventID}/rsvp", h.RSVPEvent)

		// Teams
		cr.Post("/teams", h.CreateTeam)
		cr.Get("/teams", h.ListTeams)
		cr.Get("/teams/{teamID}", h.GetTeamDetail)
		cr.Patch("/teams/{teamID}", h.UpdateTeam)
		cr.Delete("/teams/{teamID}", h.DeleteTeam)
		cr.Post("/teams/{teamID}/members/{memberUserID}", h.AddTeamMember)
		cr.Delete("/teams/{teamID}/members/{memberUserID}", h.RemoveTeamMember)

		// Shared rounds
		cr.Get("/rounds", h.ListSharedRounds)
		cr.Delete("/rounds/{roundID}", h.RemoveSharedRound)

		// Tournaments
		cr.Post("/tournaments", h.CreateTournament)
		cr.Get("/tournaments", h.ListTournaments)
		cr.Get("/tournaments/{tournamentID}", h.GetTournament)
		cr.Post("/tournaments/{tournamentID}/register", h.RegisterForTournament)
		cr.Post("/tournaments/{tournamentID}/start", h.StartTournament)
		cr.Get("/tournaments/{tournamentID}/leaderboard", h.TournamentLeaderboard)
		cr.Post("/tournaments/{tournamentID}/complete", h.CompleteTournament)
		cr.Post("/tournaments/{tournamentID}/withdraw", h.WithdrawFromTournament)
		cr.Post("/tournaments/{tournamentID}/submit-score", h.SubmitTournamentScore)

		// Tournament Rounds
		cr.Post("/tournaments/{tournamentID}/rounds", h.AddTournamentRound)
		cr.Get("/tournaments/{tournamentID}/rounds", h.ListTournamentRounds)
		cr.Post("/tournaments/{tournamentID}/rounds/{roundID}/start", h.StartTournamentRound)
		cr.Post("/tournaments/{tournamentID}/rounds/{roundID}/submit-score", h.SubmitTournamentRoundScore)
		cr.Get("/tournaments/{tournamentID}/rounds/{roundID}/leaderboard", h.GetTournamentRoundLeaderboard)
		cr.Post("/tournaments/{tournamentID}/rounds/{roundID}/complete", h.CompleteTournamentRound)

		// Tournament Matchups
		cr.Get("/tournaments/{tournamentID}/rounds/{roundID}/matchups", h.GetTournamentMatchups)
		cr.Post("/tournaments/{tournamentID}/rounds/{roundID}/matchups/{matchupID}/submit-score", h.SubmitTournamentMatchupScore)
		cr.Put("/tournaments/{tournamentID}/rounds/{roundID}/matchups/{matchupID}", h.UpdateTournamentMatchup)
	})
}

// ── Club CRUD ─────────────────────────────────────────────────────────

type clubCreate struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

func (h *ClubsHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req clubCreate
	if !Decode(w, r, &req) {
		return
	}
	if req.Name == "" {
		ValidationError(w, "name is required")
		return
	}

	userID := middleware.GetUserID(r.Context())
	id := uuid.New().String()

	club, err := h.Clubs.Create(r.Context(), id, req.Name, req.Description, userID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	JSON(w, http.StatusCreated, club)
}

func (h *ClubsHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	clubs, err := h.Clubs.ListForUser(r.Context(), userID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	JSON(w, http.StatusOK, clubs)
}

func (h *ClubsHandler) GetDetail(w http.ResponseWriter, r *http.Request) {
	clubID := chi.URLParam(r, "clubID")
	userID := middleware.GetUserID(r.Context())

	detail, err := h.Clubs.GetDetail(r.Context(), clubID, userID)
	if err != nil {
		Error(w, http.StatusNotFound, "Club not found")
		return
	}

	JSON(w, http.StatusOK, detail)
}

type clubUpdate struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

func (h *ClubsHandler) Update(w http.ResponseWriter, r *http.Request) {
	clubID := chi.URLParam(r, "clubID")
	userID := middleware.GetUserID(r.Context())

	var req clubUpdate
	if !Decode(w, r, &req) {
		return
	}

	club, err := h.Clubs.Update(r.Context(), clubID, userID, req.Name, req.Description)
	if err != nil {
		Error(w, http.StatusUnauthorized, "Only the owner can update the club")
		return
	}

	JSON(w, http.StatusOK, club)
}

func (h *ClubsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	clubID := chi.URLParam(r, "clubID")
	userID := middleware.GetUserID(r.Context())

	if err := h.Clubs.Delete(r.Context(), clubID, userID); err != nil {
		Error(w, http.StatusUnauthorized, "Only the owner can delete the club")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ── Invites ───────────────────────────────────────────────────────────

type inviteCreate struct {
	MaxUses        *int `json:"max_uses"`
	ExpiresInHours *int `json:"expires_in_hours"`
}

func (h *ClubsHandler) CreateInvite(w http.ResponseWriter, r *http.Request) {
	clubID := chi.URLParam(r, "clubID")
	userID := middleware.GetUserID(r.Context())

	var req inviteCreate
	if !Decode(w, r, &req) {
		return
	}

	code := generateInviteCode()
	id := uuid.New().String()

	var expiresAt *time.Time
	if req.ExpiresInHours != nil {
		t := time.Now().UTC().Add(time.Duration(*req.ExpiresInHours) * time.Hour)
		expiresAt = &t
	}

	invite, err := h.Clubs.CreateInvite(r.Context(), id, clubID, userID, code, req.MaxUses, expiresAt, h.Cfg.FrontendURL)
	if err != nil {
		Error(w, http.StatusUnauthorized, "Only owner or admin can create invites")
		return
	}

	JSON(w, http.StatusCreated, invite)
}

func (h *ClubsHandler) ListInvites(w http.ResponseWriter, r *http.Request) {
	clubID := chi.URLParam(r, "clubID")
	userID := middleware.GetUserID(r.Context())

	invites, err := h.Clubs.ListInvites(r.Context(), clubID, userID, h.Cfg.FrontendURL)
	if err != nil {
		Error(w, http.StatusUnauthorized, "Only owner or admin can view invites")
		return
	}

	JSON(w, http.StatusOK, invites)
}

func (h *ClubsHandler) DeactivateInvite(w http.ResponseWriter, r *http.Request) {
	clubID := chi.URLParam(r, "clubID")
	inviteID := chi.URLParam(r, "inviteID")
	userID := middleware.GetUserID(r.Context())

	if err := h.Clubs.DeactivateInvite(r.Context(), clubID, inviteID, userID); err != nil {
		Error(w, http.StatusNotFound, "Invite not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ClubsHandler) PreviewInvite(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	userID := middleware.GetUserID(r.Context())

	club, err := h.Clubs.PreviewInvite(r.Context(), code, userID)
	if err != nil {
		Error(w, http.StatusNotFound, "Invite not found or expired")
		return
	}

	JSON(w, http.StatusOK, club)
}

func (h *ClubsHandler) JoinViaInvite(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	userID := middleware.GetUserID(r.Context())

	result, err := h.Clubs.JoinViaInvite(r.Context(), code, userID)
	if err != nil {
		if err == repository.ErrAlreadyMember {
			Error(w, http.StatusConflict, "Already a member of this club")
			return
		}
		Error(w, http.StatusNotFound, "Invite not found or expired")
		return
	}

	JSON(w, http.StatusOK, result)
}

// ── Members ───────────────────────────────────────────────────────────

func (h *ClubsHandler) PromoteMember(w http.ResponseWriter, r *http.Request) {
	clubID := chi.URLParam(r, "clubID")
	targetUserID := chi.URLParam(r, "userID")
	userID := middleware.GetUserID(r.Context())

	if err := h.Clubs.PromoteMember(r.Context(), clubID, targetUserID, userID); err != nil {
		Error(w, http.StatusUnauthorized, "Only the owner can promote members")
		return
	}

	JSON(w, http.StatusOK, map[string]string{"detail": "Member promoted to admin"})
}

func (h *ClubsHandler) DemoteMember(w http.ResponseWriter, r *http.Request) {
	clubID := chi.URLParam(r, "clubID")
	targetUserID := chi.URLParam(r, "userID")
	userID := middleware.GetUserID(r.Context())

	if err := h.Clubs.DemoteMember(r.Context(), clubID, targetUserID, userID); err != nil {
		Error(w, http.StatusUnauthorized, "Only the owner can demote members")
		return
	}

	JSON(w, http.StatusOK, map[string]string{"detail": "Member demoted to member"})
}

func (h *ClubsHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	clubID := chi.URLParam(r, "clubID")
	targetUserID := chi.URLParam(r, "userID")
	userID := middleware.GetUserID(r.Context())

	if err := h.Clubs.RemoveMember(r.Context(), clubID, targetUserID, userID); err != nil {
		Error(w, http.StatusUnauthorized, "Cannot remove this member")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ── Leaderboard & Activity ────────────────────────────────────────────

func (h *ClubsHandler) Leaderboard(w http.ResponseWriter, r *http.Request) {
	clubID := chi.URLParam(r, "clubID")
	userID := middleware.GetUserID(r.Context())

	var templateID *string
	if tid := r.URL.Query().Get("template_id"); tid != "" {
		templateID = &tid
	}

	lb, err := h.Clubs.Leaderboard(r.Context(), clubID, userID, templateID)
	if err != nil {
		Error(w, http.StatusNotFound, "Club not found")
		return
	}

	JSON(w, http.StatusOK, lb)
}

func (h *ClubsHandler) Activity(w http.ResponseWriter, r *http.Request) {
	clubID := chi.URLParam(r, "clubID")
	userID := middleware.GetUserID(r.Context())

	limit := 20
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v >= 1 && v <= 100 {
			limit = v
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}

	items, err := h.Clubs.Activity(r.Context(), clubID, userID, limit, offset)
	if err != nil {
		Error(w, http.StatusNotFound, "Club not found")
		return
	}

	JSON(w, http.StatusOK, items)
}

// ── Shared Rounds ─────────────────────────────────────────────────────

func (h *ClubsHandler) ListSharedRounds(w http.ResponseWriter, r *http.Request) {
	clubID := chi.URLParam(r, "clubID")
	userID := middleware.GetUserID(r.Context())

	rounds, err := h.Clubs.ListSharedRounds(r.Context(), clubID, userID)
	if err != nil {
		Error(w, http.StatusNotFound, "Club not found")
		return
	}

	JSON(w, http.StatusOK, rounds)
}

func (h *ClubsHandler) RemoveSharedRound(w http.ResponseWriter, r *http.Request) {
	clubID := chi.URLParam(r, "clubID")
	roundID := chi.URLParam(r, "roundID")
	userID := middleware.GetUserID(r.Context())

	if err := h.Clubs.RemoveSharedRound(r.Context(), clubID, roundID, userID); err != nil {
		Error(w, http.StatusNotFound, "Shared round not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ── Helpers ───────────────────────────────────────────────────────────

func generateInviteCode() string {
	b := make([]byte, 18) // 18 bytes → 24 base64 chars
	rand.Read(b)
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(b)[:24]
}
