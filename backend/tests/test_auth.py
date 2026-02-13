import uuid
from datetime import timedelta

import pytest

from app.core.security import create_token, create_access_token, create_email_verification_token, create_reset_token


@pytest.mark.asyncio
async def test_register(client):
    resp = await client.post("/api/v1/auth/register", json={
        "email": "archer@test.com",
        "username": "archer1",
        "password": "securepass123",
    })
    assert resp.status_code == 201
    data = resp.json()
    assert "access_token" in data
    assert "refresh_token" in data


@pytest.mark.asyncio
async def test_register_duplicate(client):
    await client.post("/api/v1/auth/register", json={
        "email": "dup@test.com", "username": "dupuser", "password": "pass1234",
    })
    resp = await client.post("/api/v1/auth/register", json={
        "email": "dup@test.com", "username": "dupuser2", "password": "pass1234",
    })
    assert resp.status_code == 409


@pytest.mark.asyncio
async def test_login(client):
    await client.post("/api/v1/auth/register", json={
        "email": "login@test.com", "username": "loginuser", "password": "pass1234",
    })
    resp = await client.post("/api/v1/auth/login", json={
        "username": "loginuser", "password": "pass1234",
    })
    assert resp.status_code == 200
    assert "access_token" in resp.json()


@pytest.mark.asyncio
async def test_login_wrong_password(client):
    await client.post("/api/v1/auth/register", json={
        "email": "wrong@test.com", "username": "wronguser", "password": "pass1234",
    })
    resp = await client.post("/api/v1/auth/login", json={
        "username": "wronguser", "password": "wrongpass",
    })
    assert resp.status_code == 401


@pytest.mark.asyncio
async def test_refresh_token(client):
    reg = await client.post("/api/v1/auth/register", json={
        "email": "refresh@test.com", "username": "refreshuser", "password": "pass1234",
    })
    refresh_token = reg.json()["refresh_token"]
    resp = await client.post("/api/v1/auth/refresh", json={
        "refresh_token": refresh_token,
    })
    assert resp.status_code == 200
    assert "access_token" in resp.json()


@pytest.mark.asyncio
async def test_get_me(client):
    reg = await client.post("/api/v1/auth/register", json={
        "email": "me@test.com", "username": "meuser", "password": "pass1234",
    })
    token = reg.json()["access_token"]
    resp = await client.get("/api/v1/users/me", headers={"Authorization": f"Bearer {token}"})
    assert resp.status_code == 200
    assert resp.json()["username"] == "meuser"


@pytest.mark.asyncio
async def test_change_password(client):
    reg = await client.post("/api/v1/auth/register", json={
        "email": "changepw@test.com", "username": "changepwuser", "password": "oldpass123",
    })
    token = reg.json()["access_token"]
    headers = {"Authorization": f"Bearer {token}"}

    resp = await client.post("/api/v1/auth/change-password", json={
        "current_password": "oldpass123",
        "new_password": "newpass456",
    }, headers=headers)
    assert resp.status_code == 200

    # Old password no longer works
    resp = await client.post("/api/v1/auth/login", json={
        "username": "changepwuser", "password": "oldpass123",
    })
    assert resp.status_code == 401

    # New password works
    resp = await client.post("/api/v1/auth/login", json={
        "username": "changepwuser", "password": "newpass456",
    })
    assert resp.status_code == 200


@pytest.mark.asyncio
async def test_change_password_wrong_current(client):
    reg = await client.post("/api/v1/auth/register", json={
        "email": "wrongcur@test.com", "username": "wrongcuruser", "password": "mypass123",
    })
    token = reg.json()["access_token"]
    headers = {"Authorization": f"Bearer {token}"}

    resp = await client.post("/api/v1/auth/change-password", json={
        "current_password": "wrongwrong",
        "new_password": "newpass456",
    }, headers=headers)
    assert resp.status_code == 401


@pytest.mark.asyncio
async def test_register_short_password(client):
    resp = await client.post("/api/v1/auth/register", json={
        "email": "short@test.com", "username": "shortuser", "password": "short",
    })
    assert resp.status_code == 422


@pytest.mark.asyncio
async def test_refresh_invalid_token(client):
    resp = await client.post("/api/v1/auth/refresh", json={
        "refresh_token": "garbage-token-value",
    })
    assert resp.status_code == 401


@pytest.mark.asyncio
async def test_refresh_access_token_as_refresh(client):
    reg = await client.post("/api/v1/auth/register", json={
        "email": "refacc@test.com", "username": "refaccuser", "password": "pass1234",
    })
    access_token = reg.json()["access_token"]
    resp = await client.post("/api/v1/auth/refresh", json={
        "refresh_token": access_token,  # wrong token type
    })
    assert resp.status_code == 401


