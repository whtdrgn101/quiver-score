package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/quiverscore/backend-go/internal/config"
	"github.com/quiverscore/backend-go/internal/middleware"
	"github.com/quiverscore/backend-go/internal/repository"
)

type NotificationRepository interface {
	List(ctx context.Context, userID string) ([]repository.NotificationOut, error)
	UnreadCount(ctx context.Context, userID string) (int, error)
	MarkRead(ctx context.Context, userID, notificationID string) (*repository.NotificationOut, error)
	MarkAllRead(ctx context.Context, userID string) error
}

type NotificationsHandler struct {
	Notifications NotificationRepository
	Cfg           *config.Config
}

func (h *NotificationsHandler) Routes(r chi.Router) {
	r.Use(middleware.RequireAuth(h.Cfg.SecretKey))
	r.Get("/", h.List)
	r.Get("/unread-count", h.UnreadCount)
	r.Patch("/{notificationID}/read", h.MarkRead)
	r.Post("/read-all", h.MarkAllRead)
}

func (h *NotificationsHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	notifications, err := h.Notifications.List(r.Context(), userID)
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSON(w, http.StatusOK, notifications)
}

func (h *NotificationsHandler) UnreadCount(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	count, err := h.Notifications.UnreadCount(r.Context(), userID)
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSON(w, http.StatusOK, map[string]int{"count": count})
}

func (h *NotificationsHandler) MarkRead(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	notificationID := chi.URLParam(r, "notificationID")

	notification, err := h.Notifications.MarkRead(r.Context(), userID, notificationID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			Error(w, http.StatusNotFound, "Notification not found")
			return
		}
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSON(w, http.StatusOK, notification)
}

func (h *NotificationsHandler) MarkAllRead(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	err := h.Notifications.MarkAllRead(r.Context(), userID)
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSON(w, http.StatusOK, map[string]string{"message": "All notifications marked as read"})
}
