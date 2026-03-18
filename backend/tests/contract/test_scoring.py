"""
Contract tests for /api/v1/sessions (scoring) endpoints.

These validate status codes and response shapes against a live server.
"""

import uuid


def _submit_end(client, session_id, stage_id, arrows, headers):
    """Submit an end of arrows. arrows is a list of score_value strings."""
    return client.post(f"/api/v1/sessions/{session_id}/ends", json={
        "stage_id": stage_id,
        "arrows": [{"score_value": v} for v in arrows],
    }, headers=headers)


# ── Create Session ────────────────────────────────────────────────────


def test_create_session(client, auth_headers, create_round):
    """POST /api/v1/sessions creates a session."""
    rnd = create_round()
    resp = client.post("/api/v1/sessions", json={
        "template_id": rnd["id"],
        "notes": "Test session",
        "location": "Indoor range",
        "weather": "N/A",
    }, headers=auth_headers)
    assert resp.status_code == 201
    data = resp.json()
    assert data["template_id"] == rnd["id"]
    assert data["status"] == "in_progress"
    assert data["total_score"] == 0
    assert data["total_x_count"] == 0
    assert data["total_arrows"] == 0
    assert data["notes"] == "Test session"
    assert data["location"] == "Indoor range"
    assert data["ends"] == []
    assert "id" in data
    assert "started_at" in data
    assert data["completed_at"] is None

    # Cleanup
    client.post(f"/api/v1/sessions/{data['id']}/abandon", headers=auth_headers)
    client.delete(f"/api/v1/sessions/{data['id']}", headers=auth_headers)


def test_create_session_minimal(client, auth_headers, create_round):
    """POST /api/v1/sessions with only template_id."""
    rnd = create_round()
    resp = client.post("/api/v1/sessions", json={
        "template_id": rnd["id"],
    }, headers=auth_headers)
    assert resp.status_code == 201
    data = resp.json()
    assert data["notes"] is None
    assert data["location"] is None
    assert data["weather"] is None
    assert data["setup_profile_id"] is None

    # Cleanup
    client.post(f"/api/v1/sessions/{data['id']}/abandon", headers=auth_headers)
    client.delete(f"/api/v1/sessions/{data['id']}", headers=auth_headers)


def test_create_session_with_setup(client, auth_headers, create_round, create_setup):
    """POST /api/v1/sessions with setup_profile_id."""
    rnd = create_round()
    setup = create_setup()
    resp = client.post("/api/v1/sessions", json={
        "template_id": rnd["id"],
        "setup_profile_id": setup["id"],
    }, headers=auth_headers)
    assert resp.status_code == 201
    data = resp.json()
    assert data["setup_profile_id"] == setup["id"]

    # Cleanup
    client.post(f"/api/v1/sessions/{data['id']}/abandon", headers=auth_headers)
    client.delete(f"/api/v1/sessions/{data['id']}", headers=auth_headers)


def test_create_session_unauthenticated(client, create_round):
    """POST /api/v1/sessions without auth returns 401."""
    rnd = create_round()
    resp = client.post("/api/v1/sessions", json={
        "template_id": rnd["id"],
    })
    assert resp.status_code == 401


# ── List Sessions ─────────────────────────────────────────────────────


def test_list_sessions_empty(client, auth_headers):
    """GET /api/v1/sessions for a fresh user returns empty list."""
    resp = client.get("/api/v1/sessions", headers=auth_headers)
    assert resp.status_code == 200
    assert resp.json() == []


def test_list_sessions_unauthenticated(client):
    """GET /api/v1/sessions without auth returns 401."""
    resp = client.get("/api/v1/sessions")
    assert resp.status_code == 401


def test_list_sessions_with_items(client, auth_headers, create_session):
    """GET /api/v1/sessions includes created sessions."""
    session = create_session()
    resp = client.get("/api/v1/sessions", headers=auth_headers)
    assert resp.status_code == 200
    data = resp.json()
    ids = [s["id"] for s in data]
    assert session["id"] in ids
    match = next(s for s in data if s["id"] == session["id"])
    assert "template_name" in match
    assert "status" in match
    assert "total_score" in match
    assert "started_at" in match


