package handler

import (
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

type ClubsHandler struct {
	Clubs *repository.ClubRepo
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

// ── Teams ─────────────────────────────────────────────────────────────

type teamCreate struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
	LeaderID    string  `json:"leader_id"`
}

func (h *ClubsHandler) CreateTeam(w http.ResponseWriter, r *http.Request) {
	clubID := chi.URLParam(r, "clubID")
	userID := middleware.GetUserID(r.Context())

	var req teamCreate
	if !Decode(w, r, &req) {
		return
	}

	id := uuid.New().String()
	team, err := h.Clubs.CreateTeam(r.Context(), id, clubID, userID, req.Name, req.Description, req.LeaderID)
	if err != nil {
		Error(w, http.StatusUnauthorized, "Only owner or admin can create teams")
		return
	}

	JSON(w, http.StatusCreated, team)
}

func (h *ClubsHandler) ListTeams(w http.ResponseWriter, r *http.Request) {
	clubID := chi.URLParam(r, "clubID")
	userID := middleware.GetUserID(r.Context())

	teams, err := h.Clubs.ListTeams(r.Context(), clubID, userID)
	if err != nil {
		Error(w, http.StatusNotFound, "Club not found")
		return
	}

	JSON(w, http.StatusOK, teams)
}

func (h *ClubsHandler) GetTeamDetail(w http.ResponseWriter, r *http.Request) {
	clubID := chi.URLParam(r, "clubID")
	teamID := chi.URLParam(r, "teamID")
	userID := middleware.GetUserID(r.Context())

	detail, err := h.Clubs.GetTeamDetail(r.Context(), clubID, teamID, userID)
	if err != nil {
		Error(w, http.StatusNotFound, "Team not found")
		return
	}

	JSON(w, http.StatusOK, detail)
}

type teamUpdate struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	LeaderID    *string `json:"leader_id"`
}

func (h *ClubsHandler) UpdateTeam(w http.ResponseWriter, r *http.Request) {
	clubID := chi.URLParam(r, "clubID")
	teamID := chi.URLParam(r, "teamID")
	userID := middleware.GetUserID(r.Context())

	var req teamUpdate
	if !Decode(w, r, &req) {
		return
	}

	team, err := h.Clubs.UpdateTeam(r.Context(), clubID, teamID, userID, req.Name, req.Description, req.LeaderID)
	if err != nil {
		Error(w, http.StatusNotFound, "Team not found")
		return
	}

	JSON(w, http.StatusOK, team)
}

