"""
Contract tests for Social & Coaching endpoints.

Covers:
  - Follow/unfollow + list followers/following
  - Activity feed
  - Coach/athlete invites, respond, list
  - Coach viewing athlete sessions
  - Session annotations (create, list)
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


def _complete_session(client, session_id, headers):
    """Submit one end and complete the session."""
    # Get session to find stage_id
    resp = client.get(f"/api/v1/sessions/{session_id}", headers=headers)
    session = resp.json()
    stage_id = session["template"]["stages"][0]["id"]
    _submit_end(client, session_id, stage_id, ["X", "10", "9"], headers)
    resp = client.post(f"/api/v1/sessions/{session_id}/complete", headers=headers)
    assert resp.status_code == 200
    return resp.json()


# ══════════════════════════════════════════════════════════════════════════
# SOCIAL: Follow / Unfollow / Feed
# ══════════════════════════════════════════════════════════════════════════


def test_follow_user(client, register_user):
    """POST /api/v1/social/follow/{user_id} creates a follow relationship."""
    user_a = register_user()
    user_b = register_user()

    # Get user_b's ID
    resp = client.get("/api/v1/users/me", headers=user_b["headers"])
    user_b_id = resp.json()["id"]

    resp = client.post(f"/api/v1/social/follow/{user_b_id}", headers=user_a["headers"])
    assert resp.status_code == 201
    data = resp.json()
    assert data["following_id"] == user_b_id
    assert "id" in data
    assert "created_at" in data
    assert data["following_username"] == user_b["username"]


def test_follow_self_rejected(client, register_user):
    """POST /api/v1/social/follow/{self} returns 422."""
    user = register_user()
    resp = client.get("/api/v1/users/me", headers=user["headers"])
    user_id = resp.json()["id"]

    resp = client.post(f"/api/v1/social/follow/{user_id}", headers=user["headers"])
    assert resp.status_code == 422


def test_follow_duplicate_rejected(client, register_user):
    """Following same user twice returns 409."""
    user_a = register_user()
    user_b = register_user()

    resp = client.get("/api/v1/users/me", headers=user_b["headers"])
    user_b_id = resp.json()["id"]

    client.post(f"/api/v1/social/follow/{user_b_id}", headers=user_a["headers"])
    resp = client.post(f"/api/v1/social/follow/{user_b_id}", headers=user_a["headers"])
    assert resp.status_code == 409


def test_follow_unauthenticated(client, register_user):
    """POST follow without auth returns 401."""
    user = register_user()
    resp = client.get("/api/v1/users/me", headers=user["headers"])
    user_id = resp.json()["id"]

    resp = client.post(f"/api/v1/social/follow/{user_id}")
    assert resp.status_code == 401


def test_unfollow_user(client, register_user):
    """DELETE /api/v1/social/follow/{user_id} removes follow."""
    user_a = register_user()
    user_b = register_user()

    resp = client.get("/api/v1/users/me", headers=user_b["headers"])
    user_b_id = resp.json()["id"]

    client.post(f"/api/v1/social/follow/{user_b_id}", headers=user_a["headers"])
    resp = client.delete(f"/api/v1/social/follow/{user_b_id}", headers=user_a["headers"])
    assert resp.status_code == 204


def test_unfollow_not_following(client, register_user):
    """DELETE follow when not following returns 404."""
    user_a = register_user()
    user_b = register_user()

    resp = client.get("/api/v1/users/me", headers=user_b["headers"])
    user_b_id = resp.json()["id"]

    resp = client.delete(f"/api/v1/social/follow/{user_b_id}", headers=user_a["headers"])
    assert resp.status_code == 404


def test_list_followers(client, register_user):
    """GET /api/v1/social/followers returns users who follow me."""
    user_a = register_user()
    user_b = register_user()

    resp = client.get("/api/v1/users/me", headers=user_b["headers"])
    user_b_id = resp.json()["id"]

    # A follows B
    client.post(f"/api/v1/social/follow/{user_b_id}", headers=user_a["headers"])

    resp = client.get("/api/v1/social/followers", headers=user_b["headers"])
    assert resp.status_code == 200
    data = resp.json()
    assert isinstance(data, list)
    assert len(data) >= 1
    assert data[0]["follower_username"] == user_a["username"]


def test_list_following(client, register_user):
    """GET /api/v1/social/following returns users I follow."""
    user_a = register_user()
    user_b = register_user()

    resp = client.get("/api/v1/users/me", headers=user_b["headers"])
    user_b_id = resp.json()["id"]

    # A follows B
    client.post(f"/api/v1/social/follow/{user_b_id}", headers=user_a["headers"])

    resp = client.get("/api/v1/social/following", headers=user_a["headers"])
    assert resp.status_code == 200
    data = resp.json()
    assert isinstance(data, list)
    assert len(data) >= 1
    assert data[0]["following_username"] == user_b["username"]


def test_feed_empty(client, register_user):
    """GET /api/v1/social/feed returns empty list when not following anyone."""
    user = register_user()
    resp = client.get("/api/v1/social/feed", headers=user["headers"])
    assert resp.status_code == 200
    assert resp.json() == []


def test_feed_unauthenticated(client):
    """GET feed without auth returns 401."""
    resp = client.get("/api/v1/social/feed")
    assert resp.status_code == 401


# ══════════════════════════════════════════════════════════════════════════
# COACHING: Invites, Respond, Lists
# ══════════════════════════════════════════════════════════════════════════


def test_coaching_invite(client, register_user):
    """POST /api/v1/coaching/invite creates a pending link."""
    coach = register_user()
    athlete = register_user()

    resp = client.post("/api/v1/coaching/invite", json={
        "athlete_username": athlete["username"],
    }, headers=coach["headers"])
    assert resp.status_code == 201
    data = resp.json()
    assert data["status"] == "pending"
    assert data["coach_username"] == coach["username"]
    assert data["athlete_username"] == athlete["username"]
    assert "id" in data
    assert "created_at" in data


def test_coaching_invite_self_rejected(client, register_user):
    """Coach cannot invite themselves."""
    user = register_user()
    resp = client.post("/api/v1/coaching/invite", json={
        "athlete_username": user["username"],
    }, headers=user["headers"])
    assert resp.status_code == 422


def test_coaching_invite_duplicate_rejected(client, register_user):
    """Duplicate invite returns 409."""
    coach = register_user()
    athlete = register_user()

    client.post("/api/v1/coaching/invite", json={
        "athlete_username": athlete["username"],
    }, headers=coach["headers"])
    resp = client.post("/api/v1/coaching/invite", json={
        "athlete_username": athlete["username"],
    }, headers=coach["headers"])
    assert resp.status_code == 409


def test_coaching_invite_user_not_found(client, register_user):
    """Invite to nonexistent user returns 404."""
    coach = register_user()
    resp = client.post("/api/v1/coaching/invite", json={
        "athlete_username": "nonexistent_user_xyz",
    }, headers=coach["headers"])
    assert resp.status_code == 404


def test_coaching_invite_unauthenticated(client):
    """POST invite without auth returns 401."""
    resp = client.post("/api/v1/coaching/invite", json={
        "athlete_username": "someone",
    })
    assert resp.status_code == 401


def test_coaching_accept_invite(client, register_user):
    """POST /api/v1/coaching/respond with accept=true activates link."""
    coach = register_user()
    athlete = register_user()

    resp = client.post("/api/v1/coaching/invite", json={
        "athlete_username": athlete["username"],
    }, headers=coach["headers"])
    link_id = resp.json()["id"]

    resp = client.post("/api/v1/coaching/respond", json={
        "link_id": link_id,
        "accept": True,
    }, headers=athlete["headers"])
    assert resp.status_code == 200
    assert resp.json()["status"] == "active"


def test_coaching_reject_invite(client, register_user):
    """POST /api/v1/coaching/respond with accept=false revokes link."""
    coach = register_user()
    athlete = register_user()

    resp = client.post("/api/v1/coaching/invite", json={
        "athlete_username": athlete["username"],
    }, headers=coach["headers"])
    link_id = resp.json()["id"]

    resp = client.post("/api/v1/coaching/respond", json={
        "link_id": link_id,
        "accept": False,
    }, headers=athlete["headers"])
    assert resp.status_code == 200
    assert resp.json()["status"] == "revoked"


def test_coaching_respond_already_responded(client, register_user):
    """Responding to already-responded invite returns 422."""
    coach = register_user()
    athlete = register_user()

    resp = client.post("/api/v1/coaching/invite", json={
        "athlete_username": athlete["username"],
    }, headers=coach["headers"])
    link_id = resp.json()["id"]

    client.post("/api/v1/coaching/respond", json={
        "link_id": link_id, "accept": True,
    }, headers=athlete["headers"])

    resp = client.post("/api/v1/coaching/respond", json={
        "link_id": link_id, "accept": False,
    }, headers=athlete["headers"])
    assert resp.status_code == 422


def test_coaching_list_athletes(client, register_user):
    """GET /api/v1/coaching/athletes lists coach's athletes."""
    coach = register_user()
    athlete = register_user()

    resp = client.post("/api/v1/coaching/invite", json={
        "athlete_username": athlete["username"],
    }, headers=coach["headers"])

    resp = client.get("/api/v1/coaching/athletes", headers=coach["headers"])
    assert resp.status_code == 200
    data = resp.json()
    assert isinstance(data, list)
    assert len(data) >= 1
    assert data[0]["athlete_username"] == athlete["username"]