def test_list_sessions_filter_by_search(client, auth_headers, create_round):
    """GET /api/v1/sessions?search= filters by notes/location."""
    rnd = create_round()
    resp = client.post("/api/v1/sessions", json={
        "template_id": rnd["id"],
        "location": "Unicorn Range Alpha",
    }, headers=auth_headers)
    assert resp.status_code == 201
    sid = resp.json()["id"]

    resp = client.get("/api/v1/sessions", params={"search": "Unicorn"}, headers=auth_headers)
    assert resp.status_code == 200
    ids = [s["id"] for s in resp.json()]
    assert sid in ids

    # Cleanup
    client.post(f"/api/v1/sessions/{sid}/abandon", headers=auth_headers)
    client.delete(f"/api/v1/sessions/{sid}", headers=auth_headers)


# ── Get Session ───────────────────────────────────────────────────────


def test_get_session(client, auth_headers, create_session):
    """GET /api/v1/sessions/{id} returns full session detail."""
    session = create_session()
    resp = client.get(f"/api/v1/sessions/{session['id']}", headers=auth_headers)
    assert resp.status_code == 200
    data = resp.json()
    assert data["id"] == session["id"]
    assert "template" in data
    assert "ends" in data
    assert "is_personal_best" in data


def test_get_session_not_found(client, auth_headers):
    """GET /api/v1/sessions/{id} with bogus ID returns 404."""
    fake_id = str(uuid.uuid4())
    resp = client.get(f"/api/v1/sessions/{fake_id}", headers=auth_headers)
    assert resp.status_code == 404


def test_get_session_unauthenticated(client, create_session):
    """GET /api/v1/sessions/{id} without auth returns 401."""
    session = create_session()
    resp = client.get(f"/api/v1/sessions/{session['id']}")
    assert resp.status_code == 401


# ── Submit End ────────────────────────────────────────────────────────


def test_submit_end(client, auth_headers, create_session):
    """POST /api/v1/sessions/{id}/ends submits arrows and returns end."""
    session = create_session()
    stage_id = session["template"]["stages"][0]["id"]

    resp = _submit_end(client, session["id"], stage_id, ["X", "10", "9"], auth_headers)
    assert resp.status_code == 201
    data = resp.json()
    assert data["end_number"] == 1
    assert data["end_total"] == 29  # X=10 + 10 + 9
    assert len(data["arrows"]) == 3
    assert data["arrows"][0]["score_value"] == "X"
    assert data["arrows"][0]["score_numeric"] == 10
    assert "id" in data
    assert "created_at" in data

    # Verify session totals updated
    resp = client.get(f"/api/v1/sessions/{session['id']}", headers=auth_headers)
    s = resp.json()
    assert s["total_score"] == 29
    assert s["total_x_count"] == 1
    assert s["total_arrows"] == 3


def test_submit_multiple_ends(client, auth_headers, create_session):
    """Submitting multiple ends increments end_number and session totals."""
    session = create_session()
    stage_id = session["template"]["stages"][0]["id"]

    _submit_end(client, session["id"], stage_id, ["10", "10", "10"], auth_headers)
    resp = _submit_end(client, session["id"], stage_id, ["9", "9", "9"], auth_headers)
    assert resp.status_code == 201
    assert resp.json()["end_number"] == 2

    resp = client.get(f"/api/v1/sessions/{session['id']}", headers=auth_headers)
    s = resp.json()
    assert s["total_score"] == 57  # 30 + 27
    assert s["total_arrows"] == 6
    assert len(s["ends"]) == 2


def test_submit_end_wrong_arrow_count(client, auth_headers, create_session):
    """POST end with wrong number of arrows returns 422."""
    session = create_session()
    stage_id = session["template"]["stages"][0]["id"]

    # Stage expects 3 arrows, send 2
    resp = _submit_end(client, session["id"], stage_id, ["10", "10"], auth_headers)
    assert resp.status_code == 422


def test_submit_end_invalid_arrow_value(client, auth_headers, create_session):
    """POST end with invalid arrow value returns 422."""
    session = create_session()
    stage_id = session["template"]["stages"][0]["id"]

    resp = _submit_end(client, session["id"], stage_id, ["10", "10", "INVALID"], auth_headers)
    assert resp.status_code == 422


def test_submit_end_session_not_in_progress(client, auth_headers, create_session):
    """POST end to a completed session returns 422."""
    session = create_session()
    # Complete the session first
    client.post(f"/api/v1/sessions/{session['id']}/complete", headers=auth_headers)

    stage_id = session["template"]["stages"][0]["id"]
    resp = _submit_end(client, session["id"], stage_id, ["10", "10", "10"], auth_headers)
    assert resp.status_code == 422


