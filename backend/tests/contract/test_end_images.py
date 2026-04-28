"""
Contract tests for /api/v1/scoring/{sessionId}/... (end images) endpoints.

These validate status codes and response shapes against a live server.
"""

import io
import uuid


def _submit_end(client, session_id, stage_id, arrows, headers):
    """Submit an end of arrows to get an end_id."""
    return client.post(f"/api/v1/sessions/{session_id}/ends", json={
        "stage_id": stage_id,
        "arrows": [{"score_value": v} for v in arrows],
    }, headers=headers)


def _create_session_with_end(client, auth_headers, create_round):
    """Helper: create a session, submit one end, return (session, end, stage, round)."""
    rnd = create_round()
    resp = client.post("/api/v1/sessions", json={
        "template_id": rnd["id"],
    }, headers=auth_headers)
    assert resp.status_code == 201
    session = resp.json()

    stage_id = rnd["stages"][0]["id"]
    end_resp = _submit_end(client, session["id"], stage_id, ["10", "9", "8"], auth_headers)
    assert end_resp.status_code in (200, 201)

    # Reload session to get ends
    get_resp = client.get(f"/api/v1/sessions/{session['id']}", headers=auth_headers)
    session = get_resp.json()

    return session, session["ends"][0], stage_id, rnd


def _sample_jpeg():
    """Return minimal valid JPEG bytes."""
    # Minimal JPEG: SOI + APP0 marker + EOI
    return bytes([
        0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46,
        0x49, 0x46, 0x00, 0x01, 0x01, 0x00, 0x00, 0x01,
        0x00, 0x01, 0x00, 0x00, 0xFF, 0xD9,
    ])


def _upload_image(client, session_id, end_id, headers, image_data=None):
    """Upload an image to an end."""
    image_data = image_data or _sample_jpeg()
    files = {
        "image": ("target.jpg", io.BytesIO(image_data), "image/jpeg"),
    }
    return client.post(
        f"/api/v1/scoring/{session_id}/ends/{end_id}/images",
        files=files,
        headers=headers,
    )


# ── Upload ────────────────────────────────────────────────────────────


def test_upload_image(client, auth_headers, create_round):
    """POST /scoring/{sessionId}/ends/{endId}/images uploads an image."""
    session, end, _, rnd = _create_session_with_end(client, auth_headers, create_round)

    resp = _upload_image(client, session["id"], end["id"], auth_headers)
    assert resp.status_code == 201
    data = resp.json()
    assert "id" in data
    assert data["end_id"] == end["id"]
    assert data["session_id"] == session["id"]
    assert data["content_type"] == "image/jpeg"
    assert data["file_size"] > 0
    assert "created_at" in data

    # Cleanup
    client.delete(f"/api/v1/scoring/{session['id']}/images/{data['id']}", headers=auth_headers)
    client.post(f"/api/v1/sessions/{session['id']}/abandon", headers=auth_headers)
    client.delete(f"/api/v1/sessions/{session['id']}", headers=auth_headers)


def test_upload_image_session_not_found(client, auth_headers, create_round):
    """POST upload returns 404 for non-existent session."""
    fake_session = str(uuid.uuid4())
    fake_end = str(uuid.uuid4())
    resp = _upload_image(client, fake_session, fake_end, auth_headers)
    assert resp.status_code == 404


def test_upload_image_end_not_in_session(client, auth_headers, create_round):
    """POST upload returns 404 when end doesn't belong to session."""
    session, end, _, rnd = _create_session_with_end(client, auth_headers, create_round)
    fake_end = str(uuid.uuid4())

    resp = _upload_image(client, session["id"], fake_end, auth_headers)
    assert resp.status_code == 404

    # Cleanup
    client.post(f"/api/v1/sessions/{session['id']}/abandon", headers=auth_headers)
    client.delete(f"/api/v1/sessions/{session['id']}", headers=auth_headers)


def test_upload_image_no_auth(client, create_round, auth_headers):
    """POST upload without auth returns 401."""
    session, end, _, rnd = _create_session_with_end(client, auth_headers, create_round)

    resp = _upload_image(client, session["id"], end["id"], headers={})
    assert resp.status_code == 401

    # Cleanup
    client.post(f"/api/v1/sessions/{session['id']}/abandon", headers=auth_headers)
    client.delete(f"/api/v1/sessions/{session['id']}", headers=auth_headers)


def test_upload_image_wrong_user(client, register_user, create_round):
    """POST upload returns 404 when session belongs to different user."""
    user1 = register_user()
    user2 = register_user()

    # User1 creates a session + end
    rnd = create_round(headers=user1["headers"])
    resp = client.post("/api/v1/sessions", json={
        "template_id": rnd["id"],
    }, headers=user1["headers"])
    assert resp.status_code == 201
    session = resp.json()

    stage_id = rnd["stages"][0]["id"]
    _submit_end(client, session["id"], stage_id, ["10", "9", "8"], user1["headers"])
    get_resp = client.get(f"/api/v1/sessions/{session['id']}", headers=user1["headers"])
    end_id = get_resp.json()["ends"][0]["id"]

    # User2 tries to upload to user1's session
    resp = _upload_image(client, session["id"], end_id, user2["headers"])
    assert resp.status_code == 404

    # Cleanup
    client.post(f"/api/v1/sessions/{session['id']}/abandon", headers=user1["headers"])
    client.delete(f"/api/v1/sessions/{session['id']}", headers=user1["headers"])


# ── Get Image ─────────────────────────────────────────────────────────


