package handler

import (
	"context"
	"log/slog"
	"net/http"
	"net/mail"
	"strings"
	"unicode/utf8"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/quiverscore/backend-go/internal/auth"
	"github.com/quiverscore/backend-go/internal/config"
	"github.com/quiverscore/backend-go/internal/middleware"
)

// AuthHandler holds dependencies for auth endpoints.
type AuthHandler struct {
	DB  *pgxpool.Pool
	Cfg *config.Config
}

// Routes mounts all auth routes on the given router.
func (h *AuthHandler) Routes(r chi.Router) {
	r.Post("/register", h.Register)
	r.Post("/login", h.Login)
	r.Post("/refresh", h.Refresh)
	r.Post("/verify-email", h.VerifyEmail)
	r.Post("/forgot-password", h.ForgotPassword)
	r.Post("/reset-password", h.ResetPassword)

	// These require authentication
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAuth(h.Cfg.SecretKey))
		r.Post("/resend-verification", h.ResendVerification)
		r.Post("/change-password", h.ChangePassword)
		r.Post("/delete-account", h.DeleteAccount)
	})
}

// ── Register ───────────────────────────────────────────────────────────

type registerRequest struct {
	Email       string  `json:"email"`
	Username    string  `json:"username"`
	Password    string  `json:"password"`
	DisplayName *string `json:"display_name"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if !Decode(w, r, &req) {
		return
	}

	// Validation
	if _, err := mail.ParseAddress(req.Email); err != nil || req.Email == "" {
		ValidationError(w, "Invalid email address")
		return
	}
	if utf8.RuneCountInString(req.Username) < 3 {
		ValidationError(w, "String should have at least 3 characters")
		return
	}
	if utf8.RuneCountInString(req.Password) < 8 {
		ValidationError(w, "String should have at least 8 characters")
		return
	}

	ctx := r.Context()

	// Check for existing user
	var exists bool
	err := h.DB.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 OR username = $2)",
		req.Email, req.Username,
	).Scan(&exists)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	if exists {
		Error(w, http.StatusConflict, "Email or username already registered")
		return
	}

	hashedPw, err := auth.HashPassword(req.Password)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	verificationToken, err := auth.CreateEmailVerificationToken(req.Email, h.Cfg.EmailVerificationTokenExpireHours, h.Cfg.SecretKey)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	displayName := req.Username
	if req.DisplayName != nil {
		displayName = *req.DisplayName
	}

	userID := uuid.New().String()
	_, err = h.DB.Exec(ctx,
		`INSERT INTO users (id, email, username, hashed_password, display_name, email_verification_token)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		userID, req.Email, req.Username, hashedPw, displayName, verificationToken,
	)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			Error(w, http.StatusConflict, "Email or username already registered")
			return
		}
		slog.Error("insert user failed", "error", err)
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	accessToken, err := auth.CreateAccessToken(userID, h.Cfg.AccessTokenExpireMinutes, h.Cfg.SecretKey)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	refreshToken, err := auth.CreateRefreshToken(userID, h.Cfg.RefreshTokenExpireDays, h.Cfg.SecretKey)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	JSON(w, http.StatusCreated, map[string]string{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"token_type":    "bearer",
	})
}

// ── Login ──────────────────────────────────────────────────────────────

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if !Decode(w, r, &req) {
		return
	}

	var userID, hashedPw string
	err := h.DB.QueryRow(r.Context(),
		"SELECT id, hashed_password FROM users WHERE username = $1",
		req.Username,
	).Scan(&userID, &hashedPw)
	if err != nil {
		Error(w, http.StatusUnauthorized, "Invalid username or password")
		return
	}

	if !auth.VerifyPassword(req.Password, hashedPw) {
		Error(w, http.StatusUnauthorized, "Invalid username or password")
		return
	}

	accessToken, _ := auth.CreateAccessToken(userID, h.Cfg.AccessTokenExpireMinutes, h.Cfg.SecretKey)
	refreshToken, _ := auth.CreateRefreshToken(userID, h.Cfg.RefreshTokenExpireDays, h.Cfg.SecretKey)

	JSON(w, http.StatusOK, map[string]string{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"token_type":    "bearer",
	})
}