def test_submit_end_unauthenticated(client, create_session):
    """POST end without auth returns 401."""
    session = create_session()
    stage_id = session["template"]["stages"][0]["id"]
    resp = client.post(f"/api/v1/sessions/{session['id']}/ends", json={
        "stage_id": stage_id,
        "arrows": [{"score_value": "10"}, {"score_value": "10"}, {"score_value": "10"}],
    })
    assert resp.status_code == 401


# ── Undo Last End ─────────────────────────────────────────────────────


def test_undo_last_end(client, auth_headers, create_session):
    """DELETE /api/v1/sessions/{id}/ends/last undoes the last end."""
    session = create_session()
    stage_id = session["template"]["stages"][0]["id"]

    _submit_end(client, session["id"], stage_id, ["X", "10", "9"], auth_headers)
    _submit_end(client, session["id"], stage_id, ["8", "7", "6"], auth_headers)

    resp = client.delete(f"/api/v1/sessions/{session['id']}/ends/last", headers=auth_headers)
    assert resp.status_code == 200
    data = resp.json()
    assert data["total_score"] == 29  # only first end remains (10+10+9)
    assert data["total_arrows"] == 3
    assert len(data["ends"]) == 1


def test_undo_last_end_no_ends(client, auth_headers, create_session):
    """DELETE ends/last with no ends returns 422."""
    session = create_session()
    resp = client.delete(f"/api/v1/sessions/{session['id']}/ends/last", headers=auth_headers)
    assert resp.status_code == 422


def test_undo_last_end_not_in_progress(client, auth_headers, create_session):
    """DELETE ends/last on completed session returns 422."""
    session = create_session()
    client.post(f"/api/v1/sessions/{session['id']}/complete", headers=auth_headers)
    resp = client.delete(f"/api/v1/sessions/{session['id']}/ends/last", headers=auth_headers)
    assert resp.status_code == 422


# ── Complete Session ──────────────────────────────────────────────────


def test_complete_session(client, auth_headers, create_session):
    """POST /api/v1/sessions/{id}/complete marks session completed."""
    session = create_session()
    stage_id = session["template"]["stages"][0]["id"]
    _submit_end(client, session["id"], stage_id, ["X", "10", "9"], auth_headers)

    resp = client.post(f"/api/v1/sessions/{session['id']}/complete", headers=auth_headers)
    assert resp.status_code == 200
    data = resp.json()
    assert data["status"] == "completed"
    assert data["completed_at"] is not None
    assert "is_personal_best" in data


def test_complete_session_with_metadata(client, auth_headers, create_session):
    """POST complete with notes/location/weather updates them."""
    session = create_session()
    resp = client.post(f"/api/v1/sessions/{session['id']}/complete", json={
        "notes": "Great session",
        "location": "Main hall",
        "weather": "Clear",
    }, headers=auth_headers)
    assert resp.status_code == 200
    data = resp.json()
    assert data["notes"] == "Great session"
    assert data["location"] == "Main hall"
    assert data["weather"] == "Clear"


def test_complete_session_personal_best(client, auth_headers, create_round):
    """Completing a session sets is_personal_best for first score on a template."""
    rnd = create_round()
    resp = client.post("/api/v1/sessions", json={"template_id": rnd["id"]}, headers=auth_headers)
    session = resp.json()
    stage_id = session["template"]["stages"][0]["id"]
    _submit_end(client, session["id"], stage_id, ["10", "10", "10"], auth_headers)

    resp = client.post(f"/api/v1/sessions/{session['id']}/complete", headers=auth_headers)
    assert resp.status_code == 200
    assert resp.json()["is_personal_best"] is True


def test_complete_session_not_found(client, auth_headers):
    """POST complete on nonexistent session returns 404."""
    fake_id = str(uuid.uuid4())
    resp = client.post(f"/api/v1/sessions/{fake_id}/complete", headers=auth_headers)
    assert resp.status_code == 404


# ── Abandon Session ───────────────────────────────────────────────────


def test_abandon_session(client, auth_headers, create_session):
    """POST /api/v1/sessions/{id}/abandon marks session abandoned."""
    session = create_session()
    resp = client.post(f"/api/v1/sessions/{session['id']}/abandon", headers=auth_headers)
    assert resp.status_code == 200
    assert resp.json()["detail"] == "Session abandoned"

    # Verify status changed
    resp = client.get(f"/api/v1/sessions/{session['id']}", headers=auth_headers)
    assert resp.json()["status"] == "abandoned"


