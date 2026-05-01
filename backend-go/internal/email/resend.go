package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

type Sender struct {
	APIKey    string
	FromEmail string
}

func (s *Sender) SendVerificationEmail(toEmail, token, frontendURL string, expireHours int) error {
	verifyLink := fmt.Sprintf("%s/verify-email?token=%s", frontendURL, token)

	if s.APIKey == "" {
		slog.Info("Email verification link", "email", toEmail, "link", verifyLink)
		return nil
	}

	html := fmt.Sprintf(`
	<div style="font-family: sans-serif; max-width: 480px; margin: 0 auto;">
	  <h2 style="color: #059669;">QuiverScore</h2>
	  <p>Welcome to QuiverScore! Please verify your email address by clicking the button below:</p>
	  <p><a href="%s" style="display: inline-block; background: #059669; color: white; padding: 12px 24px; border-radius: 6px; text-decoration: none;">Verify Email</a></p>
	  <p style="color: #6b7280; font-size: 14px;">This link expires in %d hours. If you didn't create this account, you can safely ignore this email.</p>
	</div>`, verifyLink, expireHours)

	return s.send(toEmail, "Verify your QuiverScore email", html)
}

func (s *Sender) SendPasswordResetEmail(toEmail, token, frontendURL string, expireMinutes int) error {
	resetLink := fmt.Sprintf("%s/reset-password?token=%s", frontendURL, token)

	if s.APIKey == "" {
		slog.Info("Password reset link", "email", toEmail, "link", resetLink)
		return nil
	}

	html := fmt.Sprintf(`
	<div style="font-family: sans-serif; max-width: 480px; margin: 0 auto;">
	  <h2 style="color: #059669;">QuiverScore</h2>
	  <p>You requested a password reset. Click the link below to set a new password:</p>
	  <p><a href="%s" style="display: inline-block; background: #059669; color: white; padding: 12px 24px; border-radius: 6px; text-decoration: none;">Reset Password</a></p>
	  <p style="color: #6b7280; font-size: 14px;">This link expires in %d minutes. If you didn't request this, you can safely ignore this email.</p>
	</div>`, resetLink, expireMinutes)

	return s.send(toEmail, "Reset your QuiverScore password", html)
}

type resendPayload struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
}

func (s *Sender) send(toEmail, subject, html string) error {
	payload := resendPayload{
		From:    s.FromEmail,
		To:      []string{toEmail},
		Subject: subject,
		HTML:    html,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+s.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		slog.Error("Resend API error", "status", resp.StatusCode, "to", toEmail, "body", string(respBody))
		return fmt.Errorf("resend: status %d", resp.StatusCode)
	}

	slog.Info("email sent", "to", toEmail, "subject", subject)
	return nil
}
