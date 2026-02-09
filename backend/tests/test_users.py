import io
import pytest


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