def test_abandon_completed_session(client, auth_headers, create_session):
    """POST abandon on completed session returns 422."""
    session = create_session()
    client.post(f"/api/v1/sessions/{session['id']}/complete", headers=auth_headers)
    resp = client.post(f"/api/v1/sessions/{session['id']}/abandon", headers=auth_headers)
    assert resp.status_code == 422


def test_abandon_session_not_found(client, auth_headers):
    """POST abandon on nonexistent session returns 404."""
    fake_id = str(uuid.uuid4())
    resp = client.post(f"/api/v1/sessions/{fake_id}/abandon", headers=auth_headers)
    assert resp.status_code == 404


# ── Delete Session ────────────────────────────────────────────────────


def test_delete_abandoned_session(client, auth_headers, create_session):
    """DELETE /api/v1/sessions/{id} works for abandoned sessions."""
    session = create_session()
    client.post(f"/api/v1/sessions/{session['id']}/abandon", headers=auth_headers)
    resp = client.delete(f"/api/v1/sessions/{session['id']}", headers=auth_headers)
    assert resp.status_code == 204

    # Confirm gone
    resp = client.get(f"/api/v1/sessions/{session['id']}", headers=auth_headers)
    assert resp.status_code == 404


def test_delete_in_progress_session(client, auth_headers, create_session):
    """DELETE /api/v1/sessions/{id} on in_progress returns 422."""
    session = create_session()
    resp = client.delete(f"/api/v1/sessions/{session['id']}", headers=auth_headers)
    assert resp.status_code == 422


def test_delete_completed_session(client, auth_headers, create_session):
    """DELETE /api/v1/sessions/{id} on completed returns 422."""
    session = create_session()
    client.post(f"/api/v1/sessions/{session['id']}/complete", headers=auth_headers)
    resp = client.delete(f"/api/v1/sessions/{session['id']}", headers=auth_headers)
    assert resp.status_code == 422


def test_delete_session_not_found(client, auth_headers):
    """DELETE /api/v1/sessions/{id} on nonexistent returns 404."""
    fake_id = str(uuid.uuid4())
    resp = client.delete(f"/api/v1/sessions/{fake_id}", headers=auth_headers)
    assert resp.status_code == 404


def test_delete_session_unauthenticated(client, create_session):
    """DELETE /api/v1/sessions/{id} without auth returns 401."""
    session = create_session()
    resp = client.delete(f"/api/v1/sessions/{session['id']}")
    assert resp.status_code == 401


# ── Stats ─────────────────────────────────────────────────────────────


def test_stats_empty(client, auth_headers):
    """GET /api/v1/sessions/stats for a fresh user returns zero stats."""
    resp = client.get("/api/v1/sessions/stats", headers=auth_headers)
    assert resp.status_code == 200
    data = resp.json()
    assert data["total_sessions"] == 0
    assert data["completed_sessions"] == 0
    assert data["total_arrows"] == 0
    assert data["total_x_count"] == 0
    assert data["personal_best_score"] is None
    assert data["avg_by_round_type"] == []
    assert data["recent_trend"] == []
    assert data["personal_records"] == []


def test_stats_with_completed_session(client, auth_headers, create_round):
    """GET /api/v1/sessions/stats reflects completed sessions."""
    rnd = create_round()
    resp = client.post("/api/v1/sessions", json={"template_id": rnd["id"]}, headers=auth_headers)
    session = resp.json()
    stage_id = session["template"]["stages"][0]["id"]
    _submit_end(client, session["id"], stage_id, ["X", "10", "9"], auth_headers)
    client.post(f"/api/v1/sessions/{session['id']}/complete", headers=auth_headers)

    resp = client.get("/api/v1/sessions/stats", headers=auth_headers)
    assert resp.status_code == 200
    data = resp.json()
    assert data["total_sessions"] >= 1
    assert data["completed_sessions"] >= 1
    assert data["total_arrows"] >= 3
    assert data["personal_best_score"] is not None
    assert len(data["avg_by_round_type"]) >= 1
    avg = data["avg_by_round_type"][0]
    assert "template_name" in avg
    assert "avg_score" in avg
    assert "count" in avg


