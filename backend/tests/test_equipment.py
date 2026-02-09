import pytest


async def _register_and_get_token(client):
    resp = await client.post("/api/v1/auth/register", json={
        "email": "equip@test.com", "username": "equip_user", "password": "pass1234",
    })
    return resp.json()["access_token"]


@pytest.mark.asyncio
async def test_equipment_crud(client, db_session):
    token = await _register_and_get_token(client)
    headers = {"Authorization": f"Bearer {token}"}

    # Create
    resp = await client.post("/api/v1/equipment", json={
        "category": "riser", "name": "Hoyt Formula Xi", "brand": "Hoyt", "model": "Formula Xi",
    }, headers=headers)
    assert resp.status_code == 201
    eq = resp.json()
    eq_id = eq["id"]
    assert eq["category"] == "riser"
    assert eq["name"] == "Hoyt Formula Xi"

    # List
    resp = await client.get("/api/v1/equipment", headers=headers)
    assert resp.status_code == 200
    assert len(resp.json()) == 1

    # Get
    resp = await client.get(f"/api/v1/equipment/{eq_id}", headers=headers)
    assert resp.status_code == 200
    assert resp.json()["brand"] == "Hoyt"

    # Update
    resp = await client.put(f"/api/v1/equipment/{eq_id}", json={
        "name": "Hoyt Formula Xi 25\"",
    }, headers=headers)
    assert resp.status_code == 200
    assert resp.json()["name"] == "Hoyt Formula Xi 25\""

    # Delete
    resp = await client.delete(f"/api/v1/equipment/{eq_id}", headers=headers)
    assert resp.status_code == 204

    # Verify deleted
    resp = await client.get(f"/api/v1/equipment/{eq_id}", headers=headers)
    assert resp.status_code == 404


@pytest.mark.asyncio
async def test_invalid_category(client, db_session):
    token = await _register_and_get_token(client)
    headers = {"Authorization": f"Bearer {token}"}

    resp = await client.post("/api/v1/equipment", json={
        "category": "banana", "name": "Test",
    }, headers=headers)
    assert resp.status_code == 422


@pytest.mark.asyncio
async def test_equipment_with_specs(client, db_session):
    token = await _register_and_get_token(client)
    headers = {"Authorization": f"Bearer {token}"}

    resp = await client.post("/api/v1/equipment", json={
        "category": "arrows",
        "name": "Easton X10",
        "brand": "Easton",
        "specs": {"spine": "450", "length": "28.5", "points": "120gr"},
    }, headers=headers)
    assert resp.status_code == 201
    assert resp.json()["specs"]["spine"] == "450"
