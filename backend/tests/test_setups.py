import pytest


async def _register_and_get_token(client, suffix=""):
    resp = await client.post("/api/v1/auth/register", json={
        "email": f"setup{suffix}@test.com", "username": f"setup_user{suffix}", "password": "pass1234",
    })
    return resp.json()["access_token"]


async def _create_equipment(client, headers, name="Test Riser", category="riser"):
    resp = await client.post("/api/v1/equipment", json={
        "category": category, "name": name,
    }, headers=headers)
    return resp.json()["id"]


@pytest.mark.asyncio
async def test_setup_crud(client, db_session):
    token = await _register_and_get_token(client)
    headers = {"Authorization": f"Bearer {token}"}

    # Create
    resp = await client.post("/api/v1/setups", json={
        "name": "Indoor Setup", "description": "My indoor config",
        "draw_weight": 38.0, "brace_height": 8.75,
    }, headers=headers)
    assert resp.status_code == 201
    setup = resp.json()
    setup_id = setup["id"]
    assert setup["name"] == "Indoor Setup"
    assert setup["draw_weight"] == 38.0
    assert setup["equipment"] == []

    # List
    resp = await client.get("/api/v1/setups", headers=headers)
    assert resp.status_code == 200
    assert len(resp.json()) == 1
    assert resp.json()[0]["equipment_count"] == 0

    # Get
    resp = await client.get(f"/api/v1/setups/{setup_id}", headers=headers)
    assert resp.status_code == 200
    assert resp.json()["description"] == "My indoor config"

    # Update
    resp = await client.put(f"/api/v1/setups/{setup_id}", json={
        "name": "Indoor Setup v2",
    }, headers=headers)
    assert resp.status_code == 200
    assert resp.json()["name"] == "Indoor Setup v2"

    # Delete
    resp = await client.delete(f"/api/v1/setups/{setup_id}", headers=headers)
    assert resp.status_code == 204

    resp = await client.get(f"/api/v1/setups/{setup_id}", headers=headers)
    assert resp.status_code == 404


@pytest.mark.asyncio
async def test_setup_equipment_linking(client, db_session):
    token = await _register_and_get_token(client, suffix="2")
    headers = {"Authorization": f"Bearer {token}"}

    # Create setup and equipment
    resp = await client.post("/api/v1/setups", json={"name": "Competition"}, headers=headers)
    setup_id = resp.json()["id"]

    eq1_id = await _create_equipment(client, headers, "Hoyt Riser", "riser")
    eq2_id = await _create_equipment(client, headers, "WinEx Limbs", "limbs")

    # Add equipment to setup
    resp = await client.post(f"/api/v1/setups/{setup_id}/equipment/{eq1_id}", headers=headers)
    assert resp.status_code == 201
    assert len(resp.json()["equipment"]) == 1

    resp = await client.post(f"/api/v1/setups/{setup_id}/equipment/{eq2_id}", headers=headers)
    assert resp.status_code == 201
    assert len(resp.json()["equipment"]) == 2

    # Duplicate link should fail
    resp = await client.post(f"/api/v1/setups/{setup_id}/equipment/{eq1_id}", headers=headers)
    assert resp.status_code == 409

    # List should show equipment count
    resp = await client.get("/api/v1/setups", headers=headers)
    assert resp.json()[0]["equipment_count"] == 2

    # Remove equipment from setup
    resp = await client.delete(f"/api/v1/setups/{setup_id}/equipment/{eq1_id}", headers=headers)
    assert resp.status_code == 204

    resp = await client.get(f"/api/v1/setups/{setup_id}", headers=headers)
    assert len(resp.json()["equipment"]) == 1


@pytest.mark.asyncio
async def test_setup_cascade_delete(client, db_session):
    token = await _register_and_get_token(client, suffix="3")
    headers = {"Authorization": f"Bearer {token}"}

    resp = await client.post("/api/v1/setups", json={"name": "Temp Setup"}, headers=headers)
    setup_id = resp.json()["id"]

    eq_id = await _create_equipment(client, headers, "Arrow Set", "arrows")
    await client.post(f"/api/v1/setups/{setup_id}/equipment/{eq_id}", headers=headers)

    # Delete setup - should succeed (cascade deletes links)
    resp = await client.delete(f"/api/v1/setups/{setup_id}", headers=headers)
    assert resp.status_code == 204

    # Equipment itself should still exist
    resp = await client.get(f"/api/v1/equipment/{eq_id}", headers=headers)
    assert resp.status_code == 200