def test_stats_unauthenticated(client):
    """GET /api/v1/sessions/stats without auth returns 401."""
    resp = client.get("/api/v1/sessions/stats")
    assert resp.status_code == 401


# ── Personal Records ──────────────────────────────────────────────────


def test_personal_records_empty(client, auth_headers):
    """GET /api/v1/sessions/personal-records for fresh user returns empty list."""
    resp = client.get("/api/v1/sessions/personal-records", headers=auth_headers)
    assert resp.status_code == 200
    assert resp.json() == []


def test_personal_records_after_completion(client, auth_headers, create_round):
    """GET /api/v1/sessions/personal-records shows PR after completing a session."""
    rnd = create_round()
    resp = client.post("/api/v1/sessions", json={"template_id": rnd["id"]}, headers=auth_headers)
    session = resp.json()
    stage_id = session["template"]["stages"][0]["id"]
    _submit_end(client, session["id"], stage_id, ["10", "10", "10"], auth_headers)
    client.post(f"/api/v1/sessions/{session['id']}/complete", headers=auth_headers)

    resp = client.get("/api/v1/sessions/personal-records", headers=auth_headers)
    assert resp.status_code == 200
    data = resp.json()
    assert len(data) >= 1
    pr = data[0]
    assert "template_name" in pr
    assert "score" in pr
    assert "max_score" in pr
    assert "achieved_at" in pr
    assert "session_id" in pr


def test_personal_records_unauthenticated(client):
    """GET /api/v1/sessions/personal-records without auth returns 401."""
    resp = client.get("/api/v1/sessions/personal-records")
    assert resp.status_code == 401


# ── Trends ────────────────────────────────────────────────────────────


def test_trends_empty(client, auth_headers):
    """GET /api/v1/sessions/trends for fresh user returns empty list."""
    resp = client.get("/api/v1/sessions/trends", headers=auth_headers)
    assert resp.status_code == 200
    assert resp.json() == []


def test_trends_after_completion(client, auth_headers, create_round):
    """GET /api/v1/sessions/trends shows data after completing a session."""
    rnd = create_round()
    resp = client.post("/api/v1/sessions", json={"template_id": rnd["id"]}, headers=auth_headers)
    session = resp.json()
    stage_id = session["template"]["stages"][0]["id"]
    _submit_end(client, session["id"], stage_id, ["10", "10", "10"], auth_headers)
    client.post(f"/api/v1/sessions/{session['id']}/complete", headers=auth_headers)

    resp = client.get("/api/v1/sessions/trends", headers=auth_headers)
    assert resp.status_code == 200
    data = resp.json()
    assert len(data) >= 1
    item = data[0]
    assert "session_id" in item
    assert "template_name" in item
    assert "total_score" in item
    assert "max_score" in item
    assert "percentage" in item
    assert "completed_at" in item


def test_trends_unauthenticated(client):
    """GET /api/v1/sessions/trends without auth returns 401."""
    resp = client.get("/api/v1/sessions/trends")
    assert resp.status_code == 401


# ── CSV Export (Single Session) ───────────────────────────────────────


def test_export_session_csv(client, auth_headers, create_round):
    """GET /api/v1/sessions/{id}/export?format=csv returns CSV."""
    rnd = create_round()
    resp = client.post("/api/v1/sessions", json={"template_id": rnd["id"]}, headers=auth_headers)
    session = resp.json()
    stage_id = session["template"]["stages"][0]["id"]
    _submit_end(client, session["id"], stage_id, ["X", "10", "9"], auth_headers)
    client.post(f"/api/v1/sessions/{session['id']}/complete", headers=auth_headers)

    resp = client.get(
        f"/api/v1/sessions/{session['id']}/export",
        params={"format": "csv"},
        headers=auth_headers,
    )
    assert resp.status_code == 200
    assert "text/csv" in resp.headers["content-type"]
    assert "session-" in resp.headers.get("content-disposition", "")
    assert len(resp.text) > 0


def test_export_session_csv_not_found(client, auth_headers):
    """GET export for nonexistent session returns 404."""
    fake_id = str(uuid.uuid4())
    resp = client.get(f"/api/v1/sessions/{fake_id}/export", headers=auth_headers)
    assert resp.status_code == 404


def test_export_session_unauthenticated(client, create_session):
    """GET export without auth returns 401."""
    session = create_session()
    resp = client.get(f"/api/v1/sessions/{session['id']}/export")
    assert resp.status_code == 401


