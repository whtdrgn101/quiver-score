package auth

import (
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

func TestTokenFormatMatchesPython(t *testing.T) {
	// Verify that Go-generated tokens use the same claims structure as Python:
	// {sub: "<user_id>", type: "access"|"refresh"|..., exp: <timestamp>}
	// This ensures cross-compatibility since both use HS256 + same secret.

	userID := "550e8400-e29b-41d4-a716-446655440000"
	token, err := CreateAccessToken(userID, 15, testSecret)
	if err != nil {
		t.Fatalf("create token: %v", err)
	}

	// Parse without validation to inspect raw claims (like Python's jose.jwt.decode)
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	parsed, _, err := parser.ParseUnverified(token, jwt.MapClaims{})
	if err != nil {
		t.Fatalf("parse unverified: %v", err)
	}

	claims := parsed.Claims.(jwt.MapClaims)

	// Python tokens have exactly these fields: sub, type, exp
	if claims["sub"] != userID {
		t.Errorf("sub = %v, want %v", claims["sub"], userID)
	}
	if claims["type"] != "access" {
		t.Errorf("type = %v, want access", claims["type"])
	}
	if _, ok := claims["exp"]; !ok {
		t.Error("missing exp claim")
	}

	// Verify signing method matches Python (HS256)
	if parsed.Method.Alg() != "HS256" {
		t.Errorf("algorithm = %v, want HS256", parsed.Method.Alg())
	}
}
