from app.config import settings


def send_verification_email(email: str, token: str) -> None:
    verify_link = f"{settings.FRONTEND_URL}/verify-email?token={token}"

    if not settings.SENDGRID_API_KEY:
        print(f"[DEV] Email verification link for {email}: {verify_link}")
        return

    from sendgrid import SendGridAPIClient
    from sendgrid.helpers.mail import Mail

    html_content = f"""
    <div style="font-family: sans-serif; max-width: 480px; margin: 0 auto;">
      <h2 style="color: #059669;">QuiverScore</h2>
      <p>Welcome to QuiverScore! Please verify your email address by clicking the button below:</p>
      <p><a href="{verify_link}" style="display: inline-block; background: #059669; color: white; padding: 12px 24px; border-radius: 6px; text-decoration: none;">Verify Email</a></p>
      <p style="color: #6b7280; font-size: 14px;">This link expires in {settings.EMAIL_VERIFICATION_TOKEN_EXPIRE_HOURS} hours. If you didn't create this account, you can safely ignore this email.</p>
    </div>
    """

    message = Mail(
        from_email=settings.SENDGRID_FROM_EMAIL,
        to_emails=email,
        subject="Verify your QuiverScore email",
        html_content=html_content,
    )
    sg = SendGridAPIClient(settings.SENDGRID_API_KEY)
    sg.send(message)


def send_password_reset_email(email: str, reset_token: str) -> None:
    reset_link = f"{settings.FRONTEND_URL}/reset-password?token={reset_token}"

    if not settings.SENDGRID_API_KEY:
        print(f"[DEV] Password reset link for {email}: {reset_link}")
        return

    from sendgrid import SendGridAPIClient
    from sendgrid.helpers.mail import Mail

    html_content = f"""
    <div style="font-family: sans-serif; max-width: 480px; margin: 0 auto;">
      <h2 style="color: #059669;">QuiverScore</h2>
      <p>You requested a password reset. Click the link below to set a new password:</p>
      <p><a href="{reset_link}" style="display: inline-block; background: #059669; color: white; padding: 12px 24px; border-radius: 6px; text-decoration: none;">Reset Password</a></p>
      <p style="color: #6b7280; font-size: 14px;">This link expires in {settings.PASSWORD_RESET_TOKEN_EXPIRE_MINUTES} minutes. If you didn't request this, you can safely ignore this email.</p>
    </div>
    """

    message = Mail(
        from_email=settings.SENDGRID_FROM_EMAIL,
        to_emails=email,
        subject="Reset your QuiverScore password",
        html_content=html_content,
    )
    sg = SendGridAPIClient(settings.SENDGRID_API_KEY)
    sg.send(message)
