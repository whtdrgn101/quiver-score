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

	"github.com/quiverscore/backend-go/internal/auth"
	"github.com/quiverscore/backend-go/internal/config"
	"github.com/quiverscore/backend-go/internal/middleware"
)

type AuthUserRepository interface {
	ExistsByEmailOrUsername(ctx context.Context, email, username string) (bool, error)
	Create(ctx context.Context, id, email, username, hashedPw, displayName, verificationToken string) error
	GetCredentialsByUsername(ctx context.Context, username string) (userID, hashedPw string, err error)
	Exists(ctx context.Context, id string) (bool, error)
	VerifyEmail(ctx context.Context, email string) (bool, error)
	GetEmailInfo(ctx context.Context, userID string) (email string, verified bool, err error)
	UpdateVerificationToken(ctx context.Context, userID, token string) error
	GetHashedPassword(ctx context.Context, userID string) (string, error)
	UpdatePassword(ctx context.Context, userID, hashedPw string) error
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	ResetPasswordByEmail(ctx context.Context, email, hashedPw string) (bool, error)
	DeleteUserData(ctx context.Context, userID string) error
}

type AuthEmailSender interface {
	SendVerificationEmail(toEmail, token, frontendURL string, expireHours int) error
	SendPasswordResetEmail(toEmail, token, frontendURL string, expireMinutes int) error
}

type AuthHandler struct {
	Users AuthUserRepository
	Email AuthEmailSender
	Cfg   *config.Config
}

func (h *AuthHandler) Routes(r chi.Router) {
	r.Post("/register", h.Register)
	r.Post("/login", h.Login)
	r.Post("/refresh", h.Refresh)
	r.Post("/verify-email", h.VerifyEmail)
	r.Post("/forgot-password", h.ForgotPassword)
	r.Post("/reset-password", h.ResetPassword)

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

	exists, err := h.Users.ExistsByEmailOrUsername(ctx, req.Email, req.Username)
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
	err = h.Users.Create(ctx, userID, req.Email, req.Username, hashedPw, displayName, verificationToken)
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

	// Send verification email (fire and forget)
	go h.Email.SendVerificationEmail(req.Email, verificationToken, h.Cfg.FrontendURL, h.Cfg.EmailVerificationTokenExpireHours)

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

	userID, hashedPw, err := h.Users.GetCredentialsByUsername(r.Context(), req.Username)
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

	exists, err := h.Users.Exists(r.Context(), claims.Subject)
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

	ok, err := h.Users.VerifyEmail(r.Context(), email)
	if err != nil || !ok {
		Error(w, http.StatusUnauthorized, "Invalid or expired verification token")
		return
	}

	JSON(w, http.StatusOK, map[string]string{"detail": "Email verified successfully."})
}

// ── Resend Verification ────────────────────────────────────────────────

func (h *AuthHandler) ResendVerification(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	email, emailVerified, err := h.Users.GetEmailInfo(r.Context(), userID)
	if err != nil {
		Error(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	if emailVerified {
		JSON(w, http.StatusOK, map[string]string{"detail": "Email is already verified."})
		return
	}

	token, _ := auth.CreateEmailVerificationToken(email, h.Cfg.EmailVerificationTokenExpireHours, h.Cfg.SecretKey)
	h.Users.UpdateVerificationToken(r.Context(), userID, token)

	go h.Email.SendVerificationEmail(email, token, h.Cfg.FrontendURL, h.Cfg.EmailVerificationTokenExpireHours)

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

	hashedPw, err := h.Users.GetHashedPassword(r.Context(), userID)
	if err != nil {
		Error(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	if !auth.VerifyPassword(req.CurrentPassword, hashedPw) {
		Error(w, http.StatusUnauthorized, "Current password is incorrect")
		return
	}

	newHash, _ := auth.HashPassword(req.NewPassword)
	h.Users.UpdatePassword(r.Context(), userID, newHash)

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

	// Always return success to avoid email enumeration
	JSON(w, http.StatusOK, map[string]string{
		"detail": "If that email is registered, you'll receive a reset link shortly.",
	})

	// Send reset email in background if user exists
	go func() {
		exists, _ := h.Users.ExistsByEmail(r.Context(), req.Email)
		if exists {
			token, _ := auth.CreateResetToken(req.Email, h.Cfg.PasswordResetTokenExpireMinutes, h.Cfg.SecretKey)
			h.Email.SendPasswordResetEmail(req.Email, token, h.Cfg.FrontendURL, h.Cfg.PasswordResetTokenExpireMinutes)
		}
	}()
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
	ok, err := h.Users.ResetPasswordByEmail(r.Context(), email, newHash)
	if err != nil || !ok {
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

	if err := h.Users.DeleteUserData(r.Context(), userID); err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	JSON(w, http.StatusOK, map[string]string{"detail": "Account and all data deleted"})
}
