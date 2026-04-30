package handler

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/quiverscore/backend-go/internal/middleware"
	"github.com/quiverscore/backend-go/internal/repository"
)

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
		log.Printf("ListTournaments error club=%s user=%s: %v", clubID, userID, err)
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
		"message":        "Score submitted successfully",
		"final_score":    score,
		"final_x_count":  xcount,
	})
}

// ── Tournament Rounds ────────────────────────────────────────────────

type tournamentRoundCreate struct {
	Name        string  `json:"name"`
	TemplateID  *string `json:"template_id"`
	Advancement *int    `json:"advancement"`
}

func (h *ClubsHandler) AddTournamentRound(w http.ResponseWriter, r *http.Request) {
	clubID := chi.URLParam(r, "clubID")
	tournamentID := chi.URLParam(r, "tournamentID")
	userID := middleware.GetUserID(r.Context())

	var req tournamentRoundCreate
	if !Decode(w, r, &req) {
		return
	}
	if req.Name == "" {
		ValidationError(w, "name is required")
		return
	}

	id := uuid.New().String()
	round, err := h.Clubs.AddTournamentRound(r.Context(), id, clubID, tournamentID, userID, req.Name, req.TemplateID, req.Advancement)
	if err != nil {
		Error(w, http.StatusForbidden, "Cannot add round")
		return
	}

	JSON(w, http.StatusCreated, round)
}

func (h *ClubsHandler) ListTournamentRounds(w http.ResponseWriter, r *http.Request) {
	clubID := chi.URLParam(r, "clubID")
	tournamentID := chi.URLParam(r, "tournamentID")
	userID := middleware.GetUserID(r.Context())

	rounds, err := h.Clubs.ListTournamentRounds(r.Context(), clubID, tournamentID, userID)
	if err != nil {
		Error(w, http.StatusNotFound, "Tournament not found")
		return
	}

	JSON(w, http.StatusOK, rounds)
}

func (h *ClubsHandler) StartTournamentRound(w http.ResponseWriter, r *http.Request) {
	clubID := chi.URLParam(r, "clubID")
	tournamentID := chi.URLParam(r, "tournamentID")
	roundID := chi.URLParam(r, "roundID")
	userID := middleware.GetUserID(r.Context())

	round, err := h.Clubs.StartTournamentRound(r.Context(), clubID, tournamentID, roundID, userID)
	if err != nil {
		Error(w, http.StatusForbidden, "Cannot start round")
		return
	}

	JSON(w, http.StatusOK, round)
}

func (h *ClubsHandler) SubmitTournamentRoundScore(w http.ResponseWriter, r *http.Request) {
	clubID := chi.URLParam(r, "clubID")
	tournamentID := chi.URLParam(r, "tournamentID")
	roundID := chi.URLParam(r, "roundID")
	userID := middleware.GetUserID(r.Context())

	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		ValidationError(w, "session_id query parameter is required")
		return
	}

	score, err := h.Clubs.SubmitTournamentRoundScore(r.Context(), clubID, tournamentID, roundID, userID, sessionID)
	if err != nil {
		Error(w, http.StatusNotFound, "Cannot submit round score")
		return
	}

	JSON(w, http.StatusOK, score)
}

func (h *ClubsHandler) GetTournamentRoundLeaderboard(w http.ResponseWriter, r *http.Request) {
	clubID := chi.URLParam(r, "clubID")
	tournamentID := chi.URLParam(r, "tournamentID")
	roundID := chi.URLParam(r, "roundID")
	userID := middleware.GetUserID(r.Context())

	lb, err := h.Clubs.GetTournamentRoundLeaderboard(r.Context(), clubID, tournamentID, roundID, userID)
	if err != nil {
		Error(w, http.StatusNotFound, "Round not found")
		return
	}

	JSON(w, http.StatusOK, lb)
}

func (h *ClubsHandler) CompleteTournamentRound(w http.ResponseWriter, r *http.Request) {
	clubID := chi.URLParam(r, "clubID")
	tournamentID := chi.URLParam(r, "tournamentID")
	roundID := chi.URLParam(r, "roundID")
	userID := middleware.GetUserID(r.Context())

	round, err := h.Clubs.CompleteTournamentRound(r.Context(), clubID, tournamentID, roundID, userID)
	if err != nil {
		Error(w, http.StatusForbidden, "Cannot complete round")
		return
	}

	JSON(w, http.StatusOK, round)
}
