"""
Contract tests for /api/v1/auth endpoints.

These validate status codes and response shapes against a live server.
They do NOT reach into the database or import app internals — that's
what makes them portable across Python and Go implementations.
"""


# ── Registration ────────────────────────────────────────────────────────


def test_register_success(client, unique):
    resp = client.post("/api/v1/auth/register", json={
        "email": f"{unique('reg')}@test.com",
        "username": unique("reg"),
        "password": "securepass123",
    })
    assert resp.status_code == 201
    data = resp.json()
    assert "access_token" in data
    assert "refresh_token" in data

    # Cleanup
    client.post("/api/v1/auth/delete-account", json={
        "confirmation": "Yes, delete ALL of my data",
    }, headers={"Authorization": f"Bearer {data['access_token']}"})


def test_register_duplicate_email(client, register_user):
    user = register_user()
    resp = client.post("/api/v1/auth/register", json={
        "email": user["email"],
        "username": "other_unique_name",
        "password": "securepass123",
    })
    assert resp.status_code == 409


def test_register_duplicate_username(client, register_user):
    user = register_user()
    resp = client.post("/api/v1/auth/register", json={
        "email": "other_unique@test.com",
        "username": user["username"],
        "password": "securepass123",
    })
    assert resp.status_code == 409


def test_register_short_password(client, unique):
    resp = client.post("/api/v1/auth/register", json={
        "email": f"{unique('short')}@test.com",
        "username": unique("short"),
        "password": "short",
    })
    assert resp.status_code == 422


def test_register_short_username(client, unique):
    resp = client.post("/api/v1/auth/register", json={
        "email": f"{unique('su')}@test.com",
        "username": "ab",
        "password": "securepass123",
    })
    assert resp.status_code == 422


def test_register_missing_email(client, unique):
    resp = client.post("/api/v1/auth/register", json={
        "username": unique("noemail"),
        "password": "securepass123",
    })
    assert resp.status_code == 422


# ── Login ───────────────────────────────────────────────────────────────


def test_login_success(client, register_user):
    user = register_user()
    resp = client.post("/api/v1/auth/login", json={
        "username": user["username"],
        "password": user["password"],
    })
    assert resp.status_code == 200
    data = resp.json()
    assert "access_token" in data
    assert "refresh_token" in data


def test_login_wrong_password(client, register_user):
    user = register_user()
    resp = client.post("/api/v1/auth/login", json={
        "username": user["username"],
        "password": "wrongpassword",
    })
    assert resp.status_code == 401


def test_login_nonexistent_user(client):
    resp = client.post("/api/v1/auth/login", json={
        "username": "ghost_user_999",
        "password": "pass1234",
    })
    assert resp.status_code == 401


# ── Token Refresh ───────────────────────────────────────────────────────


def test_refresh_success(client, register_user):
    user = register_user()
    resp = client.post("/api/v1/auth/refresh", json={
        "refresh_token": user["refresh_token"],
    })
    assert resp.status_code == 200
    data = resp.json()
    assert "access_token" in data
    assert "refresh_token" in data


def test_refresh_invalid_token(client):
    resp = client.post("/api/v1/auth/refresh", json={
        "refresh_token": "garbage-token-value",
    })
    assert resp.status_code == 401


def test_refresh_access_token_rejected(client, register_user):
    """Using an access token as a refresh token should fail."""
    user = register_user()
    resp = client.post("/api/v1/auth/refresh", json={
        "refresh_token": user["access_token"],
    })
    assert resp.status_code == 401


# ── Get Current User ────────────────────────────────────────────────────


def test_get_me(client, register_user):
    user = register_user()
    resp = client.get("/api/v1/users/me", headers=user["headers"])
    assert resp.status_code == 200
    data = resp.json()
    assert data["username"] == user["username"]
    assert data["email"] == user["email"]


def test_get_me_unauthenticated(client):
    resp = client.get("/api/v1/users/me")
    assert resp.status_code == 401


# ── Change Password ─────────────────────────────────────────────────────


def test_change_password_success(client, register_user):
    user = register_user()
    resp = client.post("/api/v1/auth/change-password", json={
        "current_password": user["password"],
        "new_password": "newpass456",
    }, headers=user["headers"])
    assert resp.status_code == 200

    # Old password no longer works
    resp = client.post("/api/v1/auth/login", json={
        "username": user["username"],
        "password": user["password"],
    })
    assert resp.status_code == 401

    # New password works
    resp = client.post("/api/v1/auth/login", json={
        "username": user["username"],
        "password": "newpass456",
    })
    assert resp.status_code == 200


def test_change_password_wrong_current(client, register_user):
    user = register_user()
    resp = client.post("/api/v1/auth/change-password", json={
        "current_password": "wrongwrong",
        "new_password": "newpass456",
    }, headers=user["headers"])
    assert resp.status_code == 401


# ── Forgot / Reset Password ────────────────────────────────────────────


def test_forgot_password_registered(client, register_user):
    user = register_user()
    resp = client.post("/api/v1/auth/forgot-password", json={
        "email": user["email"],
    })
    assert resp.status_code == 200
    # Should always return a generic success message
    assert "detail" in resp.json()


def test_forgot_password_unregistered(client):
    resp = client.post("/api/v1/auth/forgot-password", json={
        "email": "nobody@doesnotexist.com",
    })
    assert resp.status_code == 200
    # Same generic message for security (no email enumeration)
    assert "detail" in resp.json()


def test_reset_password_invalid_token(client):
    resp = client.post("/api/v1/auth/reset-password", json={
        "token": "bad-token",
        "new_password": "newpass1234",
    })
    assert resp.status_code == 401


# ── Email Verification ──────────────────────────────────────────────────


def test_verify_email_invalid_token(client):
    resp = client.post("/api/v1/auth/verify-email", json={
        "token": "bad-token-value",
    })
    assert resp.status_code == 401


def test_resend_verification(client, register_user):
    """Newly registered user can request verification resend."""
    user = register_user()
    resp = client.post("/api/v1/auth/resend-verification", headers=user["headers"])
    assert resp.status_code == 200
    assert "detail" in resp.json()


# ── Delete Account ──────────────────────────────────────────────────────


def test_delete_account_wrong_confirmation(client, register_user):
    user = register_user()
    resp = client.post("/api/v1/auth/delete-account", json={
        "confirmation": "yes delete my data",
    }, headers=user["headers"])
    assert resp.status_code == 401


def test_delete_account_unauthenticated(client):
    resp = client.post("/api/v1/auth/delete-account", json={
        "confirmation": "Yes, delete ALL of my data",
    })
    assert resp.status_code == 401


def test_delete_account_success(client, register_user):
    user = register_user()

    resp = client.post("/api/v1/auth/delete-account", json={
        "confirmation": "Yes, delete ALL of my data",
    }, headers=user["headers"])
    assert resp.status_code == 200

    # Login should fail after deletion
    resp = client.post("/api/v1/auth/login", json={
        "username": user["username"],
        "password": user["password"],
    })
    assert resp.status_code == 401
