import pytest

from app.seed.round_templates import seed_round_templates


async def _register_and_get_token(client, email="equip@test.com", username="equip_user"):
    resp = await client.post("/api/v1/auth/register", json={
        "email": email, "username": username, "password": "pass1234",
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


@pytest.mark.asyncio
async def test_equipment_stats_empty(client, db_session):
    token = await _register_and_get_token(client, "eqstats@test.com", "eqstats_user")
    headers = {"Authorization": f"Bearer {token}"}

    resp = await client.get("/api/v1/equipment/stats", headers=headers)
    assert resp.status_code == 200
    assert resp.json() == []


@pytest.mark.asyncio
async def test_equipment_stats_with_usage(client, db_session):
    await seed_round_templates(db_session)
    token = await _register_and_get_token(client, "equse@test.com", "equse_user")
    headers = {"Authorization": f"Bearer {token}"}

    # Create equipment
    resp = await client.post("/api/v1/equipment", json={
        "category": "riser", "name": "Hoyt GMX",
    }, headers=headers)
    eq_id = resp.json()["id"]

    # Create setup with equipment
    resp = await client.post("/api/v1/setups", json={"name": "Test Setup"}, headers=headers)
    setup_id = resp.json()["id"]
    await client.post(f"/api/v1/setups/{setup_id}/equipment/{eq_id}", headers=headers)

    # Create session with setup and complete it
    rounds = (await client.get("/api/v1/rounds")).json()
    vegas = next(r for r in rounds if r["name"] == "Vegas 300")
    stage_id = vegas["stages"][0]["id"]

    resp = await client.post("/api/v1/sessions", json={
        "template_id": vegas["id"], "setup_profile_id": setup_id,
    }, headers=headers)
    session_id = resp.json()["id"]

    await client.post(f"/api/v1/sessions/{session_id}/ends", json={
        "stage_id": stage_id,
        "arrows": [{"score_value": "X"}, {"score_value": "10"}, {"score_value": "9"}],
    }, headers=headers)
    await client.post(f"/api/v1/sessions/{session_id}/complete", headers=headers)

    # Check stats
    resp = await client.get("/api/v1/equipment/stats", headers=headers)
    assert resp.status_code == 200
    stats = resp.json()
    assert len(stats) == 1
    assert stats[0]["item_name"] == "Hoyt GMX"
    assert stats[0]["sessions_count"] == 1
    assert stats[0]["total_arrows"] == 3