def test_get_image(client, auth_headers, create_round):
    """GET /scoring/{sessionId}/images/{imageId} returns binary image data."""
    session, end, _, rnd = _create_session_with_end(client, auth_headers, create_round)
    jpeg_data = _sample_jpeg()

    upload_resp = _upload_image(client, session["id"], end["id"], auth_headers, jpeg_data)
    assert upload_resp.status_code == 201
    image_id = upload_resp.json()["id"]

    resp = client.get(
        f"/api/v1/scoring/{session['id']}/images/{image_id}",
        headers=auth_headers,
    )
    assert resp.status_code == 200
    assert resp.headers["content-type"] == "image/jpeg"
    assert resp.content == jpeg_data

    # Cleanup
    client.delete(f"/api/v1/scoring/{session['id']}/images/{image_id}", headers=auth_headers)
    client.post(f"/api/v1/sessions/{session['id']}/abandon", headers=auth_headers)
    client.delete(f"/api/v1/sessions/{session['id']}", headers=auth_headers)


def test_get_image_not_found(client, auth_headers, create_round):
    """GET image returns 404 for non-existent image."""
    session, _, _, _ = _create_session_with_end(client, auth_headers, create_round)
    fake_image = str(uuid.uuid4())

    resp = client.get(
        f"/api/v1/scoring/{session['id']}/images/{fake_image}",
        headers=auth_headers,
    )
    assert resp.status_code == 404

    # Cleanup
    client.post(f"/api/v1/sessions/{session['id']}/abandon", headers=auth_headers)
    client.delete(f"/api/v1/sessions/{session['id']}", headers=auth_headers)


# ── List by End ───────────────────────────────────────────────────────


def test_list_images_by_end(client, auth_headers, create_round):
    """GET /scoring/{sessionId}/ends/{endId}/images returns images for an end."""
    session, end, _, rnd = _create_session_with_end(client, auth_headers, create_round)

    # Upload two images
    up1 = _upload_image(client, session["id"], end["id"], auth_headers)
    up2 = _upload_image(client, session["id"], end["id"], auth_headers)
    assert up1.status_code == 201
    assert up2.status_code == 201

    resp = client.get(
        f"/api/v1/scoring/{session['id']}/ends/{end['id']}/images",
        headers=auth_headers,
    )
    assert resp.status_code == 200
    data = resp.json()
    assert isinstance(data, list)
    assert len(data) == 2
    assert all("id" in img for img in data)
    assert all(img["end_id"] == end["id"] for img in data)

    # Cleanup
    for img in data:
        client.delete(f"/api/v1/scoring/{session['id']}/images/{img['id']}", headers=auth_headers)
    client.post(f"/api/v1/sessions/{session['id']}/abandon", headers=auth_headers)
    client.delete(f"/api/v1/sessions/{session['id']}", headers=auth_headers)


def test_list_images_by_end_empty(client, auth_headers, create_round):
    """GET list by end returns empty array when no images."""
    session, end, _, rnd = _create_session_with_end(client, auth_headers, create_round)

    resp = client.get(
        f"/api/v1/scoring/{session['id']}/ends/{end['id']}/images",
        headers=auth_headers,
    )
    assert resp.status_code == 200
    assert resp.json() == []

    # Cleanup
    client.post(f"/api/v1/sessions/{session['id']}/abandon", headers=auth_headers)
    client.delete(f"/api/v1/sessions/{session['id']}", headers=auth_headers)


# ── List by Session ───────────────────────────────────────────────────


def test_list_images_by_session(client, auth_headers, create_round):
    """GET /scoring/{sessionId}/images returns all images for a session."""
    session, end, _, rnd = _create_session_with_end(client, auth_headers, create_round)

    _upload_image(client, session["id"], end["id"], auth_headers)

    resp = client.get(
        f"/api/v1/scoring/{session['id']}/images",
        headers=auth_headers,
    )
    assert resp.status_code == 200
    data = resp.json()
    assert isinstance(data, list)
    assert len(data) == 1
    assert data[0]["session_id"] == session["id"]

    # Cleanup
    for img in data:
        client.delete(f"/api/v1/scoring/{session['id']}/images/{img['id']}", headers=auth_headers)
    client.post(f"/api/v1/sessions/{session['id']}/abandon", headers=auth_headers)
    client.delete(f"/api/v1/sessions/{session['id']}", headers=auth_headers)


# ── Delete ────────────────────────────────────────────────────────────


def test_delete_image(client, auth_headers, create_round):
    """DELETE /scoring/{sessionId}/images/{imageId} removes the image."""
    session, end, _, rnd = _create_session_with_end(client, auth_headers, create_round)

    upload_resp = _upload_image(client, session["id"], end["id"], auth_headers)
    assert upload_resp.status_code == 201
    image_id = upload_resp.json()["id"]

    resp = client.delete(
        f"/api/v1/scoring/{session['id']}/images/{image_id}",
        headers=auth_headers,
    )
    assert resp.status_code == 204

    # Verify it's gone
    get_resp = client.get(
        f"/api/v1/scoring/{session['id']}/images/{image_id}",
        headers=auth_headers,
    )
    assert get_resp.status_code == 404

    # Cleanup
    client.post(f"/api/v1/sessions/{session['id']}/abandon", headers=auth_headers)
    client.delete(f"/api/v1/sessions/{session['id']}", headers=auth_headers)


def test_delete_image_not_found(client, auth_headers, create_round):
    """DELETE image returns 404 for non-existent image."""
    session, _, _, _ = _create_session_with_end(client, auth_headers, create_round)
    fake_image = str(uuid.uuid4())

    resp = client.delete(
        f"/api/v1/scoring/{session['id']}/images/{fake_image}",
        headers=auth_headers,
    )
    assert resp.status_code == 404

    # Cleanup
    client.post(f"/api/v1/sessions/{session['id']}/abandon", headers=auth_headers)
    client.delete(f"/api/v1/sessions/{session['id']}", headers=auth_headers)