def test_coaching_list_coaches(client, register_user):
    """GET /api/v1/coaching/coaches lists athlete's coaches."""
    coach = register_user()
    athlete = register_user()

    resp = client.post("/api/v1/coaching/invite", json={
        "athlete_username": athlete["username"],
    }, headers=coach["headers"])

    resp = client.get("/api/v1/coaching/coaches", headers=athlete["headers"])
    assert resp.status_code == 200
    data = resp.json()
    assert isinstance(data, list)
    assert len(data) >= 1
    assert data[0]["coach_username"] == coach["username"]


# ══════════════════════════════════════════════════════════════════════════
# COACHING: Athlete Sessions & Annotations
# ══════════════════════════════════════════════════════════════════════════


def _create_coaching_link(client, register_user, create_round):
    """Helper: create coach + athlete with active link and a completed session."""
    coach = register_user()
    athlete = register_user()

    # Create coaching link
    resp = client.post("/api/v1/coaching/invite", json={
        "athlete_username": athlete["username"],
    }, headers=coach["headers"])
    link_id = resp.json()["id"]
    client.post("/api/v1/coaching/respond", json={
        "link_id": link_id, "accept": True,
    }, headers=athlete["headers"])

    # Create and complete a session as athlete
    rnd = create_round(headers=athlete["headers"])
    resp = client.post("/api/v1/sessions", json={
        "template_id": rnd["id"],
    }, headers=athlete["headers"])
    assert resp.status_code == 201
    session = resp.json()
    _complete_session(client, session["id"], athlete["headers"])

    # Get athlete ID
    resp = client.get("/api/v1/users/me", headers=athlete["headers"])
    athlete_id = resp.json()["id"]

    return coach, athlete, athlete_id, session["id"]


