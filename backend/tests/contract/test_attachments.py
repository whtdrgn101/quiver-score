"""
Contract tests for /api/v1/attachments — the polymorphic image storage
endpoint shared by session_end, equipment, and setup owner types.

These run against a live server (Go API) and validate the full stack:
auth, owner verification, multipart upload, imaging processing, GCS
write/read, and DB persistence.
"""

import base64
import io
import uuid


# A valid 32x24 JPEG (~1 KB) so the imaging processor has real bytes to decode.
# Generated via Go image/jpeg encoder; embedded as base64 to avoid adding a
# Pillow dependency just for tests.
SAMPLE_JPEG_B64 = (
    "/9j/2wCEAAYEBQYFBAYGBQYHBwYIChAKCgkJChQODwwQFxQYGBcUFhYaHSUfGhsjHBYWICwgIyYnKSopGR"
    "8tMC0oMCUoKSgBBwcHCggKEwoKEygaFhooKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCgo"
    "KCgoKCgoKCgoKCgoKP/AABEIABgAIAMBIgACEQEDEQH/xAGiAAABBQEBAQEBAQAAAAAAAAAAAQIDBAUGBw"
    "gJCgsQAAIBAwMCBAMFBQQEAAABfQECAwAEEQUSITFBBhNRYQcicRQygZGhCCNCscEVUtHwJDNicoIJChYX"
    "GBkaJSYnKCkqNDU2Nzg5OkNERUZHSElKU1RVVldYWVpjZGVmZ2hpanN0dXZ3eHl6g4SFhoeIiYqSk5SVlp"
    "eYmZqio6Slpqeoqaqys7S1tre4ubrCw8TFxsfIycrS09TV1tfY2drh4uPk5ebn6Onq8fLz9PX29/j5+gEA"
    "AwEBAQEBAQEBAQAAAAAAAAECAwQFBgcICQoLEQACAQIEBAMEBwUEBAABAncAAQIDEQQFITEGEkFRB2FxEy"
    "IygQgUQpGhscEJIzNS8BVictEKFiQ04SXxFxgZGiYnKCkqNTY3ODk6Q0RFRkdISUpTVFVWV1hZWmNkZWZn"
    "aGlqc3R1dnd4eXqCg4SFhoeIiYqSk5SVlpeYmZqio6Slpqeoqaqys7S1tre4ubrCw8TFxsfIycrS09TV1t"
    "fY2dri4+Tl5ufo6ery8/T19vf4+fr/2gAMAwEAAhEDEQA/APCrTSuny1tWmldPlrpbTSuny1s2mldPlr0a"
    "+Y+ZyZVm22pzVppXT5a2rTSuny10tppXT5a2bTSuny149fMfM/ScqzbbUo2mldPlratNK6fLTrTtWzaV4d"
    "fETP51yrFVNNRtppXT5a2rTSuny060rZtO1ePXxEz9JyrFVNNT/9k="
)


def _sample_jpeg() -> bytes:
    return base64.b64decode(SAMPLE_JPEG_B64)


def _create_end(client, auth_headers, create_round) -> str:
    """Create a session, submit one end, return the end's id."""
    rnd = create_round()
    resp = client.post("/api/v1/sessions", json={"template_id": rnd["id"]}, headers=auth_headers)
    assert resp.status_code == 201, resp.text
    session = resp.json()
    stage_id = rnd["stages"][0]["id"]
    end_resp = client.post(
        f"/api/v1/sessions/{session['id']}/ends",
        json={"stage_id": stage_id, "arrows": [{"score_value": v} for v in ["10", "9", "8"]]},
        headers=auth_headers,
    )
    assert end_resp.status_code in (200, 201), end_resp.text
    detail = client.get(f"/api/v1/sessions/{session['id']}", headers=auth_headers).json()
    return detail["ends"][0]["id"], session["id"]


