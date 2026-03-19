package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/quiverscore/backend-go/internal/auth"
	"github.com/quiverscore/backend-go/internal/config"
	"github.com/quiverscore/backend-go/internal/middleware"
)

// ── Mock User Repository ─────────────────────────────────────────────

type mockAuthUserRepo struct {
	existsByEmailOrUsernameResult bool
	existsByEmailOrUsernameErr    error

	createErr error

	getCredUserID   string
	getCredHashedPw string
	getCredErr      error

	existsResult bool
	existsErr    error

	verifyEmailResult bool
	verifyEmailErr    error

	getEmailInfoEmail    string
	getEmailInfoVerified bool
	getEmailInfoErr      error

	updateVerificationTokenErr error

	getHashedPasswordResult string
	getHashedPasswordErr    error

	updatePasswordErr error

	existsByEmailResult bool
	existsByEmailErr    error

	resetPasswordByEmailResult bool
	resetPasswordByEmailErr    error

	deleteUserDataErr error
}

func (m *mockAuthUserRepo) ExistsByEmailOrUsername(_ context.Context, _, _ string) (bool, error) {
	return m.existsByEmailOrUsernameResult, m.existsByEmailOrUsernameErr
}

func (m *mockAuthUserRepo) Create(_ context.Context, _, _, _, _, _, _ string) error {
	return m.createErr
}

func (m *mockAuthUserRepo) GetCredentialsByUsername(_ context.Context, _ string) (string, string, error) {
	return m.getCredUserID, m.getCredHashedPw, m.getCredErr
}

func (m *mockAuthUserRepo) Exists(_ context.Context, _ string) (bool, error) {
	return m.existsResult, m.existsErr
}

func (m *mockAuthUserRepo) VerifyEmail(_ context.Context, _ string) (bool, error) {
	return m.verifyEmailResult, m.verifyEmailErr
}

func (m *mockAuthUserRepo) GetEmailInfo(_ context.Context, _ string) (string, bool, error) {
	return m.getEmailInfoEmail, m.getEmailInfoVerified, m.getEmailInfoErr
}

func (m *mockAuthUserRepo) UpdateVerificationToken(_ context.Context, _, _ string) error {
	return m.updateVerificationTokenErr
}

func (m *mockAuthUserRepo) GetHashedPassword(_ context.Context, _ string) (string, error) {
	return m.getHashedPasswordResult, m.getHashedPasswordErr
}

func (m *mockAuthUserRepo) UpdatePassword(_ context.Context, _, _ string) error {
	return m.updatePasswordErr
}

func (m *mockAuthUserRepo) ExistsByEmail(_ context.Context, _ string) (bool, error) {
	return m.existsByEmailResult, m.existsByEmailErr
}

func (m *mockAuthUserRepo) ResetPasswordByEmail(_ context.Context, _, _ string) (bool, error) {
	return m.resetPasswordByEmailResult, m.resetPasswordByEmailErr
}

func (m *mockAuthUserRepo) DeleteUserData(_ context.Context, _ string) error {
	return m.deleteUserDataErr
}

// ── Mock Email Sender ────────────────────────────────────────────────

type mockAuthEmailSender struct{}

func (m *mockAuthEmailSender) SendVerificationEmail(_, _, _ string, _ int) error { return nil }
func (m *mockAuthEmailSender) SendPasswordResetEmail(_, _, _ string, _ int) error {
	return nil
}

// ── Test Config ──────────────────────────────────────────────────────

var testCfg = &config.Config{
	SecretKey:                        "test-secret",
	AccessTokenExpireMinutes:         30,
	RefreshTokenExpireDays:           7,
	EmailVerificationTokenExpireHours: 24,
	PasswordResetTokenExpireMinutes:  30,
	FrontendURL:                      "https://test.example.com",
}

// ── Helper ───────────────────────────────────────────────────────────

func parseDetail(t *testing.T, rr *httptest.ResponseRecorder) string {
	t.Helper()
	var resp map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	return resp["detail"]
}

// ── Register Tests ───────────────────────────────────────────────────