def test_view_athlete_sessions(client, register_user, create_round):
    """GET /api/v1/coaching/athletes/{id}/sessions returns completed sessions."""
    coach, athlete, athlete_id, session_id = _create_coaching_link(
        client, register_user, create_round
    )

    resp = client.get(
        f"/api/v1/coaching/athletes/{athlete_id}/sessions",
        headers=coach["headers"],
    )
    assert resp.status_code == 200
    data = resp.json()
    assert isinstance(data, list)
    assert len(data) >= 1
    s = data[0]
    assert "id" in s
    assert "template_name" in s
    assert "total_score" in s
    assert "completed_at" in s


def test_view_athlete_sessions_no_link(client, register_user):
    """Viewing sessions without active link returns 403."""
    coach = register_user()
    athlete = register_user()

    resp = client.get("/api/v1/users/me", headers=athlete["headers"])
    athlete_id = resp.json()["id"]

    resp = client.get(
        f"/api/v1/coaching/athletes/{athlete_id}/sessions",
        headers=coach["headers"],
    )
    assert resp.status_code == 403


def test_add_annotation(client, register_user, create_round):
    """POST /api/v1/coaching/sessions/{id}/annotations adds annotation."""
    coach, athlete, athlete_id, session_id = _create_coaching_link(
        client, register_user, create_round
    )

    resp = client.post(
        f"/api/v1/coaching/sessions/{session_id}/annotations",
        json={"text": "Nice grouping on end 1", "end_number": 1},
        headers=coach["headers"],
    )
    assert resp.status_code == 201
    data = resp.json()
    assert data["text"] == "Nice grouping on end 1"
    assert data["end_number"] == 1
    assert data["session_id"] == session_id
    assert "id" in data
    assert "created_at" in data


