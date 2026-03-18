package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	DatabaseURL                      string
	SecretKey                        string
	Port                             string
	CORSOrigins                      []string
	RateLimitEnabled                 bool
	AccessTokenExpireMinutes         int
	RefreshTokenExpireDays           int
	Algorithm                        string
	SendGridAPIKey                   string
	SendGridFromEmail                string
	FrontendURL                      string
	PasswordResetTokenExpireMinutes  int
	EmailVerificationTokenExpireHours int
}

func Load() *Config {
	return &Config{
		DatabaseURL:                      envOrDefault("DATABASE_URL", "postgres://quiverscore:quiverscore@localhost:5432/quiverscore"),
		SecretKey:                        envOrDefault("SECRET_KEY", "dev-secret-key-change-in-production"),
		Port:                             envOrDefault("PORT", "8080"),
		CORSOrigins:                      strings.Split(envOrDefault("CORS_ORIGINS", "http://localhost:5173"), ","),
		RateLimitEnabled:                 envOrDefaultBool("RATE_LIMIT_ENABLED", true),
		AccessTokenExpireMinutes:         envOrDefaultInt("ACCESS_TOKEN_EXPIRE_MINUTES", 15),
		RefreshTokenExpireDays:           envOrDefaultInt("REFRESH_TOKEN_EXPIRE_DAYS", 30),
		Algorithm:                        envOrDefault("ALGORITHM", "HS256"),
		SendGridAPIKey:                   envOrDefault("SENDGRID_API_KEY", ""),
		SendGridFromEmail:                envOrDefault("SENDGRID_FROM_EMAIL", "noreply@quiverscore.com"),
		FrontendURL:                      envOrDefault("FRONTEND_URL", "http://localhost:5173"),
		PasswordResetTokenExpireMinutes:  envOrDefaultInt("PASSWORD_RESET_TOKEN_EXPIRE_MINUTES", 60),
		EmailVerificationTokenExpireHours: envOrDefaultInt("EMAIL_VERIFICATION_TOKEN_EXPIRE_HOURS", 24),
	}
}

// NormalizeDatabaseURL converts a SQLAlchemy-style URL to a standard postgres URL.
// Python uses "postgresql+asyncpg://..." but Go pgx needs "postgres://...".
func (c *Config) NormalizeDatabaseURL() string {
	url := c.DatabaseURL
	url = strings.Replace(url, "postgresql+asyncpg://", "postgres://", 1)
	url = strings.Replace(url, "postgresql://", "postgres://", 1)
	return url
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envOrDefaultInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func envOrDefaultBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return fallback
}
