package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/quiverscore/backend-go/internal/config"
	"github.com/quiverscore/backend-go/internal/middleware"
	"github.com/quiverscore/backend-go/internal/repository"
)

type ClassificationsHandler struct {
	Classifications *repository.ClassificationRepo
	Cfg             *config.Config
}

func (h *ClassificationsHandler) Routes(r chi.Router) {
	r.Use(middleware.RequireAuth(h.Cfg.SecretKey))
	r.Get("/", h.List)
	r.Get("/current", h.Current)
}

func (h *ClassificationsHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	items, err := h.Classifications.List(r.Context(), userID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	JSON(w, http.StatusOK, items)
}

func (h *ClassificationsHandler) Current(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	items, err := h.Classifications.Current(r.Context(), userID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	JSON(w, http.StatusOK, items)
}