def test_add_annotation_as_owner(client, register_user, create_round):
    """Session owner can also add annotations."""
    coach, athlete, athlete_id, session_id = _create_coaching_link(
        client, register_user, create_round
    )

    resp = client.post(
        f"/api/v1/coaching/sessions/{session_id}/annotations",
        json={"text": "I felt good on this end"},
        headers=athlete["headers"],
    )
    assert resp.status_code == 201
    assert resp.json()["author_username"] == athlete["username"]


def test_add_annotation_unauthorized(client, register_user, create_round):
    """User without coaching link cannot annotate."""
    athlete = register_user()
    stranger = register_user()

    rnd = create_round(headers=athlete["headers"])
    resp = client.post("/api/v1/sessions", json={
        "template_id": rnd["id"],
    }, headers=athlete["headers"])
    session = resp.json()
    _complete_session(client, session["id"], athlete["headers"])

    resp = client.post(
        f"/api/v1/coaching/sessions/{session['id']}/annotations",
        json={"text": "Should not work"},
        headers=stranger["headers"],
    )
    assert resp.status_code == 403


def test_list_annotations(client, register_user, create_round):
    """GET /api/v1/coaching/sessions/{id}/annotations lists annotations."""
    coach, athlete, athlete_id, session_id = _create_coaching_link(
        client, register_user, create_round
    )

    # Add two annotations
    client.post(
        f"/api/v1/coaching/sessions/{session_id}/annotations",
        json={"text": "First note"},
        headers=coach["headers"],
    )
    client.post(
        f"/api/v1/coaching/sessions/{session_id}/annotations",
        json={"text": "Second note", "end_number": 1, "arrow_number": 2},
        headers=athlete["headers"],
    )

    resp = client.get(
        f"/api/v1/coaching/sessions/{session_id}/annotations",
        headers=coach["headers"],
    )
    assert resp.status_code == 200
    data = resp.json()
    assert len(data) == 2
    assert data[0]["text"] == "First note"
    assert data[1]["text"] == "Second note"
    assert data[1]["end_number"] == 1
    assert data[1]["arrow_number"] == 2


def test_list_annotations_unauthorized(client, register_user, create_round):
    """User without coaching link cannot list annotations."""
    athlete = register_user()
    stranger = register_user()

    rnd = create_round(headers=athlete["headers"])
    resp = client.post("/api/v1/sessions", json={
        "template_id": rnd["id"],
    }, headers=athlete["headers"])
    session = resp.json()
    _complete_session(client, session["id"], athlete["headers"])

    resp = client.get(
        f"/api/v1/coaching/sessions/{session['id']}/annotations",
        headers=stranger["headers"],
    )
    assert resp.status_code == 403
