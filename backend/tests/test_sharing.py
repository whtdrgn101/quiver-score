import uuid

import pytest

from app.seed.round_templates import seed_round_templates


async def _setup_completed_session(client, db_session):
    """Register a user, seed templates, create and complete a session. Returns (headers, session_id)."""
    await seed_round_templates(db_session)
    reg = await client.post("/api/v1/auth/register", json={
        "email": "sharer@test.com", "username": "sharer", "password": "pass1234",
    })
    token = reg.json()["access_token"]
    headers = {"Authorization": f"Bearer {token}"}

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

    return headers, session_id


@pytest.mark.asyncio
async def test_create_share_link(client, db_session):
    headers, session_id = await _setup_completed_session(client, db_session)

    resp = await client.post(f"/api/v1/share/sessions/{session_id}", headers=headers)
    assert resp.status_code == 200
    data = resp.json()
    assert "share_token" in data
    assert "url" in data
    assert data["share_token"] in data["url"]


@pytest.mark.asyncio
async def test_create_share_link_idempotent(client, db_session):
    headers, session_id = await _setup_completed_session(client, db_session)

    resp1 = await client.post(f"/api/v1/share/sessions/{session_id}", headers=headers)
    resp2 = await client.post(f"/api/v1/share/sessions/{session_id}", headers=headers)
    assert resp1.json()["share_token"] == resp2.json()["share_token"]


@pytest.mark.asyncio
async def test_get_shared_session(client, db_session):
    headers, session_id = await _setup_completed_session(client, db_session)

    share = await client.post(f"/api/v1/share/sessions/{session_id}", headers=headers)
    token = share.json()["share_token"]

    # Public endpoint — no auth needed
    resp = await client.get(f"/api/v1/share/s/{token}")
    assert resp.status_code == 200
    data = resp.json()
    assert data["archer_name"] == "sharer"
    assert data["total_score"] == 29
    assert data["total_x_count"] == 1
    assert len(data["ends"]) == 1


@pytest.mark.asyncio
async def test_get_shared_session_invalid_token(client, db_session):
    resp = await client.get("/api/v1/share/s/nonexistent_token")
    assert resp.status_code == 404


@pytest.mark.asyncio
async def test_revoke_share_link(client, db_session):
    headers, session_id = await _setup_completed_session(client, db_session)

    share = await client.post(f"/api/v1/share/sessions/{session_id}", headers=headers)
    token = share.json()["share_token"]

    # Revoke
    resp = await client.delete(f"/api/v1/share/sessions/{session_id}", headers=headers)
    assert resp.status_code == 200

    # Token no longer works
    resp = await client.get(f"/api/v1/share/s/{token}")
    assert resp.status_code == 404


@pytest.mark.asyncio
async def test_share_not_found(client, db_session):
    reg = await client.post("/api/v1/auth/register", json={
        "email": "nosession@test.com", "username": "nosession", "password": "pass1234",
    })
    token = reg.json()["access_token"]
    headers = {"Authorization": f"Bearer {token}"}

    fake_id = str(uuid.uuid4())
    resp = await client.post(f"/api/v1/share/sessions/{fake_id}", headers=headers)
    assert resp.status_code == 404