func TestRegister_Success(t *testing.T) {
	mock := &mockAuthUserRepo{
		existsByEmailOrUsernameResult: false,
	}
	h := &AuthHandler{Users: mock, Email: &mockAuthEmailSender{}, Cfg: testCfg}

	body := strings.NewReader(`{"email":"user@example.com","username":"testuser","password":"password123"}`)
	req := httptest.NewRequest(http.MethodPost, "/register", body)
	rr := httptest.NewRecorder()
	h.Register(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["access_token"] == "" {
		t.Error("expected access_token in response")
	}
	if resp["refresh_token"] == "" {
		t.Error("expected refresh_token in response")
	}
	if resp["token_type"] != "bearer" {
		t.Errorf("expected token_type 'bearer', got '%s'", resp["token_type"])
	}
}

func TestRegister_ShortUsername(t *testing.T) {
	h := &AuthHandler{Users: &mockAuthUserRepo{}, Email: &mockAuthEmailSender{}, Cfg: testCfg}

	body := strings.NewReader(`{"email":"user@example.com","username":"ab","password":"password123"}`)
	req := httptest.NewRequest(http.MethodPost, "/register", body)
	rr := httptest.NewRecorder()
	h.Register(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", rr.Code)
	}
}

func TestRegister_ShortPassword(t *testing.T) {
	h := &AuthHandler{Users: &mockAuthUserRepo{}, Email: &mockAuthEmailSender{}, Cfg: testCfg}

	body := strings.NewReader(`{"email":"user@example.com","username":"testuser","password":"short"}`)
	req := httptest.NewRequest(http.MethodPost, "/register", body)
	rr := httptest.NewRecorder()
	h.Register(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", rr.Code)
	}
}

func TestRegister_InvalidEmail(t *testing.T) {
	h := &AuthHandler{Users: &mockAuthUserRepo{}, Email: &mockAuthEmailSender{}, Cfg: testCfg}

	body := strings.NewReader(`{"email":"not-an-email","username":"testuser","password":"password123"}`)
	req := httptest.NewRequest(http.MethodPost, "/register", body)
	rr := httptest.NewRecorder()
	h.Register(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", rr.Code)
	}
}

func TestRegister_Duplicate(t *testing.T) {
	mock := &mockAuthUserRepo{
		existsByEmailOrUsernameResult: true,
	}
	h := &AuthHandler{Users: mock, Email: &mockAuthEmailSender{}, Cfg: testCfg}

	body := strings.NewReader(`{"email":"user@example.com","username":"testuser","password":"password123"}`)
	req := httptest.NewRequest(http.MethodPost, "/register", body)
	rr := httptest.NewRecorder()
	h.Register(rr, req)

	if rr.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d", rr.Code)
	}
}

// ── Login Tests ──────────────────────────────────────────────────────

