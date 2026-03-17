package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/quiverscore/backend-go/internal/config"
	"github.com/quiverscore/backend-go/internal/middleware"
)

type UsersHandler struct {
	DB  *pgxpool.Pool
	Cfg *config.Config
}

func (h *UsersHandler) Routes(r chi.Router) {
	r.Use(middleware.RequireAuth(h.Cfg.SecretKey))
	r.Get("/me", h.GetMe)
}

type userOut struct {
	ID            string     `json:"id"`
	Email         string     `json:"email"`
	Username      string     `json:"username"`
	DisplayName   *string    `json:"display_name"`
	BowType       *string    `json:"bow_type"`
	Classification *string   `json:"classification"`
	Bio           *string    `json:"bio"`
	Avatar        *string    `json:"avatar"`
	EmailVerified bool       `json:"email_verified"`
	ProfilePublic bool       `json:"profile_public"`
	CreatedAt     time.Time  `json:"created_at"`
}

func (h *UsersHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	var u userOut
	err := h.DB.QueryRow(r.Context(),
		`SELECT id, email, username, display_name, bow_type, classification,
		        bio, avatar, email_verified, profile_public, created_at
		 FROM users WHERE id = $1`, userID,
	).Scan(
		&u.ID, &u.Email, &u.Username, &u.DisplayName, &u.BowType, &u.Classification,
		&u.Bio, &u.Avatar, &u.EmailVerified, &u.ProfilePublic, &u.CreatedAt,
	)
	if err != nil {
		Error(w, http.StatusUnauthorized, "User not found")
		return
	}

	JSON(w, http.StatusOK, u)
}
