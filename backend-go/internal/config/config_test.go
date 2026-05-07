package config

import (
	"os"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	// Clear env vars that would override defaults
	for _, key := range []string{
		"ACCESS_TOKEN_EXPIRE_MINUTES",
		"REFRESH_TOKEN_EXPIRE_DAYS",
		"PASSWORD_RESET_TOKEN_EXPIRE_MINUTES",
		"EMAIL_VERIFICATION_TOKEN_EXPIRE_HOURS",
	} {
		os.Unsetenv(key)
	}

	cfg := Load()

	if cfg.AccessTokenExpireMinutes != 15 {
		t.Errorf("AccessTokenExpireMinutes = %d, want 15", cfg.AccessTokenExpireMinutes)
	}
	if cfg.RefreshTokenExpireDays != 90 {
		t.Errorf("RefreshTokenExpireDays = %d, want 90", cfg.RefreshTokenExpireDays)
	}
	if cfg.PasswordResetTokenExpireMinutes != 60 {
		t.Errorf("PasswordResetTokenExpireMinutes = %d, want 60", cfg.PasswordResetTokenExpireMinutes)
	}
	if cfg.EmailVerificationTokenExpireHours != 24 {
		t.Errorf("EmailVerificationTokenExpireHours = %d, want 24", cfg.EmailVerificationTokenExpireHours)
	}
}

func TestLoadRefreshTokenFromEnv(t *testing.T) {
	os.Setenv("REFRESH_TOKEN_EXPIRE_DAYS", "180")
	defer os.Unsetenv("REFRESH_TOKEN_EXPIRE_DAYS")

	cfg := Load()

	if cfg.RefreshTokenExpireDays != 180 {
		t.Errorf("RefreshTokenExpireDays = %d, want 180", cfg.RefreshTokenExpireDays)
	}
}

func TestNormalizeDatabaseURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"asyncpg format", "postgresql+asyncpg://user:pass@host/db", "postgres://user:pass@host/db"},
		{"postgresql format", "postgresql://user:pass@host/db", "postgres://user:pass@host/db"},
		{"already postgres", "postgres://user:pass@host/db", "postgres://user:pass@host/db"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{DatabaseURL: tt.input}
			got := cfg.NormalizeDatabaseURL()
			if got != tt.expected {
				t.Errorf("NormalizeDatabaseURL() = %q, want %q", got, tt.expected)
			}
		})
	}
}
