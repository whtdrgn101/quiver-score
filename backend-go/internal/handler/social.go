package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/quiverscore/backend-go/internal/config"
	"github.com/quiverscore/backend-go/internal/middleware"
	"github.com/quiverscore/backend-go/internal/repository"
)

type SocialHandler struct {
	Social *repository.SocialRepo
	Cfg    *config.Config
}

func (h *SocialHandler) Routes(r chi.Router) {
	r.Use(middleware.RequireAuth(h.Cfg.SecretKey))
	r.Post("/follow/{userID}", h.Follow)
	r.Delete("/follow/{userID}", h.Unfollow)
	r.Get("/followers", h.ListFollowers)
	r.Get("/following", h.ListFollowing)
	r.Get("/feed", h.GetFeed)
}

func (h *SocialHandler) Follow(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	targetID := chi.URLParam(r, "userID")

	if userID == targetID {
		Error(w, http.StatusUnprocessableEntity, "Cannot follow yourself")
		return
	}

	follow, err := h.Social.Follow(r.Context(), userID, targetID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			Error(w, http.StatusNotFound, "User not found")
			return
		}
		if errors.Is(err, repository.ErrAlreadyMember) {
			Error(w, http.StatusConflict, "Already following this user")
			return
		}
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	JSON(w, http.StatusCreated, follow)
}

func (h *SocialHandler) Unfollow(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	targetID := chi.URLParam(r, "userID")

	err := h.Social.Unfollow(r.Context(), userID, targetID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			Error(w, http.StatusNotFound, "Not following this user")
			return
		}
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *SocialHandler) ListFollowers(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	followers, err := h.Social.ListFollowers(r.Context(), userID)
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSON(w, http.StatusOK, followers)
}

func (h *SocialHandler) ListFollowing(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	following, err := h.Social.ListFollowing(r.Context(), userID)
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSON(w, http.StatusOK, following)
}

func (h *SocialHandler) GetFeed(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	limit := 20
	offset := 0
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 1 && n <= 100 {
			limit = n
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}

	items, err := h.Social.GetFeed(r.Context(), userID, limit, offset)
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSON(w, http.StatusOK, items)
}