def _upload(client, headers, owner_type, owner_id, content_type="image/jpeg", body=None):
    files = {"image": ("upload.jpg", io.BytesIO(body if body is not None else _sample_jpeg()), content_type)}
    return client.post(
        f"/api/v1/attachments?owner_type={owner_type}&owner_id={owner_id}",
        files=files,
        headers=headers,
    )


# ── Upload happy paths per owner type ──────────────────────────────────


def test_upload_to_session_end(client, auth_headers, create_round):
    end_id, _ = _create_end(client, auth_headers, create_round)
    resp = _upload(client, auth_headers, "session_end", end_id)
    assert resp.status_code == 201, resp.text
    data = resp.json()
    assert data["owner_type"] == "session_end"
    assert data["owner_id"] == end_id
    assert data["content_type"] == "image/jpeg"
    assert data["full_size"] > 0 and data["thumb_size"] > 0
    # Thumb should be smaller than (or equal to) full for our tiny source — equal
    # is fine when the original is already below the thumb max dimension.
    assert data["thumb_size"] <= data["full_size"]
    assert data["width"] > 0 and data["height"] > 0


def test_upload_to_equipment(client, auth_headers, create_equipment):
    eq = create_equipment()
    resp = _upload(client, auth_headers, "equipment", eq["id"])
    assert resp.status_code == 201, resp.text
    data = resp.json()
    assert data["owner_type"] == "equipment"
    assert data["owner_id"] == eq["id"]


def test_upload_to_setup(client, auth_headers, create_setup):
    s = create_setup()
    resp = _upload(client, auth_headers, "setup", s["id"])
    assert resp.status_code == 201, resp.text
    data = resp.json()
    assert data["owner_type"] == "setup"
    assert data["owner_id"] == s["id"]


# ── Validation errors ──────────────────────────────────────────────────


def test_upload_missing_owner_type(client, auth_headers, create_equipment):
    eq = create_equipment()
    files = {"image": ("u.jpg", io.BytesIO(_sample_jpeg()), "image/jpeg")}
    resp = client.post(
        f"/api/v1/attachments?owner_id={eq['id']}", files=files, headers=auth_headers,
    )
    assert resp.status_code == 400


def test_upload_unknown_owner_type(client, auth_headers):
    resp = _upload(client, auth_headers, "bogus_type", str(uuid.uuid4()))
    assert resp.status_code == 400


def test_upload_invalid_owner_id_uuid(client, auth_headers):
    resp = _upload(client, auth_headers, "equipment", "not-a-uuid")
    assert resp.status_code == 400


def test_upload_unsupported_content_type(client, auth_headers, create_equipment):
    eq = create_equipment()
    resp = _upload(client, auth_headers, "equipment", eq["id"], content_type="image/heic")
    assert resp.status_code == 400
    assert "HEIC" in resp.text


# ── Auth + ownership ───────────────────────────────────────────────────


def test_upload_no_auth(client, create_equipment):
    # create_equipment uses auth_headers internally; we just need an owner_id to point at.
    fake_id = str(uuid.uuid4())
    files = {"image": ("u.jpg", io.BytesIO(_sample_jpeg()), "image/jpeg")}
    resp = client.post(
        f"/api/v1/attachments?owner_type=equipment&owner_id={fake_id}",
        files=files,
    )
    assert resp.status_code == 401


def test_upload_owner_belongs_to_other_user(client, register_user, create_equipment):
    user_a = register_user()
    user_b = register_user()
    # user_a's equipment, attempted upload by user_b → 404 (no leak that the owner exists).
    eq = create_equipment(headers=user_a["headers"])
    resp = _upload(client, user_b["headers"], "equipment", eq["id"])
    assert resp.status_code == 404


# ── List ───────────────────────────────────────────────────────────────


def test_list_returns_uploaded_attachments(client, auth_headers, create_equipment):
    eq = create_equipment()
    created_ids = []
    for _ in range(2):
        resp = _upload(client, auth_headers, "equipment", eq["id"])
        assert resp.status_code == 201
        created_ids.append(resp.json()["id"])

    resp = client.get(
        f"/api/v1/attachments?owner_type=equipment&owner_id={eq['id']}", headers=auth_headers,
    )
    assert resp.status_code == 200
    rows = resp.json()
    returned_ids = sorted(r["id"] for r in rows)
    assert returned_ids == sorted(created_ids)


