"""
Contract tests for Notification endpoints.

Covers:
  - List notifications (empty, with items)
  - Unread count
  - Mark single notification as read
  - Mark all notifications as read
  - Unauthenticated access
"""

import os

import httpx
import pytest


API_BASE_URL = os.environ.get("API_BASE_URL", "")
if not API_BASE_URL:
    pytest.skip("API_BASE_URL not set — skipping contract tests", allow_module_level=True)


# ── Helpers ──────────────────────────────────────────────────────────────


def _submit_end(client, session_id, stage_id, arrows, headers):
    """Submit an end of arrows."""
    return client.post(f"/api/v1/sessions/{session_id}/ends", json={
        "stage_id": stage_id,
        "arrows": [{"score_value": v} for v in arrows],
    }, headers=headers)


def _create_and_complete_session(client, headers, create_round):
    """Create a session, submit one end, and complete it. Returns session data."""
    rnd = create_round(headers=headers)
    resp = client.post("/api/v1/sessions", json={"template_id": rnd["id"]}, headers=headers)
    assert resp.status_code == 201
    session = resp.json()
    stage_id = session["template"]["stages"][0]["id"]
    _submit_end(client, session["id"], stage_id, ["X", "10", "9"], headers)
    resp = client.post(f"/api/v1/sessions/{session['id']}/complete", headers=headers)
    assert resp.status_code == 200
    return resp.json()


# ══════════════════════════════════════════════════════════════════════════
# LIST NOTIFICATIONS
# ══════════════════════════════════════════════════════════════════════════


def test_list_notifications_empty(client, register_user):
    """GET /api/v1/notifications returns empty list for new user."""
    user = register_user()
    resp = client.get("/api/v1/notifications", headers=user["headers"])
    assert resp.status_code == 200
    assert resp.json() == []


def test_list_notifications_unauthenticated(client):
    """GET /api/v1/notifications without auth returns 401."""
    resp = client.get("/api/v1/notifications")
    assert resp.status_code == 401


def test_list_notifications_after_pr(client, register_user, create_round):
    """Completing a session that sets a personal record creates a notification."""
    user = register_user()
    _create_and_complete_session(client, user["headers"], create_round)

    resp = client.get("/api/v1/notifications", headers=user["headers"])
    assert resp.status_code == 200
    data = resp.json()
    assert len(data) >= 1
    notif = data[0]
    assert notif["type"] == "personal_record"
    assert "id" in notif
    assert "title" in notif
    assert "message" in notif
    assert notif["read"] is False
    assert "created_at" in notif


# ══════════════════════════════════════════════════════════════════════════
# UNREAD COUNT
# ══════════════════════════════════════════════════════════════════════════


def test_unread_count_zero(client, register_user):
    """GET /api/v1/notifications/unread-count returns 0 for new user."""
    user = register_user()
    resp = client.get("/api/v1/notifications/unread-count", headers=user["headers"])
    assert resp.status_code == 200
    assert resp.json()["count"] == 0


def test_unread_count_after_pr(client, register_user, create_round):
    """Unread count increments after personal record notification."""
    user = register_user()
    _create_and_complete_session(client, user["headers"], create_round)

    resp = client.get("/api/v1/notifications/unread-count", headers=user["headers"])
    assert resp.status_code == 200
    assert resp.json()["count"] >= 1


def test_unread_count_unauthenticated(client):
    """GET /api/v1/notifications/unread-count without auth returns 401."""
    resp = client.get("/api/v1/notifications/unread-count")
    assert resp.status_code == 401


# ══════════════════════════════════════════════════════════════════════════
# MARK READ
# ══════════════════════════════════════════════════════════════════════════


def test_mark_notification_read(client, register_user, create_round):
    """PATCH /api/v1/notifications/{id}/read marks a notification as read."""
    user = register_user()
    _create_and_complete_session(client, user["headers"], create_round)

    # Get the notification
    resp = client.get("/api/v1/notifications", headers=user["headers"])
    notif_id = resp.json()[0]["id"]

    resp = client.patch(f"/api/v1/notifications/{notif_id}/read", headers=user["headers"])
    assert resp.status_code == 200
    assert resp.json()["read"] is True
    assert resp.json()["id"] == notif_id


def test_mark_notification_read_not_found(client, register_user):
    """PATCH read on nonexistent notification returns 404."""
    user = register_user()
    resp = client.patch(
        "/api/v1/notifications/00000000-0000-0000-0000-000000000000/read",
        headers=user["headers"],
    )
    assert resp.status_code == 404


def test_mark_notification_read_unauthenticated(client):
    """PATCH read without auth returns 401."""
    resp = client.patch("/api/v1/notifications/00000000-0000-0000-0000-000000000000/read")
    assert resp.status_code == 401


# ══════════════════════════════════════════════════════════════════════════
# MARK ALL READ
# ══════════════════════════════════════════════════════════════════════════


def test_mark_all_read(client, register_user, create_round):
    """POST /api/v1/notifications/read-all marks all as read."""
    user = register_user()
    _create_and_complete_session(client, user["headers"], create_round)

    resp = client.post("/api/v1/notifications/read-all", headers=user["headers"])
    assert resp.status_code == 200
    assert "message" in resp.json()

    # Verify count is now 0
    resp = client.get("/api/v1/notifications/unread-count", headers=user["headers"])
    assert resp.json()["count"] == 0


def test_mark_all_read_unauthenticated(client):
    """POST read-all without auth returns 401."""
    resp = client.post("/api/v1/notifications/read-all")
    assert resp.status_code == 401