// ── Refresh ────────────────────────────────────────────────────────────

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if !Decode(w, r, &req) {
		return
	}

	claims, err := auth.DecodeToken(req.RefreshToken, h.Cfg.SecretKey)
	if err != nil || claims.Type != string(auth.TokenTypeRefresh) || claims.Subject == "" {
		Error(w, http.StatusUnauthorized, "Invalid refresh token")
		return
	}

	// Verify user still exists
	var exists bool
	err = h.DB.QueryRow(r.Context(),
		"SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)",
		claims.Subject,
	).Scan(&exists)
	if err != nil || !exists {
		Error(w, http.StatusUnauthorized, "User not found")
		return
	}

	accessToken, _ := auth.CreateAccessToken(claims.Subject, h.Cfg.AccessTokenExpireMinutes, h.Cfg.SecretKey)
	refreshToken, _ := auth.CreateRefreshToken(claims.Subject, h.Cfg.RefreshTokenExpireDays, h.Cfg.SecretKey)

	JSON(w, http.StatusOK, map[string]string{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"token_type":    "bearer",
	})
}

// ── Verify Email ───────────────────────────────────────────────────────

type verifyEmailRequest struct {
	Token string `json:"token"`
}

func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req verifyEmailRequest
	if !Decode(w, r, &req) {
		return
	}

	email, err := auth.VerifyEmailVerificationToken(req.Token, h.Cfg.SecretKey)
	if err != nil || email == "" {
		Error(w, http.StatusUnauthorized, "Invalid or expired verification token")
		return
	}

	tag, err := h.DB.Exec(r.Context(),
		`UPDATE users SET email_verified = true, email_verification_token = NULL
		 WHERE email = $1`,
		email,
	)
	if err != nil || tag.RowsAffected() == 0 {
		Error(w, http.StatusUnauthorized, "Invalid or expired verification token")
		return
	}

	JSON(w, http.StatusOK, map[string]string{"detail": "Email verified successfully."})
}

// ── Resend Verification ────────────────────────────────────────────────