def test_list_empty_returns_array(client, auth_headers, create_equipment):
    eq = create_equipment()
    resp = client.get(
        f"/api/v1/attachments?owner_type=equipment&owner_id={eq['id']}", headers=auth_headers,
    )
    assert resp.status_code == 200
    assert resp.json() == []


def test_list_owner_belongs_to_other_user(client, register_user, create_equipment):
    user_a = register_user()
    user_b = register_user()
    eq = create_equipment(headers=user_a["headers"])
    resp = client.get(
        f"/api/v1/attachments?owner_type=equipment&owner_id={eq['id']}", headers=user_b["headers"],
    )
    assert resp.status_code == 404


# ── Get full / thumb ───────────────────────────────────────────────────


def test_get_full_returns_jpeg_bytes(client, auth_headers, create_equipment):
    eq = create_equipment()
    upload = _upload(client, auth_headers, "equipment", eq["id"]).json()

    resp = client.get(f"/api/v1/attachments/{upload['id']}", headers=auth_headers)
    assert resp.status_code == 200
    assert resp.headers["content-type"] == "image/jpeg"
    assert resp.headers.get("etag") == f'"{upload["id"]}"'
    assert resp.content[:2] == b"\xff\xd8"  # JPEG SOI marker


def test_get_thumb_returns_jpeg_bytes(client, auth_headers, create_equipment):
    eq = create_equipment()
    upload = _upload(client, auth_headers, "equipment", eq["id"]).json()

    resp = client.get(f"/api/v1/attachments/{upload['id']}/thumb", headers=auth_headers)
    assert resp.status_code == 200
    assert resp.headers["content-type"] == "image/jpeg"
    assert resp.content[:2] == b"\xff\xd8"


def test_get_attachment_belonging_to_other_user(client, register_user, create_equipment):
    user_a = register_user()
    user_b = register_user()
    eq = create_equipment(headers=user_a["headers"])
    upload = _upload(client, user_a["headers"], "equipment", eq["id"]).json()

    resp = client.get(f"/api/v1/attachments/{upload['id']}", headers=user_b["headers"])
    assert resp.status_code == 404


# ── Delete ─────────────────────────────────────────────────────────────


def test_delete_attachment(client, auth_headers, create_equipment):
    eq = create_equipment()
    upload = _upload(client, auth_headers, "equipment", eq["id"]).json()

    resp = client.delete(f"/api/v1/attachments/{upload['id']}", headers=auth_headers)
    assert resp.status_code == 204

    # Subsequent fetch returns 404.
    resp = client.get(f"/api/v1/attachments/{upload['id']}", headers=auth_headers)
    assert resp.status_code == 404


def test_delete_nonexistent_attachment(client, auth_headers):
    resp = client.delete(f"/api/v1/attachments/{uuid.uuid4()}", headers=auth_headers)
    assert resp.status_code == 404


# ── Account-deletion prefix wipe ───────────────────────────────────────


def test_account_delete_wipes_attachments(client, register_user, create_equipment):
    """Account deletion should make the user's attachments inaccessible."""
    user = register_user()
    eq = create_equipment(headers=user["headers"])
    upload = _upload(client, user["headers"], "equipment", eq["id"]).json()

    # Delete the account.
    resp = client.post(
        "/api/v1/auth/delete-account",
        json={"confirmation": "Yes, delete ALL of my data"},
        headers=user["headers"],
    )
    assert resp.status_code == 200

    # The attachment row is gone (DB cascade); subsequent reads should 401 (token
    # may still be valid format-wise but the user is gone) or 404. Either is
    # acceptable proof that the data is no longer accessible.
    resp = client.get(f"/api/v1/attachments/{upload['id']}", headers=user["headers"])
    assert resp.status_code in (401, 404)
