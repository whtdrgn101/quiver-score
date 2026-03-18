package handler

import (
	"net/http"

	"github.com/quiverscore/backend-go/internal/config"
	"github.com/quiverscore/backend-go/internal/middleware"
	"github.com/quiverscore/backend-go/internal/repository"
)

type UsersHandler struct {
	Users *repository.UserRepo
	Cfg   *config.Config
}

func (h *UsersHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	u, err := h.Users.GetMe(r.Context(), userID)
	if err != nil {
		Error(w, http.StatusUnauthorized, "User not found")
		return
	}

	JSON(w, http.StatusOK, u)
}
