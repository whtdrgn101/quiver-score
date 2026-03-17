package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenType identifies the purpose of a JWT.
type TokenType string

const (
	TokenTypeAccess            TokenType = "access"
	TokenTypeRefresh           TokenType = "refresh"
	TokenTypePasswordReset     TokenType = "password_reset"
	TokenTypeEmailVerification TokenType = "email_verification"
)

// Claims extends standard JWT claims with a type field.
type Claims struct {
	jwt.RegisteredClaims
	Type string `json:"type"`
}

// CreateToken creates a signed JWT with the given subject, type, and expiry.
func CreateToken(subject string, tokenType TokenType, expiry time.Duration, secretKey string) (string, error) {
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   subject,
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiry)),
		},
		Type: string(tokenType),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

// CreateAccessToken creates a short-lived access token for the given user ID.
func CreateAccessToken(userID string, expireMinutes int, secretKey string) (string, error) {
	return CreateToken(userID, TokenTypeAccess, time.Duration(expireMinutes)*time.Minute, secretKey)
}

// CreateRefreshToken creates a long-lived refresh token for the given user ID.
func CreateRefreshToken(userID string, expireDays int, secretKey string) (string, error) {
	return CreateToken(userID, TokenTypeRefresh, time.Duration(expireDays)*24*time.Hour, secretKey)
}

// CreateResetToken creates a password reset token for the given email.
func CreateResetToken(email string, expireMinutes int, secretKey string) (string, error) {
	return CreateToken(email, TokenTypePasswordReset, time.Duration(expireMinutes)*time.Minute, secretKey)
}

// CreateEmailVerificationToken creates an email verification token.
func CreateEmailVerificationToken(email string, expireHours int, secretKey string) (string, error) {
	return CreateToken(email, TokenTypeEmailVerification, time.Duration(expireHours)*time.Hour, secretKey)
}

// DecodeToken parses and validates a JWT, returning the claims or an error.
func DecodeToken(tokenString string, secretKey string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(secretKey), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}

	return claims, nil
}

// VerifyResetToken validates a password reset token and returns the email.
func VerifyResetToken(tokenString string, secretKey string) (string, error) {
	claims, err := DecodeToken(tokenString, secretKey)
	if err != nil {
		return "", err
	}
	if claims.Type != string(TokenTypePasswordReset) {
		return "", jwt.ErrSignatureInvalid
	}
	return claims.Subject, nil
}

// VerifyEmailVerificationToken validates an email verification token and returns the email.
func VerifyEmailVerificationToken(tokenString string, secretKey string) (string, error) {
	claims, err := DecodeToken(tokenString, secretKey)
	if err != nil {
		return "", err
	}
	if claims.Type != string(TokenTypeEmailVerification) {
		return "", jwt.ErrSignatureInvalid
	}
	return claims.Subject, nil
}
