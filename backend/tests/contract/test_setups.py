"""
Contract tests for /api/v1/setups endpoints.

These validate status codes and response shapes against a live server.
"""

import uuid


# ── List Setups ────────────────────────────────────────────────────────


def test_list_setups_empty(client, auth_headers):
    """GET /api/v1/setups for a fresh user returns empty list."""
    resp = client.get("/api/v1/setups", headers=auth_headers)
    assert resp.status_code == 200
    assert resp.json() == []


def test_list_setups_unauthenticated(client):
    """GET /api/v1/setups without auth returns 401."""
    resp = client.get("/api/v1/setups")
    assert resp.status_code == 401


def test_list_setups_with_items(client, auth_headers, create_setup):
    """GET /api/v1/setups includes created setups with equipment_count."""
    setup = create_setup()
    resp = client.get("/api/v1/setups", headers=auth_headers)
    assert resp.status_code == 200
    data = resp.json()
    ids = [s["id"] for s in data]
    assert setup["id"] in ids
    match = next(s for s in data if s["id"] == setup["id"])
    assert "equipment_count" in match
    assert "name" in match
    assert "created_at" in match


# ── Create Setup ───────────────────────────────────────────────────────


def test_create_setup(client, auth_headers, unique):
    """POST /api/v1/setups creates a setup profile."""
    payload = {
        "name": unique("setup"),
        "description": "My outdoor setup",
        "brace_height": 8.75,
        "tiller": 0.125,
        "draw_weight": 32.0,
        "draw_length": 28.5,
        "arrow_foc": 11.5,
    }
    resp = client.post("/api/v1/setups", json=payload, headers=auth_headers)
    assert resp.status_code == 201
    data = resp.json()
    assert data["name"] == payload["name"]
    assert data["description"] == "My outdoor setup"
    assert data["brace_height"] == 8.75
    assert data["draw_weight"] == 32.0
    assert data["equipment"] == []
    assert "id" in data
    assert "created_at" in data

    # Cleanup
    client.delete(f"/api/v1/setups/{data['id']}", headers=auth_headers)


def test_create_setup_minimal(client, auth_headers, unique):
    """POST /api/v1/setups with only name."""
    payload = {"name": unique("minimal")}
    resp = client.post("/api/v1/setups", json=payload, headers=auth_headers)
    assert resp.status_code == 201
    data = resp.json()
    assert data["name"] == payload["name"]
    assert data["description"] is None
    assert data["brace_height"] is None

    # Cleanup
    client.delete(f"/api/v1/setups/{data['id']}", headers=auth_headers)


def test_create_setup_unauthenticated(client):
    """POST /api/v1/setups without auth returns 401."""
    resp = client.post("/api/v1/setups", json={"name": "Sneaky"})
    assert resp.status_code == 401


# ── Get Setup ──────────────────────────────────────────────────────────


def test_get_setup(client, auth_headers, create_setup):
    """GET /api/v1/setups/{id} returns setup with equipment list."""
    setup = create_setup()
    resp = client.get(f"/api/v1/setups/{setup['id']}", headers=auth_headers)
    assert resp.status_code == 200
    data = resp.json()
    assert data["id"] == setup["id"]
    assert data["name"] == setup["name"]
    assert isinstance(data["equipment"], list)


def test_get_setup_not_found(client, auth_headers):
    """GET /api/v1/setups/{id} with bogus ID returns 404."""
    fake_id = str(uuid.uuid4())
    resp = client.get(f"/api/v1/setups/{fake_id}", headers=auth_headers)
    assert resp.status_code == 404


def test_get_setup_unauthenticated(client, create_setup):
    """GET /api/v1/setups/{id} without auth returns 401."""
    setup = create_setup()
    resp = client.get(f"/api/v1/setups/{setup['id']}")
    assert resp.status_code == 401


# ── Update Setup ───────────────────────────────────────────────────────


def test_update_setup(client, auth_headers, create_setup):
    """PUT /api/v1/setups/{id} updates the setup."""
    setup = create_setup()
    resp = client.put(f"/api/v1/setups/{setup['id']}", json={
        "name": setup["name"] + " Updated",
        "draw_weight": 34.0,
    }, headers=auth_headers)
    assert resp.status_code == 200
    data = resp.json()
    assert data["name"].endswith("Updated")
    assert data["draw_weight"] == 34.0


