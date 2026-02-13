import io
import pytest

from app.seed.round_templates import seed_round_templates


async def register_and_get_token(client, email="profile@test.com", username="profileuser"):
    reg = await client.post("/api/v1/auth/register", json={
        "email": email, "username": username, "password": "pass1234",
    })
    return reg.json()["access_token"]


@pytest.mark.asyncio
async def test_update_profile_bio(client):
    token = await register_and_get_token(client)
    headers = {"Authorization": f"Bearer {token}"}

    resp = await client.patch("/api/v1/users/me", json={"bio": "I shoot recurve"}, headers=headers)
    assert resp.status_code == 200
    assert resp.json()["bio"] == "I shoot recurve"

    resp = await client.get("/api/v1/users/me", headers=headers)
    assert resp.json()["bio"] == "I shoot recurve"


@pytest.mark.asyncio
async def test_upload_avatar_file(client):
    token = await register_and_get_token(client, "avatar@test.com", "avataruser")
    headers = {"Authorization": f"Bearer {token}"}

    # 1x1 red PNG
    png_bytes = (
        b"\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR\x00\x00\x00\x01"
        b"\x00\x00\x00\x01\x08\x02\x00\x00\x00\x90wS\xde\x00"
        b"\x00\x00\x0cIDATx\x9cc\xf8\x0f\x00\x00\x01\x01\x00"
        b"\x05\x18\xd8N\x00\x00\x00\x00IEND\xaeB`\x82"
    )
    resp = await client.post(
        "/api/v1/users/me/avatar",
        files={"file": ("avatar.png", io.BytesIO(png_bytes), "image/png")},
        headers=headers,
    )
    assert resp.status_code == 200
    assert resp.json()["avatar"].startswith("data:image/png;base64,")


@pytest.mark.asyncio
async def test_upload_avatar_invalid_type(client):
    token = await register_and_get_token(client, "badtype@test.com", "badtypeuser")
    headers = {"Authorization": f"Bearer {token}"}

    resp = await client.post(
        "/api/v1/users/me/avatar",
        files={"file": ("doc.pdf", io.BytesIO(b"%PDF-1.4"), "application/pdf")},
        headers=headers,
    )
    assert resp.status_code == 400


@pytest.mark.asyncio
async def test_delete_avatar(client):
    token = await register_and_get_token(client, "delavatar@test.com", "delavataruser")
    headers = {"Authorization": f"Bearer {token}"}

    # Upload first
    png_bytes = (
        b"\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR\x00\x00\x00\x01"
        b"\x00\x00\x00\x01\x08\x02\x00\x00\x00\x90wS\xde\x00"
        b"\x00\x00\x0cIDATx\x9cc\xf8\x0f\x00\x00\x01\x01\x00"
        b"\x05\x18\xd8N\x00\x00\x00\x00IEND\xaeB`\x82"
    )
    await client.post(
        "/api/v1/users/me/avatar",
        files={"file": ("avatar.png", io.BytesIO(png_bytes), "image/png")},
        headers=headers,
    )

    # Delete
    resp = await client.delete("/api/v1/users/me/avatar", headers=headers)
    assert resp.status_code == 200
    assert resp.json()["avatar"] is None


@pytest.mark.asyncio
async def test_bio_too_long(client):
    token = await register_and_get_token(client, "longbio@test.com", "longbiouser")
    headers = {"Authorization": f"Bearer {token}"}

    resp = await client.patch("/api/v1/users/me", json={"bio": "x" * 501}, headers=headers)
    assert resp.status_code == 422


@pytest.mark.asyncio
async def test_get_my_clubs(client):
    token = await register_and_get_token(client, "myclubs@test.com", "myclubsuser")
    headers = {"Authorization": f"Bearer {token}"}

    # No clubs initially
    resp = await client.get("/api/v1/users/me/clubs", headers=headers)
    assert resp.status_code == 200
    assert resp.json() == []

    # Create a club, then check
    resp = await client.post("/api/v1/clubs", json={"name": "My Club"}, headers=headers)
    assert resp.status_code == 201

    resp = await client.get("/api/v1/users/me/clubs", headers=headers)
    assert resp.status_code == 200
    clubs = resp.json()
    assert len(clubs) == 1
    assert clubs[0]["club_name"] == "My Club"
    assert clubs[0]["role"] == "owner"


@pytest.mark.asyncio
async def test_public_profile(client, db_session):
    await seed_round_templates(db_session)
    token = await register_and_get_token(client, "pubprof@test.com", "pubprofuser")
    headers = {"Authorization": f"Bearer {token}"}

    # Ensure profile is not public
    await client.patch("/api/v1/users/me", json={"profile_public": False}, headers=headers)

    resp = await client.get("/api/v1/users/pubprofuser")
    assert resp.status_code == 404

    # Make profile public
    resp = await client.patch("/api/v1/users/me", json={"profile_public": True}, headers=headers)
    assert resp.status_code == 200

    # Now accessible
    resp = await client.get("/api/v1/users/pubprofuser")
    assert resp.status_code == 200
    data = resp.json()
    assert data["username"] == "pubprofuser"
    assert data["total_sessions"] == 0
    assert data["recent_sessions"] == []


@pytest.mark.asyncio
async def test_public_profile_with_sessions(client, db_session):
    await seed_round_templates(db_session)
    token = await register_and_get_token(client, "pubsess@test.com", "pubsessuser")
    headers = {"Authorization": f"Bearer {token}"}

    # Make public
    await client.patch("/api/v1/users/me", json={"profile_public": True}, headers=headers)

    # Create and complete a session
    rounds = (await client.get("/api/v1/rounds")).json()
    vegas = next(r for r in rounds if r["name"] == "Vegas 300")
    stage_id = vegas["stages"][0]["id"]

    resp = await client.post("/api/v1/sessions", json={"template_id": vegas["id"]}, headers=headers)
    session_id = resp.json()["id"]

    await client.post(f"/api/v1/sessions/{session_id}/ends", json={
        "stage_id": stage_id,
        "arrows": [{"score_value": "X"}, {"score_value": "10"}, {"score_value": "9"}],
    }, headers=headers)
    await client.post(f"/api/v1/sessions/{session_id}/complete", headers=headers)

    resp = await client.get("/api/v1/users/pubsessuser")
    assert resp.status_code == 200
    data = resp.json()
    assert data["total_sessions"] == 1
    assert data["completed_sessions"] == 1
    assert data["total_arrows"] == 3
    assert data["personal_best_score"] == 29
    assert len(data["recent_sessions"]) == 1


@pytest.mark.asyncio
async def test_public_profile_nonexistent(client):
    resp = await client.get("/api/v1/users/nobody_here")
    assert resp.status_code == 404
