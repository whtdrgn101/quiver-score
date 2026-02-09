import pytest


@pytest.mark.asyncio
async def test_register(client):
    resp = await client.post("/api/v1/auth/register", json={
        "email": "archer@test.com",
        "username": "archer1",
        "password": "securepass123",
    })
    assert resp.status_code == 201
    data = resp.json()
    assert "access_token" in data
    assert "refresh_token" in data


@pytest.mark.asyncio
async def test_register_duplicate(client):
    await client.post("/api/v1/auth/register", json={
        "email": "dup@test.com", "username": "dupuser", "password": "pass1234",
    })
    resp = await client.post("/api/v1/auth/register", json={
        "email": "dup@test.com", "username": "dupuser2", "password": "pass1234",
    })
    assert resp.status_code == 409


@pytest.mark.asyncio
async def test_login(client):
    await client.post("/api/v1/auth/register", json={
        "email": "login@test.com", "username": "loginuser", "password": "pass1234",
    })
    resp = await client.post("/api/v1/auth/login", json={
        "username": "loginuser", "password": "pass1234",
    })
    assert resp.status_code == 200
    assert "access_token" in resp.json()


@pytest.mark.asyncio
async def test_login_wrong_password(client):
    await client.post("/api/v1/auth/register", json={
        "email": "wrong@test.com", "username": "wronguser", "password": "pass1234",
    })
    resp = await client.post("/api/v1/auth/login", json={
        "username": "wronguser", "password": "wrongpass",
    })
    assert resp.status_code == 401


@pytest.mark.asyncio
async def test_refresh_token(client):
    reg = await client.post("/api/v1/auth/register", json={
        "email": "refresh@test.com", "username": "refreshuser", "password": "pass1234",
    })
    refresh_token = reg.json()["refresh_token"]
    resp = await client.post("/api/v1/auth/refresh", json={
        "refresh_token": refresh_token,
    })
    assert resp.status_code == 200
    assert "access_token" in resp.json()


@pytest.mark.asyncio
async def test_get_me(client):
    reg = await client.post("/api/v1/auth/register", json={
        "email": "me@test.com", "username": "meuser", "password": "pass1234",
    })
    token = reg.json()["access_token"]
    resp = await client.get("/api/v1/users/me", headers={"Authorization": f"Bearer {token}"})
    assert resp.status_code == 200
    assert resp.json()["username"] == "meuser"


@pytest.mark.asyncio
async def test_change_password(client):
    reg = await client.post("/api/v1/auth/register", json={
        "email": "changepw@test.com", "username": "changepwuser", "password": "oldpass123",
    })
    token = reg.json()["access_token"]
    headers = {"Authorization": f"Bearer {token}"}

    resp = await client.post("/api/v1/auth/change-password", json={
        "current_password": "oldpass123",
        "new_password": "newpass456",
    }, headers=headers)
    assert resp.status_code == 200

    # Old password no longer works
    resp = await client.post("/api/v1/auth/login", json={
        "username": "changepwuser", "password": "oldpass123",
    })
    assert resp.status_code == 401

    # New password works
    resp = await client.post("/api/v1/auth/login", json={
        "username": "changepwuser", "password": "newpass456",
    })
    assert resp.status_code == 200


@pytest.mark.asyncio
async def test_change_password_wrong_current(client):
    reg = await client.post("/api/v1/auth/register", json={
        "email": "wrongcur@test.com", "username": "wrongcuruser", "password": "mypass123",
    })
    token = reg.json()["access_token"]
    headers = {"Authorization": f"Bearer {token}"}

    resp = await client.post("/api/v1/auth/change-password", json={
        "current_password": "wrongwrong",
        "new_password": "newpass456",
    }, headers=headers)
    assert resp.status_code == 401


@pytest.mark.asyncio
async def test_register_short_password(client):
    resp = await client.post("/api/v1/auth/register", json={
        "email": "short@test.com", "username": "shortuser", "password": "short",
    })
    assert resp.status_code == 422