# ── CSV Export (Multiple Sessions) ────────────────────────────────────


def test_export_sessions_csv(client, auth_headers):
    """GET /api/v1/sessions/export returns CSV."""
    resp = client.get("/api/v1/sessions/export", headers=auth_headers)
    assert resp.status_code == 200
    assert "text/csv" in resp.headers["content-type"]


def test_export_sessions_unauthenticated(client):
    """GET /api/v1/sessions/export without auth returns 401."""
    resp = client.get("/api/v1/sessions/export")
    assert resp.status_code == 401


# ── PDF Export ────────────────────────────────────────────────────────


def test_export_session_pdf(client, auth_headers, create_round):
    """GET /api/v1/sessions/{id}/export?format=pdf returns PDF."""
    rnd = create_round()
    resp = client.post("/api/v1/sessions", json={"template_id": rnd["id"]}, headers=auth_headers)
    session = resp.json()
    stage_id = session["template"]["stages"][0]["id"]
    _submit_end(client, session["id"], stage_id, ["10", "10", "10"], auth_headers)
    client.post(f"/api/v1/sessions/{session['id']}/complete", headers=auth_headers)

    resp = client.get(
        f"/api/v1/sessions/{session['id']}/export",
        params={"format": "pdf"},
        headers=auth_headers,
    )
    assert resp.status_code == 200
    assert "application/pdf" in resp.headers["content-type"]
    assert len(resp.content) > 0


# ── Sharing ───────────────────────────────────────────────────────────


def test_create_share_link(client, auth_headers, create_session):
    """POST /api/v1/share/sessions/{id} creates a share link."""
    session = create_session()
    # Complete first (sharing should work on any session, but let's use in_progress)
    resp = client.post(f"/api/v1/share/sessions/{session['id']}", headers=auth_headers)
    assert resp.status_code in (200, 201)
    data = resp.json()
    assert "share_token" in data
    assert "url" in data
    assert len(data["share_token"]) > 0


def test_create_share_link_idempotent(client, auth_headers, create_session):
    """POST share link twice returns same token."""
    session = create_session()
    resp1 = client.post(f"/api/v1/share/sessions/{session['id']}", headers=auth_headers)
    resp2 = client.post(f"/api/v1/share/sessions/{session['id']}", headers=auth_headers)
    assert resp1.json()["share_token"] == resp2.json()["share_token"]


def test_view_shared_session(client, auth_headers, create_session):
    """GET /api/v1/share/s/{token} returns shared session (no auth required)."""
    session = create_session()
    stage_id = session["template"]["stages"][0]["id"]
    _submit_end(client, session["id"], stage_id, ["10", "10", "10"], auth_headers)

    # Create share link
    resp = client.post(f"/api/v1/share/sessions/{session['id']}", headers=auth_headers)
    token = resp.json()["share_token"]

    # View without auth
    resp = client.get(f"/api/v1/share/s/{token}")
    assert resp.status_code == 200
    data = resp.json()
    assert "archer_name" in data
    assert "template_name" in data
    assert data["total_score"] == 30
    assert "ends" in data


def test_view_shared_session_invalid_token(client):
    """GET /api/v1/share/s/{token} with invalid token returns 404."""
    resp = client.get("/api/v1/share/s/nonexistent_token_123")
    assert resp.status_code == 404


def test_revoke_share_link(client, auth_headers, create_session):
    """DELETE /api/v1/share/sessions/{id} revokes the share link."""
    session = create_session()
    resp = client.post(f"/api/v1/share/sessions/{session['id']}", headers=auth_headers)
    token = resp.json()["share_token"]

    # Revoke
    resp = client.delete(f"/api/v1/share/sessions/{session['id']}", headers=auth_headers)
    assert resp.status_code == 200

    # Token should no longer work
    resp = client.get(f"/api/v1/share/s/{token}")
    assert resp.status_code == 404


def test_share_not_found(client, auth_headers):
    """POST share for nonexistent session returns 404."""
    fake_id = str(uuid.uuid4())
    resp = client.post(f"/api/v1/share/sessions/{fake_id}", headers=auth_headers)
    assert resp.status_code == 404


def test_share_unauthenticated(client, create_session):
    """POST share without auth returns 401."""
    session = create_session()
    resp = client.post(f"/api/v1/share/sessions/{session['id']}")
    assert resp.status_code == 401
