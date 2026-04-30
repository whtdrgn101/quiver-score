package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/quiverscore/backend-go/internal/middleware"
	"github.com/quiverscore/backend-go/internal/repository"
)

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
