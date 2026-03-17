"""
Contract tests for /api/v1/equipment endpoints.

These validate status codes and response shapes against a live server.
"""

import uuid


# ── List Equipment ─────────────────────────────────────────────────────


def test_list_equipment_empty(client, auth_headers):
    """GET /api/v1/equipment for a fresh user returns empty list."""
    resp = client.get("/api/v1/equipment", headers=auth_headers)
    assert resp.status_code == 200
    assert resp.json() == []


def test_list_equipment_unauthenticated(client):
    """GET /api/v1/equipment without auth returns 401."""
    resp = client.get("/api/v1/equipment")
    assert resp.status_code == 401


def test_list_equipment_with_items(client, auth_headers, create_equipment):
    """GET /api/v1/equipment includes created items."""
    eq = create_equipment()
    resp = client.get("/api/v1/equipment", headers=auth_headers)
    assert resp.status_code == 200
    data = resp.json()
    ids = [e["id"] for e in data]
    assert eq["id"] in ids


# ── Create Equipment ───────────────────────────────────────────────────


def test_create_equipment(client, auth_headers, unique):
    """POST /api/v1/equipment creates an equipment item."""
    payload = {
        "category": "riser",
        "name": unique("equip"),
        "brand": "Hoyt",
        "model": "Formula Xi",
        "specs": {"draw_weight": "30lbs"},
        "notes": "Test riser",
    }
    resp = client.post("/api/v1/equipment", json=payload, headers=auth_headers)
    assert resp.status_code == 201
    data = resp.json()
    assert data["category"] == "riser"
    assert data["name"] == payload["name"]
    assert data["brand"] == "Hoyt"
    assert data["model"] == "Formula Xi"
    assert data["specs"] == {"draw_weight": "30lbs"}
    assert "id" in data
    assert "created_at" in data

    # Cleanup
    client.delete(f"/api/v1/equipment/{data['id']}", headers=auth_headers)


def test_create_equipment_minimal(client, auth_headers, unique):
    """POST /api/v1/equipment with only required fields."""
    payload = {
        "category": "arrows",
        "name": unique("arrows"),
    }
    resp = client.post("/api/v1/equipment", json=payload, headers=auth_headers)
    assert resp.status_code == 201
    data = resp.json()
    assert data["category"] == "arrows"
    assert data["brand"] is None
    assert data["model"] is None

    # Cleanup
    client.delete(f"/api/v1/equipment/{data['id']}", headers=auth_headers)


def test_create_equipment_unauthenticated(client):
    """POST /api/v1/equipment without auth returns 401."""
    resp = client.post("/api/v1/equipment", json={
        "category": "riser",
        "name": "Sneaky",
    })
    assert resp.status_code == 401


def test_create_equipment_invalid_category(client, auth_headers, unique):
    """POST /api/v1/equipment with invalid category returns 422."""
    resp = client.post("/api/v1/equipment", json={
        "category": "jetpack",
        "name": unique("bad"),
    }, headers=auth_headers)
    assert resp.status_code == 422


# ── Get Equipment ──────────────────────────────────────────────────────


def test_get_equipment(client, auth_headers, create_equipment):
    """GET /api/v1/equipment/{id} returns the item."""
    eq = create_equipment()
    resp = client.get(f"/api/v1/equipment/{eq['id']}", headers=auth_headers)
    assert resp.status_code == 200
    data = resp.json()
    assert data["id"] == eq["id"]
    assert data["name"] == eq["name"]
    assert data["category"] == eq["category"]


def test_get_equipment_not_found(client, auth_headers):
    """GET /api/v1/equipment/{id} with bogus ID returns 404."""
    fake_id = str(uuid.uuid4())
    resp = client.get(f"/api/v1/equipment/{fake_id}", headers=auth_headers)
    assert resp.status_code == 404


def test_get_equipment_unauthenticated(client, create_equipment):
    """GET /api/v1/equipment/{id} without auth returns 401."""
    eq = create_equipment()
    resp = client.get(f"/api/v1/equipment/{eq['id']}")
    assert resp.status_code == 401


# ── Update Equipment ───────────────────────────────────────────────────


def test_update_equipment(client, auth_headers, create_equipment):
    """PUT /api/v1/equipment/{id} updates the item."""
    eq = create_equipment()
    resp = client.put(f"/api/v1/equipment/{eq['id']}", json={
        "name": eq["name"] + " Updated",
        "brand": "Win&Win",
    }, headers=auth_headers)
    assert resp.status_code == 200
    data = resp.json()
    assert data["name"].endswith("Updated")
    assert data["brand"] == "Win&Win"
    # Unchanged fields preserved
    assert data["category"] == eq["category"]


def test_update_equipment_not_found(client, auth_headers):
    """PUT /api/v1/equipment/{id} with bogus ID returns 404."""
    fake_id = str(uuid.uuid4())
    resp = client.put(f"/api/v1/equipment/{fake_id}", json={
        "name": "Ghost",
    }, headers=auth_headers)
    assert resp.status_code == 404


# ── Delete Equipment ───────────────────────────────────────────────────


def test_delete_equipment(client, auth_headers, create_equipment):
    """DELETE /api/v1/equipment/{id} removes the item."""
    eq = create_equipment()
    resp = client.delete(f"/api/v1/equipment/{eq['id']}", headers=auth_headers)
    assert resp.status_code == 204

    # Confirm gone
    resp = client.get(f"/api/v1/equipment/{eq['id']}", headers=auth_headers)
    assert resp.status_code == 404


def test_delete_equipment_not_found(client, auth_headers):
    """DELETE /api/v1/equipment/{id} with bogus ID returns 404."""
    fake_id = str(uuid.uuid4())
    resp = client.delete(f"/api/v1/equipment/{fake_id}", headers=auth_headers)
    assert resp.status_code == 404


def test_delete_equipment_unauthenticated(client, create_equipment):
    """DELETE /api/v1/equipment/{id} without auth returns 401."""
    eq = create_equipment()
    resp = client.delete(f"/api/v1/equipment/{eq['id']}")
    assert resp.status_code == 401


# ── Stats ──────────────────────────────────────────────────────────────


def test_equipment_stats(client, auth_headers, create_equipment):
    """GET /api/v1/equipment/stats returns usage stats."""
    create_equipment()
    resp = client.get("/api/v1/equipment/stats", headers=auth_headers)
    assert resp.status_code == 200
    data = resp.json()
    assert isinstance(data, list)
    assert len(data) >= 1
    stat = data[0]
    assert "item_id" in stat
    assert "item_name" in stat
    assert "category" in stat
    assert "sessions_count" in stat
    assert "total_arrows" in stat
    assert "last_used" in stat


def test_equipment_stats_unauthenticated(client):
    """GET /api/v1/equipment/stats without auth returns 401."""
    resp = client.get("/api/v1/equipment/stats")
    assert resp.status_code == 401
