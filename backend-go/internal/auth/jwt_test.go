package auth

import (
	"testing"
	"time"
)

const testSecret = "dev-secret-key-change-in-production"

func TestCreateAndDecodeAccessToken(t *testing.T) {
	userID := "550e8400-e29b-41d4-a716-446655440000"
	token, err := CreateAccessToken(userID, 15, testSecret)
	if err != nil {
		t.Fatalf("create access token: %v", err)
	}

	claims, err := DecodeToken(token, testSecret)
	if err != nil {
		t.Fatalf("decode token: %v", err)
	}

	if claims.Subject != userID {
		t.Errorf("subject = %q, want %q", claims.Subject, userID)
	}
	if claims.Type != "access" {
		t.Errorf("type = %q, want %q", claims.Type, "access")
	}
}

func TestCreateAndDecodeRefreshToken(t *testing.T) {
	userID := "550e8400-e29b-41d4-a716-446655440000"
	token, err := CreateRefreshToken(userID, 30, testSecret)
	if err != nil {
		t.Fatalf("create refresh token: %v", err)
	}

	claims, err := DecodeToken(token, testSecret)
	if err != nil {
		t.Fatalf("decode token: %v", err)
	}

	if claims.Subject != userID {
		t.Errorf("subject = %q, want %q", claims.Subject, userID)
	}
	if claims.Type != "refresh" {
		t.Errorf("type = %q, want %q", claims.Type, "refresh")
	}
}

func TestDecodeExpiredToken(t *testing.T) {
	token, err := CreateToken("test", TokenTypeAccess, -1*time.Second, testSecret)
	if err != nil {
		t.Fatalf("create token: %v", err)
	}

	_, err = DecodeToken(token, testSecret)
	if err == nil {
		t.Error("expected error for expired token, got nil")
	}
}

func TestDecodeInvalidToken(t *testing.T) {
	_, err := DecodeToken("garbage-token", testSecret)
	if err == nil {
		t.Error("expected error for invalid token, got nil")
	}
}

func TestDecodeWrongSecret(t *testing.T) {
	token, err := CreateAccessToken("user1", 15, testSecret)
	if err != nil {
		t.Fatalf("create token: %v", err)
	}

	_, err = DecodeToken(token, "wrong-secret")
	if err == nil {
		t.Error("expected error for wrong secret, got nil")
	}
}

func TestVerifyResetToken(t *testing.T) {
	email := "test@example.com"
	token, err := CreateResetToken(email, 60, testSecret)
	if err != nil {
		t.Fatalf("create reset token: %v", err)
	}

	got, err := VerifyResetToken(token, testSecret)
	if err != nil {
		t.Fatalf("verify reset token: %v", err)
	}
	if got != email {
		t.Errorf("email = %q, want %q", got, email)
	}
}

func TestVerifyResetTokenRejectsAccessToken(t *testing.T) {
	token, err := CreateAccessToken("user1", 15, testSecret)
	if err != nil {
		t.Fatalf("create token: %v", err)
	}

	_, err = VerifyResetToken(token, testSecret)
	if err == nil {
		t.Error("expected error when using access token as reset token")
	}
}

func TestVerifyEmailVerificationToken(t *testing.T) {
	email := "test@example.com"
	token, err := CreateEmailVerificationToken(email, 24, testSecret)
	if err != nil {
		t.Fatalf("create token: %v", err)
	}

	got, err := VerifyEmailVerificationToken(token, testSecret)
	if err != nil {
		t.Fatalf("verify email token: %v", err)
	}
	if got != email {
		t.Errorf("email = %q, want %q", got, email)
	}
}