func (h *ClubsHandler) DeleteTeam(w http.ResponseWriter, r *http.Request) {
	clubID := chi.URLParam(r, "clubID")
	teamID := chi.URLParam(r, "teamID")
	userID := middleware.GetUserID(r.Context())

	if err := h.Clubs.DeleteTeam(r.Context(), clubID, teamID, userID); err != nil {
		Error(w, http.StatusNotFound, "Team not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ClubsHandler) AddTeamMember(w http.ResponseWriter, r *http.Request) {
	clubID := chi.URLParam(r, "clubID")
	teamID := chi.URLParam(r, "teamID")
	targetUserID := chi.URLParam(r, "memberUserID")
	userID := middleware.GetUserID(r.Context())

	if err := h.Clubs.AddTeamMember(r.Context(), clubID, teamID, targetUserID, userID); err != nil {
		if err == repository.ErrAlreadyMember {
			Error(w, http.StatusConflict, "User is already on this team")
			return
		}
		Error(w, http.StatusUnauthorized, "Cannot add team member")
		return
	}

	JSON(w, http.StatusCreated, map[string]string{"detail": "Member added to team"})
}

func (h *ClubsHandler) RemoveTeamMember(w http.ResponseWriter, r *http.Request) {
	clubID := chi.URLParam(r, "clubID")
	teamID := chi.URLParam(r, "teamID")
	targetUserID := chi.URLParam(r, "memberUserID")
	userID := middleware.GetUserID(r.Context())

	if err := h.Clubs.RemoveTeamMember(r.Context(), clubID, teamID, targetUserID, userID); err != nil {
		Error(w, http.StatusNotFound, "Team member not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
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

// ── Tournaments ───────────────────────────────────────────────────────

type tournamentCreate struct {
	Name                 string  `json:"name"`
	Description          *string `json:"description"`
	TemplateID           string  `json:"template_id"`
	MaxParticipants      *int    `json:"max_participants"`
	RegistrationDeadline string  `json:"registration_deadline"`
	StartDate            string  `json:"start_date"`
	EndDate              string  `json:"end_date"`
}

func (h *ClubsHandler) CreateTournament(w http.ResponseWriter, r *http.Request) {
	clubID := chi.URLParam(r, "clubID")
	userID := middleware.GetUserID(r.Context())

	var req tournamentCreate
	if !Decode(w, r, &req) {
		return
	}

	regDeadline, err := time.Parse(time.RFC3339, req.RegistrationDeadline)
	if err != nil {
		ValidationError(w, "Invalid registration_deadline")
		return
	}
	startDate, err := time.Parse(time.RFC3339, req.StartDate)
	if err != nil {
		ValidationError(w, "Invalid start_date")
		return
	}
	endDate, err := time.Parse(time.RFC3339, req.EndDate)
	if err != nil {
		ValidationError(w, "Invalid end_date")
		return
	}

	id := uuid.New().String()
	tourney, err := h.Clubs.CreateTournament(r.Context(), id, clubID, userID, req.Name, req.Description, req.TemplateID, req.MaxParticipants, regDeadline, startDate, endDate)
	if err != nil {
		Error(w, http.StatusUnauthorized, "Only owner or admin can create tournaments")
		return
	}

	JSON(w, http.StatusCreated, tourney)
}

func (h *ClubsHandler) ListTournaments(w http.ResponseWriter, r *http.Request) {
	clubID := chi.URLParam(r, "clubID")
	userID := middleware.GetUserID(r.Context())

	var status *string
	if s := r.URL.Query().Get("status"); s != "" {
		status = &s
	}

	tournaments, err := h.Clubs.ListTournaments(r.Context(), clubID, userID, status)
	if err != nil {
		Error(w, http.StatusNotFound, "Club not found")
		return
	}

	JSON(w, http.StatusOK, tournaments)
}

func (h *ClubsHandler) GetTournament(w http.ResponseWriter, r *http.Request) {
	clubID := chi.URLParam(r, "clubID")
	tournamentID := chi.URLParam(r, "tournamentID")
	userID := middleware.GetUserID(r.Context())

	detail, err := h.Clubs.GetTournamentDetail(r.Context(), clubID, tournamentID, userID)
	if err != nil {
		Error(w, http.StatusNotFound, "Tournament not found")
		return
	}

	JSON(w, http.StatusOK, detail)
}

func (h *ClubsHandler) RegisterForTournament(w http.ResponseWriter, r *http.Request) {
	clubID := chi.URLParam(r, "clubID")
	tournamentID := chi.URLParam(r, "tournamentID")
	userID := middleware.GetUserID(r.Context())

	err := h.Clubs.RegisterForTournament(r.Context(), clubID, tournamentID, userID)
	if err != nil {
		if err == repository.ErrAlreadyMember {
			Error(w, http.StatusConflict, "Already registered")
			return
		}
		Error(w, http.StatusNotFound, "Tournament not found or not open for registration")
		return
	}

	JSON(w, http.StatusCreated, map[string]string{"message": "Registered successfully"})
}

func (h *ClubsHandler) StartTournament(w http.ResponseWriter, r *http.Request) {
	clubID := chi.URLParam(r, "clubID")
	tournamentID := chi.URLParam(r, "tournamentID")
	userID := middleware.GetUserID(r.Context())

	tourney, err := h.Clubs.StartTournament(r.Context(), clubID, tournamentID, userID)
	if err != nil {
		Error(w, http.StatusForbidden, "Cannot start tournament")
		return
	}

	JSON(w, http.StatusOK, tourney)
}

func (h *ClubsHandler) TournamentLeaderboard(w http.ResponseWriter, r *http.Request) {
	clubID := chi.URLParam(r, "clubID")
	tournamentID := chi.URLParam(r, "tournamentID")
	userID := middleware.GetUserID(r.Context())

	lb, err := h.Clubs.TournamentLeaderboard(r.Context(), clubID, tournamentID, userID)
	if err != nil {
		Error(w, http.StatusNotFound, "Tournament not found")
		return
	}

	JSON(w, http.StatusOK, lb)
}

func (h *ClubsHandler) CompleteTournament(w http.ResponseWriter, r *http.Request) {
	clubID := chi.URLParam(r, "clubID")
	tournamentID := chi.URLParam(r, "tournamentID")
	userID := middleware.GetUserID(r.Context())

	tourney, err := h.Clubs.CompleteTournament(r.Context(), clubID, tournamentID, userID)
	if err != nil {
		Error(w, http.StatusForbidden, "Cannot complete tournament")
		return
	}

	JSON(w, http.StatusOK, tourney)
}

func (h *ClubsHandler) WithdrawFromTournament(w http.ResponseWriter, r *http.Request) {
	clubID := chi.URLParam(r, "clubID")
	tournamentID := chi.URLParam(r, "tournamentID")
	userID := middleware.GetUserID(r.Context())

	if err := h.Clubs.WithdrawFromTournament(r.Context(), clubID, tournamentID, userID); err != nil {
		Error(w, http.StatusNotFound, "Tournament not found or cannot withdraw")
		return
	}

	JSON(w, http.StatusOK, map[string]string{"message": "Withdrawn successfully"})
}

func (h *ClubsHandler) SubmitTournamentScore(w http.ResponseWriter, r *http.Request) {
	clubID := chi.URLParam(r, "clubID")
	tournamentID := chi.URLParam(r, "tournamentID")
	userID := middleware.GetUserID(r.Context())

	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		ValidationError(w, "session_id query parameter is required")
		return
	}

	score, xcount, err := h.Clubs.SubmitTournamentScore(r.Context(), clubID, tournamentID, userID, sessionID)
	if err != nil {
		Error(w, http.StatusNotFound, "Cannot submit score")
		return
	}

	JSON(w, http.StatusOK, map[string]any{
		"message":      "Score submitted successfully",
		"final_score":  score,
		"final_x_count": xcount,
	})
}

// ── Helpers ───────────────────────────────────────────────────────────

func generateInviteCode() string {
	b := make([]byte, 18) // 18 bytes → 24 base64 chars
	rand.Read(b)
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(b)[:24]
}