def test_update_setup_not_found(client, auth_headers):
    """PUT /api/v1/setups/{id} with bogus ID returns 404."""
    fake_id = str(uuid.uuid4())
    resp = client.put(f"/api/v1/setups/{fake_id}", json={
        "name": "Ghost",
    }, headers=auth_headers)
    assert resp.status_code == 404


# ── Delete Setup ───────────────────────────────────────────────────────


def test_delete_setup(client, auth_headers, create_setup):
    """DELETE /api/v1/setups/{id} removes the setup."""
    setup = create_setup()
    resp = client.delete(f"/api/v1/setups/{setup['id']}", headers=auth_headers)
    assert resp.status_code == 204

    # Confirm gone
    resp = client.get(f"/api/v1/setups/{setup['id']}", headers=auth_headers)
    assert resp.status_code == 404


def test_delete_setup_not_found(client, auth_headers):
    """DELETE /api/v1/setups/{id} with bogus ID returns 404."""
    fake_id = str(uuid.uuid4())
    resp = client.delete(f"/api/v1/setups/{fake_id}", headers=auth_headers)
    assert resp.status_code == 404


def test_delete_setup_unauthenticated(client, create_setup):
    """DELETE /api/v1/setups/{id} without auth returns 401."""
    setup = create_setup()
    resp = client.delete(f"/api/v1/setups/{setup['id']}")
    assert resp.status_code == 401


# ── Equipment Linking ──────────────────────────────────────────────────


def test_add_equipment_to_setup(client, auth_headers, create_setup, create_equipment):
    """POST /api/v1/setups/{id}/equipment/{eq_id} links equipment."""
    setup = create_setup()
    eq = create_equipment()
    resp = client.post(
        f"/api/v1/setups/{setup['id']}/equipment/{eq['id']}",
        headers=auth_headers,
    )
    assert resp.status_code == 201
    data = resp.json()
    eq_ids = [e["id"] for e in data["equipment"]]
    assert eq["id"] in eq_ids


def test_add_equipment_to_setup_duplicate(client, auth_headers, create_setup, create_equipment):
    """POST duplicate equipment link returns 409."""
    setup = create_setup()
    eq = create_equipment()
    # First link
    resp = client.post(
        f"/api/v1/setups/{setup['id']}/equipment/{eq['id']}",
        headers=auth_headers,
    )
    assert resp.status_code == 201
    # Duplicate
    resp = client.post(
        f"/api/v1/setups/{setup['id']}/equipment/{eq['id']}",
        headers=auth_headers,
    )
    assert resp.status_code == 409


def test_add_nonexistent_equipment_to_setup(client, auth_headers, create_setup):
    """POST with bogus equipment ID returns 404."""
    setup = create_setup()
    fake_id = str(uuid.uuid4())
    resp = client.post(
        f"/api/v1/setups/{setup['id']}/equipment/{fake_id}",
        headers=auth_headers,
    )
    assert resp.status_code == 404


def test_remove_equipment_from_setup(client, auth_headers, create_setup, create_equipment):
    """DELETE /api/v1/setups/{id}/equipment/{eq_id} unlinks equipment."""
    setup = create_setup()
    eq = create_equipment()
    # Link first
    client.post(
        f"/api/v1/setups/{setup['id']}/equipment/{eq['id']}",
        headers=auth_headers,
    )
    # Unlink
    resp = client.delete(
        f"/api/v1/setups/{setup['id']}/equipment/{eq['id']}",
        headers=auth_headers,
    )
    assert resp.status_code == 204

    # Verify equipment is gone from setup
    resp = client.get(f"/api/v1/setups/{setup['id']}", headers=auth_headers)
    eq_ids = [e["id"] for e in resp.json()["equipment"]]
    assert eq["id"] not in eq_ids


def test_remove_equipment_not_linked(client, auth_headers, create_setup, create_equipment):
    """DELETE equipment not linked to setup returns 404."""
    setup = create_setup()
    eq = create_equipment()
    resp = client.delete(
        f"/api/v1/setups/{setup['id']}/equipment/{eq['id']}",
        headers=auth_headers,
    )
    assert resp.status_code == 404
