import pytest

from app.seed.round_templates import seed_round_templates

_user_counter = 0


async def _register_and_get_token(client, email=None, username=None):
    global _user_counter
    _user_counter += 1
    email = email or f"social{_user_counter}@test.com"
    username = username or f"social{_user_counter}"
    resp = await client.post("/api/v1/auth/register", json={
        "email": email, "username": username, "password": "pass1234",
    })
    token = resp.json()["access_token"]
    # Get user ID from profile
    me = await client.get("/api/v1/users/me", headers={"Authorization": f"Bearer {token}"})
    user_id = me.json()["id"]
    return token, user_id


@pytest.mark.asyncio
async def test_follow_user(client, db_session):
    """POST /social/follow/{id} creates a follow."""
    token1, _ = await _register_and_get_token(client)
    _, user2_id = await _register_and_get_token(client)

    resp = await client.post(f"/api/v1/social/follow/{user2_id}",
                             headers={"Authorization": f"Bearer {token1}"})
    assert resp.status_code == 201
    assert resp.json()["following_id"] == user2_id


@pytest.mark.asyncio
async def test_unfollow_user(client, db_session):
    """DELETE /social/follow/{id} removes a follow."""
    token1, _ = await _register_and_get_token(client)
    _, user2_id = await _register_and_get_token(client)

    await client.post(f"/api/v1/social/follow/{user2_id}",
                      headers={"Authorization": f"Bearer {token1}"})
    resp = await client.delete(f"/api/v1/social/follow/{user2_id}",
                               headers={"Authorization": f"Bearer {token1}"})
    assert resp.status_code == 204


@pytest.mark.asyncio
async def test_followers_list(client, db_session):
    """GET /social/followers lists my followers."""
    token1, user1_id = await _register_and_get_token(client)
    token2, _ = await _register_and_get_token(client)

    # User2 follows User1
    await client.post(f"/api/v1/social/follow/{user1_id}",
                      headers={"Authorization": f"Bearer {token2}"})

    resp = await client.get("/api/v1/social/followers",
                            headers={"Authorization": f"Bearer {token1}"})
    assert resp.status_code == 200
    assert len(resp.json()) == 1


@pytest.mark.asyncio
async def test_following_list(client, db_session):
    """GET /social/following lists who I follow."""
    token1, _ = await _register_and_get_token(client)
    _, user2_id = await _register_and_get_token(client)

    await client.post(f"/api/v1/social/follow/{user2_id}",
                      headers={"Authorization": f"Bearer {token1}"})

    resp = await client.get("/api/v1/social/following",
                            headers={"Authorization": f"Bearer {token1}"})
    assert resp.status_code == 200
    assert len(resp.json()) == 1


@pytest.mark.asyncio
async def test_feed_populates(client, db_session):
    """GET /social/feed shows items from followed users."""
    await seed_round_templates(db_session)
    token1, _ = await _register_and_get_token(client)
    token2, user2_id = await _register_and_get_token(client)

    # User1 follows User2
    await client.post(f"/api/v1/social/follow/{user2_id}",
                      headers={"Authorization": f"Bearer {token1}"})

    # User2 completes a session
    rounds_resp = await client.get("/api/v1/rounds", headers={"Authorization": f"Bearer {token2}"})
    template = next(r for r in rounds_resp.json() if r["name"] == "WA Indoor 18m (Recurve)")
    template_id = template["id"]
    stage_id = template["stages"][0]["id"]

    s_resp = await client.post("/api/v1/sessions", json={"template_id": template_id},
                               headers={"Authorization": f"Bearer {token2}"})
    session_id = s_resp.json()["id"]
    await client.post(f"/api/v1/sessions/{session_id}/ends", json={
        "stage_id": stage_id,
        "arrows": [{"score_value": "10"}, {"score_value": "9"}, {"score_value": "8"}],
    }, headers={"Authorization": f"Bearer {token2}"})
    await client.post(f"/api/v1/sessions/{session_id}/complete",
                      headers={"Authorization": f"Bearer {token2}"})

    # User1 checks feed
    resp = await client.get("/api/v1/social/feed",
                            headers={"Authorization": f"Bearer {token1}"})
    assert resp.status_code == 200
    assert len(resp.json()) == 1
    assert resp.json()[0]["type"] in ("session_completed", "personal_record")
    assert resp.json()[0]["data"]["total_score"] == 27