func (h *AuthHandler) ResendVerification(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	var emailVerified bool
	var email string
	err := h.DB.QueryRow(r.Context(),
		"SELECT email, email_verified FROM users WHERE id = $1", userID,
	).Scan(&email, &emailVerified)
	if err != nil {
		Error(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	if emailVerified {
		JSON(w, http.StatusOK, map[string]string{"detail": "Email is already verified."})
		return
	}

	token, _ := auth.CreateEmailVerificationToken(email, h.Cfg.EmailVerificationTokenExpireHours, h.Cfg.SecretKey)
	h.DB.Exec(r.Context(),
		"UPDATE users SET email_verification_token = $1 WHERE id = $2",
		token, userID,
	)

	JSON(w, http.StatusOK, map[string]string{"detail": "Verification email sent."})
}

// ── Change Password ────────────────────────────────────────────────────

type changePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	var req changePasswordRequest
	if !Decode(w, r, &req) {
		return
	}

	userID := middleware.GetUserID(r.Context())

	var hashedPw string
	err := h.DB.QueryRow(r.Context(),
		"SELECT hashed_password FROM users WHERE id = $1", userID,
	).Scan(&hashedPw)
	if err != nil {
		Error(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	if !auth.VerifyPassword(req.CurrentPassword, hashedPw) {
		Error(w, http.StatusUnauthorized, "Current password is incorrect")
		return
	}

	newHash, _ := auth.HashPassword(req.NewPassword)
	h.DB.Exec(r.Context(),
		"UPDATE users SET hashed_password = $1 WHERE id = $2",
		newHash, userID,
	)

	JSON(w, http.StatusOK, map[string]string{"detail": "Password changed successfully"})
}

// ── Forgot Password ───────────────────────────────────────────────────

type forgotPasswordRequest struct {
	Email string `json:"email"`
}

func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req forgotPasswordRequest
	if !Decode(w, r, &req) {
		return
	}

	// Always return the same response to prevent email enumeration
	JSON(w, http.StatusOK, map[string]string{
		"detail": "If that email is registered, you'll receive a reset link shortly.",
	})
}

// ── Reset Password ─────────────────────────────────────────────────────

type resetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req resetPasswordRequest
	if !Decode(w, r, &req) {
		return
	}

	email, err := auth.VerifyResetToken(req.Token, h.Cfg.SecretKey)
	if err != nil || email == "" {
		Error(w, http.StatusUnauthorized, "Invalid or expired reset token")
		return
	}

	newHash, _ := auth.HashPassword(req.NewPassword)
	tag, err := h.DB.Exec(r.Context(),
		"UPDATE users SET hashed_password = $1 WHERE email = $2",
		newHash, email,
	)
	if err != nil || tag.RowsAffected() == 0 {
		Error(w, http.StatusUnauthorized, "Invalid or expired reset token")
		return
	}

	JSON(w, http.StatusOK, map[string]string{"detail": "Password reset successfully. You can now sign in."})
}

// ── Delete Account ─────────────────────────────────────────────────────

type deleteAccountRequest struct {
	Confirmation string `json:"confirmation"`
}

func (h *AuthHandler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	var req deleteAccountRequest
	if !Decode(w, r, &req) {
		return
	}

	if req.Confirmation != "Yes, delete ALL of my data" {
		Error(w, http.StatusUnauthorized, "Confirmation text does not match")
		return
	}

	userID := middleware.GetUserID(r.Context())
	ctx := r.Context()

	tx, err := h.DB.Begin(ctx)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	defer tx.Rollback(ctx)

	if err := deleteUserData(ctx, tx, userID); err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	if err := tx.Commit(ctx); err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	JSON(w, http.StatusOK, map[string]string{"detail": "Account and all data deleted"})
}

// deleteUserData removes all data associated with a user, mirroring the Python cascade.
func deleteUserData(ctx context.Context, tx pgx.Tx, userID string) error {
	// 1. Scoring data: arrows → ends → sessions
	sessionIDs, err := collectIDs(ctx, tx, "SELECT id FROM scoring_sessions WHERE user_id = $1", userID)
	if err != nil {
		return err
	}
	if len(sessionIDs) > 0 {
		endIDs, err := collectIDs(ctx, tx, "SELECT id FROM ends WHERE session_id = ANY($1)", sessionIDs)
		if err != nil {
			return err
		}
		if len(endIDs) > 0 {
			if _, err := tx.Exec(ctx, "DELETE FROM arrows WHERE end_id = ANY($1)", endIDs); err != nil {
				return err
			}
		}
		if _, err := tx.Exec(ctx, "DELETE FROM ends WHERE session_id = ANY($1)", sessionIDs); err != nil {
			return err
		}
	}

	// Session annotations
	if len(sessionIDs) > 0 {
		if _, err := tx.Exec(ctx,
			"DELETE FROM session_annotations WHERE author_id = $1 OR session_id = ANY($2)",
			userID, sessionIDs); err != nil {
			return err
		}
	} else {
		if _, err := tx.Exec(ctx, "DELETE FROM session_annotations WHERE author_id = $1", userID); err != nil {
			return err
		}
	}

	// Tournament participants
	if _, err := tx.Exec(ctx, "DELETE FROM tournament_participants WHERE user_id = $1", userID); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, "DELETE FROM personal_records WHERE user_id = $1", userID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, "DELETE FROM scoring_sessions WHERE user_id = $1", userID); err != nil {
		return err
	}

	// 2. Equipment and setups
	setupIDs, err := collectIDs(ctx, tx, "SELECT id FROM setup_profiles WHERE user_id = $1", userID)
	if err != nil {
		return err
	}
	if len(setupIDs) > 0 {
		if _, err := tx.Exec(ctx, "DELETE FROM setup_equipment WHERE setup_id = ANY($1)", setupIDs); err != nil {
			return err
		}
	}
	if _, err := tx.Exec(ctx, "DELETE FROM setup_profiles WHERE user_id = $1", userID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, "DELETE FROM equipment WHERE user_id = $1", userID); err != nil {
		return err
	}

	// 3. Coaching
	if _, err := tx.Exec(ctx, "DELETE FROM coach_athlete_links WHERE coach_id = $1 OR athlete_id = $1", userID); err != nil {
		return err
	}

	// 4. Clubs owned by user — delete entire club cascade
	ownedClubIDs, err := collectIDs(ctx, tx, "SELECT id FROM clubs WHERE owner_id = $1", userID)
	if err != nil {
		return err
	}
	if len(ownedClubIDs) > 0 {
		// Tournament participants for club tournaments
		tournamentIDs, err := collectIDs(ctx, tx, "SELECT id FROM tournaments WHERE club_id = ANY($1)", ownedClubIDs)
		if err != nil {
			return err
		}
		if len(tournamentIDs) > 0 {
			if _, err := tx.Exec(ctx, "DELETE FROM tournament_participants WHERE tournament_id = ANY($1)", tournamentIDs); err != nil {
				return err
			}
		}
		if _, err := tx.Exec(ctx, "DELETE FROM tournaments WHERE club_id = ANY($1)", ownedClubIDs); err != nil {
			return err
		}

		// Club events and participants
		eventIDs, err := collectIDs(ctx, tx, "SELECT id FROM club_events WHERE club_id = ANY($1)", ownedClubIDs)
		if err != nil {
			return err
		}
		if len(eventIDs) > 0 {
			if _, err := tx.Exec(ctx, "DELETE FROM club_event_participants WHERE event_id = ANY($1)", eventIDs); err != nil {
				return err
			}
		}
		if _, err := tx.Exec(ctx, "DELETE FROM club_events WHERE club_id = ANY($1)", ownedClubIDs); err != nil {
			return err
		}

		// Teams and team members
		teamIDs, err := collectIDs(ctx, tx, "SELECT id FROM club_teams WHERE club_id = ANY($1)", ownedClubIDs)
		if err != nil {
			return err
		}
		if len(teamIDs) > 0 {
			if _, err := tx.Exec(ctx, "DELETE FROM club_team_members WHERE team_id = ANY($1)", teamIDs); err != nil {
				return err
			}
		}
		if _, err := tx.Exec(ctx, "DELETE FROM club_teams WHERE club_id = ANY($1)", ownedClubIDs); err != nil {
			return err
		}

		if _, err := tx.Exec(ctx, "DELETE FROM club_shared_rounds WHERE club_id = ANY($1)", ownedClubIDs); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, "DELETE FROM club_invites WHERE club_id = ANY($1)", ownedClubIDs); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, "DELETE FROM club_members WHERE club_id = ANY($1)", ownedClubIDs); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, "DELETE FROM clubs WHERE id = ANY($1)", ownedClubIDs); err != nil {
			return err
		}
	}

	// 5. Tournaments organized by user (not in owned clubs)
	userTournamentIDs, err := collectIDs(ctx, tx, "SELECT id FROM tournaments WHERE organizer_id = $1", userID)
	if err != nil {
		return err
	}
	if len(userTournamentIDs) > 0 {
		if _, err := tx.Exec(ctx, "DELETE FROM tournament_participants WHERE tournament_id = ANY($1)", userTournamentIDs); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, "DELETE FROM tournaments WHERE id = ANY($1)", userTournamentIDs); err != nil {
			return err
		}
	}

	// 6. Club participation (non-owned clubs)
	if _, err := tx.Exec(ctx, "DELETE FROM club_event_participants WHERE user_id = $1", userID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, "DELETE FROM club_team_members WHERE user_id = $1", userID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, "DELETE FROM club_shared_rounds WHERE shared_by = $1", userID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, "DELETE FROM club_invites WHERE created_by = $1", userID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, "DELETE FROM club_members WHERE user_id = $1", userID); err != nil {
		return err
	}

	// 7. Custom round templates and stages
	templateIDs, err := collectIDs(ctx, tx,
		"SELECT id FROM round_templates WHERE created_by = $1 AND is_official = false", userID)
	if err != nil {
		return err
	}
	if len(templateIDs) > 0 {
		if _, err := tx.Exec(ctx, "DELETE FROM club_shared_rounds WHERE template_id = ANY($1)", templateIDs); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, "DELETE FROM round_template_stages WHERE template_id = ANY($1)", templateIDs); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, "DELETE FROM round_templates WHERE id = ANY($1)", templateIDs); err != nil {
			return err
		}
	}

	// 8. Remaining user data
	for _, table := range []string{
		"notifications",
		"classification_records",
		"sight_marks",
		"feed_items",
	} {
		if _, err := tx.Exec(ctx, "DELETE FROM "+table+" WHERE user_id = $1", userID); err != nil {
			return err
		}
	}
	if _, err := tx.Exec(ctx, "DELETE FROM follows WHERE follower_id = $1 OR following_id = $1", userID); err != nil {
		return err
	}

	// 9. Delete user
	if _, err := tx.Exec(ctx, "DELETE FROM users WHERE id = $1", userID); err != nil {
		return err
	}

	return nil
}

// collectIDs executes a query that returns a single UUID column and collects the results.
func collectIDs(ctx context.Context, tx pgx.Tx, query string, args ...any) ([]string, error) {
	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
