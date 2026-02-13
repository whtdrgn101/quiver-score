import pytest

from app.seed.round_templates import seed_round_templates

_user_counter = 0


async def _register_and_get_token(client, email=None, username=None):
    global _user_counter
    _user_counter += 1
    email = email or f"notif{_user_counter}@test.com"
    username = username or f"notif{_user_counter}"
    resp = await client.post("/api/v1/auth/register", json={
        "email": email, "username": username, "password": "pass1234",
    })
    return resp.json()["access_token"]


@pytest.mark.asyncio
async def test_list_notifications_empty(client, db_session):
    """GET /notifications returns empty list initially."""
    token = await _register_and_get_token(client)
    headers = {"Authorization": f"Bearer {token}"}
    resp = await client.get("/api/v1/notifications", headers=headers)
    assert resp.status_code == 200
    assert resp.json() == []


@pytest.mark.asyncio
async def test_unread_count_zero(client, db_session):
    """GET /notifications/unread-count returns 0 initially."""
    token = await _register_and_get_token(client)
    headers = {"Authorization": f"Bearer {token}"}
    resp = await client.get("/api/v1/notifications/unread-count", headers=headers)
    assert resp.status_code == 200
    assert resp.json()["count"] == 0


@pytest.mark.asyncio
async def test_pr_triggers_notification(client, db_session):
    """Completing a session with a PR creates a notification."""
    token = await _register_and_get_token(client)
    headers = {"Authorization": f"Bearer {token}"}
    await seed_round_templates(db_session)
    await db_session.commit()

    rounds = await client.get("/api/v1/rounds", headers=headers)
    template = rounds.json()[0]
    stage = template["stages"][0]

    session_resp = await client.post("/api/v1/sessions", json={"template_id": template["id"]}, headers=headers)
    session_id = session_resp.json()["id"]

    arrows = [{"score_value": stage["allowed_values"][0]}] * stage["arrows_per_end"]
    await client.post(f"/api/v1/sessions/{session_id}/ends", json={
        "stage_id": stage["id"],
        "arrows": arrows,
    }, headers=headers)
    await client.post(f"/api/v1/sessions/{session_id}/complete", headers=headers)

    resp = await client.get("/api/v1/notifications", headers=headers)
    assert resp.status_code == 200
    notifications = resp.json()
    assert len(notifications) == 1
    assert notifications[0]["type"] == "personal_record"
    assert "Personal Record" in notifications[0]["title"]

    # Check unread count
    resp = await client.get("/api/v1/notifications/unread-count", headers=headers)
    assert resp.json()["count"] == 1


@pytest.mark.asyncio
async def test_mark_notification_read(client, db_session):
    """PATCH /notifications/{id}/read marks a notification as read."""
    token = await _register_and_get_token(client)
    headers = {"Authorization": f"Bearer {token}"}
    await seed_round_templates(db_session)
    await db_session.commit()

    rounds = await client.get("/api/v1/rounds", headers=headers)
    template = rounds.json()[0]
    stage = template["stages"][0]

    session_resp = await client.post("/api/v1/sessions", json={"template_id": template["id"]}, headers=headers)
    session_id = session_resp.json()["id"]

    arrows = [{"score_value": stage["allowed_values"][0]}] * stage["arrows_per_end"]
    await client.post(f"/api/v1/sessions/{session_id}/ends", json={
        "stage_id": stage["id"],
        "arrows": arrows,
    }, headers=headers)
    await client.post(f"/api/v1/sessions/{session_id}/complete", headers=headers)

    notifs = await client.get("/api/v1/notifications", headers=headers)
    notif_id = notifs.json()[0]["id"]

    resp = await client.patch(f"/api/v1/notifications/{notif_id}/read", headers=headers)
    assert resp.status_code == 200
    assert resp.json()["read"] is True

    resp = await client.get("/api/v1/notifications/unread-count", headers=headers)
    assert resp.json()["count"] == 0


@pytest.mark.asyncio
async def test_mark_all_read(client, db_session):
    """POST /notifications/read-all marks all notifications as read."""
    token = await _register_and_get_token(client)
    headers = {"Authorization": f"Bearer {token}"}
    await seed_round_templates(db_session)
    await db_session.commit()

    rounds = await client.get("/api/v1/rounds", headers=headers)
    template = rounds.json()[0]
    stage = template["stages"][0]

    # Create two sessions to generate two PR notifications (different score for 2nd)
    for _ in range(2):
        session_resp = await client.post("/api/v1/sessions", json={"template_id": template["id"]}, headers=headers)
        session_id = session_resp.json()["id"]
        arrows = [{"score_value": stage["allowed_values"][0]}] * stage["arrows_per_end"]
        await client.post(f"/api/v1/sessions/{session_id}/ends", json={
            "stage_id": stage["id"],
            "arrows": arrows,
        }, headers=headers)
        await client.post(f"/api/v1/sessions/{session_id}/complete", headers=headers)

    resp = await client.get("/api/v1/notifications/unread-count", headers=headers)
    assert resp.json()["count"] >= 1

    resp = await client.post("/api/v1/notifications/read-all", headers=headers)
    assert resp.status_code == 200

    resp = await client.get("/api/v1/notifications/unread-count", headers=headers)
    assert resp.json()["count"] == 0