func TestLogin_Success(t *testing.T) {
	hashedPw, err := auth.HashPassword("password123")
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	mock := &mockAuthUserRepo{
		getCredUserID:   "user-1",
		getCredHashedPw: hashedPw,
	}
	h := &AuthHandler{Users: mock, Email: &mockAuthEmailSender{}, Cfg: testCfg}

	body := strings.NewReader(`{"username":"testuser","password":"password123"}`)
	req := httptest.NewRequest(http.MethodPost, "/login", body)
	rr := httptest.NewRecorder()
	h.Login(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["access_token"] == "" {
		t.Error("expected access_token in response")
	}
	if resp["refresh_token"] == "" {
		t.Error("expected refresh_token in response")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	hashedPw, _ := auth.HashPassword("password123")
	mock := &mockAuthUserRepo{
		getCredUserID:   "user-1",
		getCredHashedPw: hashedPw,
	}
	h := &AuthHandler{Users: mock, Email: &mockAuthEmailSender{}, Cfg: testCfg}

	body := strings.NewReader(`{"username":"testuser","password":"wrongpassword"}`)
	req := httptest.NewRequest(http.MethodPost, "/login", body)
	rr := httptest.NewRecorder()
	h.Login(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestLogin_NonexistentUser(t *testing.T) {
	mock := &mockAuthUserRepo{
		getCredErr: errors.New("no rows"),
	}
	h := &AuthHandler{Users: mock, Email: &mockAuthEmailSender{}, Cfg: testCfg}

	body := strings.NewReader(`{"username":"noone","password":"password123"}`)
	req := httptest.NewRequest(http.MethodPost, "/login", body)
	rr := httptest.NewRecorder()
	h.Login(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

// ── Refresh Tests ────────────────────────────────────────────────────

func TestRefresh_Success(t *testing.T) {
	refreshToken, err := auth.CreateRefreshToken("user-1", testCfg.RefreshTokenExpireDays, testCfg.SecretKey)
	if err != nil {
		t.Fatalf("failed to create refresh token: %v", err)
	}
	mock := &mockAuthUserRepo{
		existsResult: true,
	}
	h := &AuthHandler{Users: mock, Email: &mockAuthEmailSender{}, Cfg: testCfg}

	body := strings.NewReader(`{"refresh_token":"` + refreshToken + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/refresh", body)
	rr := httptest.NewRecorder()
	h.Refresh(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["access_token"] == "" {
		t.Error("expected access_token in response")
	}
}

func TestRefresh_WithAccessToken(t *testing.T) {
	accessToken, _ := auth.CreateAccessToken("user-1", testCfg.AccessTokenExpireMinutes, testCfg.SecretKey)
	h := &AuthHandler{Users: &mockAuthUserRepo{}, Email: &mockAuthEmailSender{}, Cfg: testCfg}

	body := strings.NewReader(`{"refresh_token":"` + accessToken + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/refresh", body)
	rr := httptest.NewRecorder()
	h.Refresh(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestRefresh_InvalidToken(t *testing.T) {
	h := &AuthHandler{Users: &mockAuthUserRepo{}, Email: &mockAuthEmailSender{}, Cfg: testCfg}

	body := strings.NewReader(`{"refresh_token":"invalid-token"}`)
	req := httptest.NewRequest(http.MethodPost, "/refresh", body)
	rr := httptest.NewRecorder()
	h.Refresh(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

// ── VerifyEmail Tests ────────────────────────────────────────────────

func TestVerifyEmail_Success(t *testing.T) {
	token, err := auth.CreateEmailVerificationToken("user@example.com", testCfg.EmailVerificationTokenExpireHours, testCfg.SecretKey)
	if err != nil {
		t.Fatalf("failed to create verification token: %v", err)
	}
	mock := &mockAuthUserRepo{
		verifyEmailResult: true,
	}
	h := &AuthHandler{Users: mock, Email: &mockAuthEmailSender{}, Cfg: testCfg}

	body := strings.NewReader(`{"token":"` + token + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/verify-email", body)
	rr := httptest.NewRecorder()
	h.VerifyEmail(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestVerifyEmail_InvalidToken(t *testing.T) {
	h := &AuthHandler{Users: &mockAuthUserRepo{}, Email: &mockAuthEmailSender{}, Cfg: testCfg}

	body := strings.NewReader(`{"token":"invalid-token"}`)
	req := httptest.NewRequest(http.MethodPost, "/verify-email", body)
	rr := httptest.NewRecorder()
	h.VerifyEmail(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

// ── ChangePassword Tests ─────────────────────────────────────────────

func TestChangePassword_Success(t *testing.T) {
	hashedPw, _ := auth.HashPassword("oldpassword")
	mock := &mockAuthUserRepo{
		getHashedPasswordResult: hashedPw,
	}
	h := &AuthHandler{Users: mock, Email: &mockAuthEmailSender{}, Cfg: testCfg}

	body := strings.NewReader(`{"current_password":"oldpassword","new_password":"newpassword123"}`)
	req := httptest.NewRequest(http.MethodPost, "/change-password", body)
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, "user-1"))
	rr := httptest.NewRecorder()
	h.ChangePassword(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestChangePassword_WrongCurrent(t *testing.T) {
	hashedPw, _ := auth.HashPassword("oldpassword")
	mock := &mockAuthUserRepo{
		getHashedPasswordResult: hashedPw,
	}
	h := &AuthHandler{Users: mock, Email: &mockAuthEmailSender{}, Cfg: testCfg}

	body := strings.NewReader(`{"current_password":"wrongpassword","new_password":"newpassword123"}`)
	req := httptest.NewRequest(http.MethodPost, "/change-password", body)
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, "user-1"))
	rr := httptest.NewRecorder()
	h.ChangePassword(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

// ── ForgotPassword Tests ─────────────────────────────────────────────

func TestForgotPassword_AlwaysReturns200(t *testing.T) {
	h := &AuthHandler{Users: &mockAuthUserRepo{}, Email: &mockAuthEmailSender{}, Cfg: testCfg}

	body := strings.NewReader(`{"email":"anyone@example.com"}`)
	req := httptest.NewRequest(http.MethodPost, "/forgot-password", body)
	rr := httptest.NewRecorder()
	h.ForgotPassword(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

// ── ResetPassword Tests ──────────────────────────────────────────────

func TestResetPassword_Success(t *testing.T) {
	token, err := auth.CreateResetToken("user@example.com", testCfg.PasswordResetTokenExpireMinutes, testCfg.SecretKey)
	if err != nil {
		t.Fatalf("failed to create reset token: %v", err)
	}
	mock := &mockAuthUserRepo{
		resetPasswordByEmailResult: true,
	}
	h := &AuthHandler{Users: mock, Email: &mockAuthEmailSender{}, Cfg: testCfg}

	body := strings.NewReader(`{"token":"` + token + `","new_password":"newpassword123"}`)
	req := httptest.NewRequest(http.MethodPost, "/reset-password", body)
	rr := httptest.NewRecorder()
	h.ResetPassword(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestResetPassword_InvalidToken(t *testing.T) {
	h := &AuthHandler{Users: &mockAuthUserRepo{}, Email: &mockAuthEmailSender{}, Cfg: testCfg}

	body := strings.NewReader(`{"token":"invalid-token","new_password":"newpassword123"}`)
	req := httptest.NewRequest(http.MethodPost, "/reset-password", body)
	rr := httptest.NewRecorder()
	h.ResetPassword(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

// ── DeleteAccount Tests ──────────────────────────────────────────────

func TestDeleteAccount_Success(t *testing.T) {
	mock := &mockAuthUserRepo{}
	h := &AuthHandler{Users: mock, Email: &mockAuthEmailSender{}, Cfg: testCfg}

	body := strings.NewReader(`{"confirmation":"Yes, delete ALL of my data"}`)
	req := httptest.NewRequest(http.MethodPost, "/delete-account", body)
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, "user-1"))
	rr := httptest.NewRecorder()
	h.DeleteAccount(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestDeleteAccount_WrongConfirmation(t *testing.T) {
	h := &AuthHandler{Users: &mockAuthUserRepo{}, Email: &mockAuthEmailSender{}, Cfg: testCfg}

	body := strings.NewReader(`{"confirmation":"wrong text"}`)
	req := httptest.NewRequest(http.MethodPost, "/delete-account", body)
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, "user-1"))
	rr := httptest.NewRecorder()
	h.DeleteAccount(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}
