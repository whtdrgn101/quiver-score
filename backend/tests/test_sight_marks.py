import pytest

_user_counter = 0


async def _register_and_get_token(client, email=None, username=None):
    global _user_counter
    _user_counter += 1
    email = email or f"sight{_user_counter}@test.com"
    username = username or f"sight{_user_counter}"
    resp = await client.post("/api/v1/auth/register", json={
        "email": email, "username": username, "password": "pass1234",
    })
    return resp.json()["access_token"]


@pytest.mark.asyncio
async def test_list_sight_marks_empty(client, db_session):
    """GET /sight-marks returns empty list initially."""
    token = await _register_and_get_token(client)
    headers = {"Authorization": f"Bearer {token}"}
    resp = await client.get("/api/v1/sight-marks", headers=headers)
    assert resp.status_code == 200
    assert resp.json() == []


@pytest.mark.asyncio
async def test_create_sight_mark(client, db_session):
    """POST /sight-marks creates a sight mark."""
    token = await _register_and_get_token(client)
    headers = {"Authorization": f"Bearer {token}"}
    resp = await client.post("/api/v1/sight-marks", json={
        "distance": "18m",
        "setting": "3.5 turns",
        "notes": "Indoor range",
        "date_recorded": "2026-02-13T10:00:00Z",
    }, headers=headers)
    assert resp.status_code == 201
    data = resp.json()
    assert data["distance"] == "18m"
    assert data["setting"] == "3.5 turns"
    assert data["notes"] == "Indoor range"


@pytest.mark.asyncio
async def test_update_sight_mark(client, db_session):
    """PUT /sight-marks/{id} updates a sight mark."""
    token = await _register_and_get_token(client)
    headers = {"Authorization": f"Bearer {token}"}
    resp = await client.post("/api/v1/sight-marks", json={
        "distance": "18m",
        "setting": "3.5 turns",
        "date_recorded": "2026-02-13T10:00:00Z",
    }, headers=headers)
    sm_id = resp.json()["id"]

    resp = await client.put(f"/api/v1/sight-marks/{sm_id}", json={
        "setting": "4.0 turns",
    }, headers=headers)
    assert resp.status_code == 200
    assert resp.json()["setting"] == "4.0 turns"


@pytest.mark.asyncio
async def test_delete_sight_mark(client, db_session):
    """DELETE /sight-marks/{id} deletes a sight mark."""
    token = await _register_and_get_token(client)
    headers = {"Authorization": f"Bearer {token}"}
    resp = await client.post("/api/v1/sight-marks", json={
        "distance": "18m",
        "setting": "3.5 turns",
        "date_recorded": "2026-02-13T10:00:00Z",
    }, headers=headers)
    sm_id = resp.json()["id"]

    resp = await client.delete(f"/api/v1/sight-marks/{sm_id}", headers=headers)
    assert resp.status_code == 204

    resp = await client.get("/api/v1/sight-marks", headers=headers)
    assert len(resp.json()) == 0


@pytest.mark.asyncio
async def test_filter_sight_marks_by_equipment(client, db_session):
    """GET /sight-marks?equipment_id= filters by equipment."""
    token = await _register_and_get_token(client)
    headers = {"Authorization": f"Bearer {token}"}

    # Create an equipment item
    eq_resp = await client.post("/api/v1/equipment", json={
        "name": "My Riser",
        "category": "riser",
    }, headers=headers)
    eq_id = eq_resp.json()["id"]

    # Create sight marks - one with equipment, one without
    await client.post("/api/v1/sight-marks", json={
        "distance": "18m",
        "setting": "3.5 turns",
        "equipment_id": eq_id,
        "date_recorded": "2026-02-13T10:00:00Z",
    }, headers=headers)
    await client.post("/api/v1/sight-marks", json={
        "distance": "30m",
        "setting": "5.0 turns",
        "date_recorded": "2026-02-13T10:00:00Z",
    }, headers=headers)

    # All sight marks
    resp = await client.get("/api/v1/sight-marks", headers=headers)
    assert len(resp.json()) == 2

    # Filtered by equipment
    resp = await client.get(f"/api/v1/sight-marks?equipment_id={eq_id}", headers=headers)
    assert len(resp.json()) == 1
    assert resp.json()[0]["distance"] == "18m"


@pytest.mark.asyncio
async def test_create_sight_mark_with_setup_id(client, db_session):
    """POST /sight-marks with setup_id links to setup profile."""
    token = await _register_and_get_token(client)
    headers = {"Authorization": f"Bearer {token}"}

    # Create a setup profile
    setup_resp = await client.post("/api/v1/setups", json={"name": "Indoor Setup"}, headers=headers)
    assert setup_resp.status_code == 201
    setup_id = setup_resp.json()["id"]

    # Create sight mark linked to the setup
    resp = await client.post("/api/v1/sight-marks", json={
        "distance": "18m",
        "setting": "3.5 turns",
        "setup_id": setup_id,
        "date_recorded": "2026-02-13T10:00:00Z",
    }, headers=headers)
    assert resp.status_code == 201
    data = resp.json()
    assert data["setup_id"] == setup_id
    assert data["distance"] == "18m"


@pytest.mark.asyncio
async def test_filter_sight_marks_by_setup_id(client, db_session):
    """GET /sight-marks?setup_id= filters by setup."""
    token = await _register_and_get_token(client)
    headers = {"Authorization": f"Bearer {token}"}

    # Create a setup profile
    setup_resp = await client.post("/api/v1/setups", json={"name": "Outdoor Setup"}, headers=headers)
    setup_id = setup_resp.json()["id"]

    # Create sight marks - one with setup, one without
    await client.post("/api/v1/sight-marks", json={
        "distance": "30m",
        "setting": "5.0 turns",
        "setup_id": setup_id,
        "date_recorded": "2026-02-13T10:00:00Z",
    }, headers=headers)
    await client.post("/api/v1/sight-marks", json={
        "distance": "18m",
        "setting": "3.5 turns",
        "date_recorded": "2026-02-13T10:00:00Z",
    }, headers=headers)

    # All sight marks
    resp = await client.get("/api/v1/sight-marks", headers=headers)
    assert len(resp.json()) == 2

    # Filtered by setup
    resp = await client.get(f"/api/v1/sight-marks?setup_id={setup_id}", headers=headers)
    assert len(resp.json()) == 1
    assert resp.json()[0]["distance"] == "30m"