@pytest.mark.asyncio
async def test_refresh_missing_sub(client):
    # JWT with type=refresh but no sub claim
    token = create_token({"type": "refresh"}, timedelta(days=1))
    resp = await client.post("/api/v1/auth/refresh", json={
        "refresh_token": token,
    })
    assert resp.status_code == 401


@pytest.mark.asyncio
async def test_refresh_deleted_user(client):
    # Create a refresh token for a nonexistent user UUID
    fake_user_id = str(uuid.uuid4())
    token = create_token({"sub": fake_user_id, "type": "refresh"}, timedelta(days=1))
    resp = await client.post("/api/v1/auth/refresh", json={
        "refresh_token": token,
    })
    assert resp.status_code == 401


@pytest.mark.asyncio
async def test_verify_email_invalid_token(client):
    resp = await client.post("/api/v1/auth/verify-email", json={
        "token": "bad-token-value",
    })
    assert resp.status_code == 401


@pytest.mark.asyncio
async def test_verify_email_success(client, db_session):
    # Register to get a verification token stored on user
    reg = await client.post("/api/v1/auth/register", json={
        "email": "verify@test.com", "username": "verifyuser", "password": "pass1234",
    })
    token = reg.json()["access_token"]

    # Read the user's email_verification_token from the DB
    from sqlalchemy import select
    from app.models.user import User
    result = await db_session.execute(select(User).where(User.username == "verifyuser"))
    user = result.scalar_one()
    verification_token = user.email_verification_token
    assert verification_token is not None

    # Verify
    resp = await client.post("/api/v1/auth/verify-email", json={
        "token": verification_token,
    })
    assert resp.status_code == 200
    assert "verified" in resp.json()["detail"].lower()

    # Check user is now verified
    headers = {"Authorization": f"Bearer {token}"}
    resp = await client.get("/api/v1/users/me", headers=headers)
    assert resp.json()["email_verified"] is True


@pytest.mark.asyncio
async def test_resend_verification_already_verified(client, db_session):
    reg = await client.post("/api/v1/auth/register", json={
        "email": "alreadyver@test.com", "username": "alreadyver", "password": "pass1234",
    })
    token = reg.json()["access_token"]
    headers = {"Authorization": f"Bearer {token}"}

    # Manually verify user in DB
    from sqlalchemy import select
    from app.models.user import User
    result = await db_session.execute(select(User).where(User.username == "alreadyver"))
    user = result.scalar_one()
    user.email_verified = True
    await db_session.commit()

    resp = await client.post("/api/v1/auth/resend-verification", headers=headers)
    assert resp.status_code == 200
    assert "already verified" in resp.json()["detail"].lower()


@pytest.mark.asyncio
async def test_resend_verification_unverified(client, db_session):
    reg = await client.post("/api/v1/auth/register", json={
        "email": "unver@test.com", "username": "unveruser", "password": "pass1234",
    })
    token = reg.json()["access_token"]
    headers = {"Authorization": f"Bearer {token}"}

    # Ensure user is unverified (SQLite server_default="false" may be truthy)
    from sqlalchemy import select
    from app.models.user import User
    result = await db_session.execute(select(User).where(User.username == "unveruser"))
    user = result.scalar_one()
    user.email_verified = False
    await db_session.commit()

    resp = await client.post("/api/v1/auth/resend-verification", headers=headers)
    assert resp.status_code == 200
    assert "sent" in resp.json()["detail"].lower()


@pytest.mark.asyncio
async def test_forgot_password_registered(client):
    await client.post("/api/v1/auth/register", json={
        "email": "forgot@test.com", "username": "forgotuser", "password": "pass1234",
    })
    resp = await client.post("/api/v1/auth/forgot-password", json={
        "email": "forgot@test.com",
    })
    assert resp.status_code == 200
    assert "reset link" in resp.json()["detail"].lower()


@pytest.mark.asyncio
async def test_forgot_password_unregistered(client):
    resp = await client.post("/api/v1/auth/forgot-password", json={
        "email": "nobody@test.com",
    })
    assert resp.status_code == 200
    # Same generic success message for security
    assert "reset link" in resp.json()["detail"].lower()


@pytest.mark.asyncio
async def test_reset_password_invalid_token(client):
    resp = await client.post("/api/v1/auth/reset-password", json={
        "token": "bad-token",
        "new_password": "newpass1234",
    })
    assert resp.status_code == 401


