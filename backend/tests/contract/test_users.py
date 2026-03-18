"""
Contract tests for /api/v1/users endpoints.

Covers:
- PATCH /api/v1/users/me (profile update)
- POST /api/v1/users/me/avatar (file upload)
- POST /api/v1/users/me/avatar-url (URL upload)
- DELETE /api/v1/users/me/avatar
- GET /api/v1/users/{username} (public profile)
"""

import io


# ── Profile Update ─────────────────────────────────────────────────────


def test_update_profile(client, register_user):
    """PATCH /users/me updates profile fields and returns full user."""
    user = register_user()
    resp = client.patch("/api/v1/users/me", json={
        "display_name": "Test Archer",
        "bow_type": "recurve",
        "classification": "Archer 3rd Class",
        "bio": "I love archery",
        "profile_public": True,
    }, headers=user["headers"])
    assert resp.status_code == 200
    data = resp.json()
    assert data["display_name"] == "Test Archer"
    assert data["bow_type"] == "recurve"
    assert data["classification"] == "Archer 3rd Class"
    assert data["bio"] == "I love archery"
    assert data["profile_public"] is True
    # Should still have core user fields
    assert data["username"] == user["username"]
    assert data["email"] == user["email"]


def test_update_profile_partial(client, register_user):
    """PATCH with a single field only updates that field."""
    user = register_user()
    # Set initial values
    client.patch("/api/v1/users/me", json={
        "display_name": "Original",
        "bow_type": "compound",
    }, headers=user["headers"])

    # Update only bow_type
    resp = client.patch("/api/v1/users/me", json={
        "bow_type": "recurve",
    }, headers=user["headers"])
    assert resp.status_code == 200
    data = resp.json()
    assert data["bow_type"] == "recurve"
    assert data["display_name"] == "Original"  # unchanged


def test_update_profile_clear_fields(client, register_user):
    """PATCH with null values clears optional fields."""
    user = register_user()
    # Set values first
    client.patch("/api/v1/users/me", json={
        "display_name": "Archer",
        "bio": "Some bio",
    }, headers=user["headers"])

    # Clear them
    resp = client.patch("/api/v1/users/me", json={
        "display_name": None,
        "bio": None,
    }, headers=user["headers"])
    assert resp.status_code == 200
    data = resp.json()
    assert data["display_name"] is None
    assert data["bio"] is None


def test_update_profile_unauthenticated(client):
    resp = client.patch("/api/v1/users/me", json={
        "display_name": "Hacker",
    })
    assert resp.status_code == 401


# ── Avatar Upload (File) ──────────────────────────────────────────────


def _tiny_png():
    """Return bytes for a minimal valid 1x1 PNG."""
    import struct
    import zlib

    def chunk(chunk_type, data):
        c = chunk_type + data
        crc = struct.pack(">I", zlib.crc32(c) & 0xFFFFFFFF)
        return struct.pack(">I", len(data)) + c + crc

    sig = b"\x89PNG\r\n\x1a\n"
    ihdr = chunk(b"IHDR", struct.pack(">IIBBBBB", 1, 1, 8, 2, 0, 0, 0))
    raw = zlib.compress(b"\x00\x00\x00\x00")
    idat = chunk(b"IDAT", raw)
    iend = chunk(b"IEND", b"")
    return sig + ihdr + idat + iend


def test_upload_avatar(client, register_user):
    """POST /users/me/avatar with a PNG sets the avatar field."""
    user = register_user()
    png_bytes = _tiny_png()
    resp = client.post(
        "/api/v1/users/me/avatar",
        files={"file": ("avatar.png", io.BytesIO(png_bytes), "image/png")},
        headers=user["headers"],
    )
    assert resp.status_code == 200
    data = resp.json()
    assert data["avatar"] is not None
    assert data["avatar"].startswith("data:image/png;base64,")


def test_upload_avatar_invalid_type(client, register_user):
    """Reject non-image file types."""
    user = register_user()
    resp = client.post(
        "/api/v1/users/me/avatar",
        files={"file": ("doc.txt", io.BytesIO(b"hello"), "text/plain")},
        headers=user["headers"],
    )
    assert resp.status_code == 400


def test_upload_avatar_unauthenticated(client):
    resp = client.post(
        "/api/v1/users/me/avatar",
        files={"file": ("avatar.png", io.BytesIO(b"\x89PNG"), "image/png")},
    )
    assert resp.status_code == 401


# ── Avatar Delete ─────────────────────────────────────────────────────


def test_delete_avatar(client, register_user):
    """DELETE /users/me/avatar removes avatar."""
    user = register_user()
    # Upload first
    png_bytes = _tiny_png()
    client.post(
        "/api/v1/users/me/avatar",
        files={"file": ("avatar.png", io.BytesIO(png_bytes), "image/png")},
        headers=user["headers"],
    )

    # Delete
    resp = client.delete("/api/v1/users/me/avatar", headers=user["headers"])
    assert resp.status_code == 200
    data = resp.json()
    assert data["avatar"] is None


def test_delete_avatar_when_none(client, register_user):
    """DELETE avatar when none set should still succeed."""
    user = register_user()
    resp = client.delete("/api/v1/users/me/avatar", headers=user["headers"])
    assert resp.status_code == 200
    assert resp.json()["avatar"] is None


# ── Public Profile ────────────────────────────────────────────────────


def test_public_profile(client, register_user):
    """GET /users/{username} returns profile when public."""
    user = register_user()
    # Make profile public
    client.patch("/api/v1/users/me", json={
        "profile_public": True,
        "display_name": "Public Archer",
        "bow_type": "recurve",
        "bio": "I shoot arrows",
    }, headers=user["headers"])

    resp = client.get(f"/api/v1/users/{user['username']}")
    assert resp.status_code == 200
    data = resp.json()
    assert data["username"] == user["username"]
    assert data["display_name"] == "Public Archer"
    assert data["bow_type"] == "recurve"
    assert data["bio"] == "I shoot arrows"
    # Should have stats fields
    assert "total_sessions" in data
    assert "completed_sessions" in data
    assert "total_arrows" in data
    assert "total_x_count" in data
    assert "recent_sessions" in data
    assert isinstance(data["recent_sessions"], list)


def test_public_profile_private(client, register_user):
    """GET /users/{username} returns 404 for private profiles."""
    user = register_user()
    # profile_public defaults to false
    resp = client.get(f"/api/v1/users/{user['username']}")
    assert resp.status_code == 404


def test_public_profile_not_found(client):
    """GET /users/{username} returns 404 for nonexistent user."""
    resp = client.get("/api/v1/users/nonexistent_user_xyz_999")
    assert resp.status_code == 404


def test_public_profile_no_auth_required(client, register_user):
    """Public profiles are accessible without authentication."""
    user = register_user()
    client.patch("/api/v1/users/me", json={
        "profile_public": True,
    }, headers=user["headers"])

    # No auth headers
    resp = client.get(f"/api/v1/users/{user['username']}")
    assert resp.status_code == 200