@pytest.mark.asyncio
async def test_reset_password_success(client):
    await client.post("/api/v1/auth/register", json={
        "email": "reset@test.com", "username": "resetuser", "password": "oldpass123",
    })

    # Generate a valid reset token
    reset_token = create_reset_token("reset@test.com")

    resp = await client.post("/api/v1/auth/reset-password", json={
        "token": reset_token,
        "new_password": "brandnew123",
    })
    assert resp.status_code == 200

    # Old password fails
    resp = await client.post("/api/v1/auth/login", json={
        "username": "resetuser", "password": "oldpass123",
    })
    assert resp.status_code == 401

    # New password works
    resp = await client.post("/api/v1/auth/login", json={
        "username": "resetuser", "password": "brandnew123",
    })
    assert resp.status_code == 200


@pytest.mark.asyncio
async def test_register_short_username(client):
    resp = await client.post("/api/v1/auth/register", json={
        "email": "shortname@test.com", "username": "ab", "password": "pass1234",
    })
    assert resp.status_code == 422


@pytest.mark.asyncio
async def test_register_missing_email(client):
    resp = await client.post("/api/v1/auth/register", json={
        "username": "noemailuser", "password": "pass1234",
    })
    assert resp.status_code == 422


@pytest.mark.asyncio
async def test_login_nonexistent_user(client):
    resp = await client.post("/api/v1/auth/login", json={
        "username": "ghost_user_999", "password": "pass1234",
    })
    assert resp.status_code == 401


@pytest.mark.asyncio
async def test_refresh_with_expired_token(client):
    # Create a token that expired 1 second ago
    token = create_token(
        {"sub": str(uuid.uuid4()), "type": "refresh"},
        timedelta(seconds=-1),
    )
    resp = await client.post("/api/v1/auth/refresh", json={
        "refresh_token": token,
    })
    assert resp.status_code == 401


# ── Delete Account ──────────────────────────────────────────────────────


async def _register_helper(client, email, username):
    resp = await client.post("/api/v1/auth/register", json={
        "email": email, "username": username, "password": "pass1234",
    })
    token = resp.json()["access_token"]
    return token, {"Authorization": f"Bearer {token}"}


@pytest.mark.asyncio
async def test_delete_account_wrong_confirmation(client):
    _, headers = await _register_helper(client, "del1@test.com", "delaccount1")
    resp = await client.post("/api/v1/auth/delete-account", json={
        "confirmation": "yes delete my data",
    }, headers=headers)
    assert resp.status_code == 401


@pytest.mark.asyncio
async def test_delete_account_case_sensitive(client):
    _, headers = await _register_helper(client, "del2@test.com", "delaccount2")
    resp = await client.post("/api/v1/auth/delete-account", json={
        "confirmation": "yes, delete ALL of my data",  # lowercase y
    }, headers=headers)
    assert resp.status_code == 401


@pytest.mark.asyncio
async def test_delete_account_success(client, db_session):
    """Full account deletion removes all user data."""
    from app.seed.round_templates import seed_round_templates
    await seed_round_templates(db_session)
    await db_session.commit()

    _, headers = await _register_helper(client, "del3@test.com", "delaccount3")

    # Create some data: custom round, scoring session with ends
    resp = await client.post("/api/v1/rounds", json={
        "name": "My Round",
        "organization": "Custom",
        "description": "test",
        "stages": [{
            "name": "Stage 1", "distance": "10m", "num_ends": 2, "arrows_per_end": 3,
            "allowed_values": ["10", "9", "M"],
            "value_score_map": {"10": 10, "9": 9, "M": 0},
            "max_score_per_arrow": 10,
        }],
    }, headers=headers)
    custom_round_id = resp.json()["id"]
    stage_id = resp.json()["stages"][0]["id"]

    resp = await client.post("/api/v1/sessions", json={
        "template_id": custom_round_id,
    }, headers=headers)
    session_id = resp.json()["id"]

    await client.post(f"/api/v1/sessions/{session_id}/ends", json={
        "stage_id": stage_id,
        "arrows": [{"score_value": "10"}, {"score_value": "9"}, {"score_value": "M"}],
    }, headers=headers)

    # Delete account
    resp = await client.post("/api/v1/auth/delete-account", json={
        "confirmation": "Yes, delete ALL of my data",
    }, headers=headers)
    assert resp.status_code == 200

    # Verify user is gone — login should fail
    resp = await client.post("/api/v1/auth/login", json={
        "username": "delaccount3", "password": "pass1234",
    })
    assert resp.status_code == 401

    # Custom round template should be gone
    resp3 = await client.get(f"/api/v1/rounds/{custom_round_id}")
    assert resp3.status_code == 404


@pytest.mark.asyncio
async def test_delete_account_unauthenticated(client):
    resp = await client.post("/api/v1/auth/delete-account", json={
        "confirmation": "Yes, delete ALL of my data",
    })
    assert resp.status_code == 401
