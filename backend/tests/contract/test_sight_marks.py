"""
Contract tests for /api/v1/sight-marks endpoints.

These validate status codes and response shapes against a live server.
"""

import uuid


# ── List Sight Marks ──────────────────────────────────────────────────


def test_list_sight_marks_empty(client, auth_headers):
    """GET /api/v1/sight-marks for a fresh user returns empty list."""
    resp = client.get("/api/v1/sight-marks", headers=auth_headers)
    assert resp.status_code == 200
    assert resp.json() == []


def test_list_sight_marks_unauthenticated(client):
    """GET /api/v1/sight-marks without auth returns 401."""
    resp = client.get("/api/v1/sight-marks")
    assert resp.status_code == 401


def test_list_sight_marks_with_items(client, auth_headers, create_sight_mark):
    """GET /api/v1/sight-marks includes created sight marks."""
    sm = create_sight_mark()
    resp = client.get("/api/v1/sight-marks", headers=auth_headers)
    assert resp.status_code == 200
    data = resp.json()
    ids = [s["id"] for s in data]
    assert sm["id"] in ids


def test_list_sight_marks_filter_by_equipment(
    client, auth_headers, create_equipment, create_sight_mark,
):
    """GET /api/v1/sight-marks?equipment_id= filters correctly."""
    eq = create_equipment()
    sm_with = create_sight_mark(equipment_id=eq["id"], distance="20m")
    create_sight_mark(distance="30m")  # no equipment

    resp = client.get(
        "/api/v1/sight-marks",
        params={"equipment_id": eq["id"]},
        headers=auth_headers,
    )
    assert resp.status_code == 200
    data = resp.json()
    ids = [s["id"] for s in data]
    assert sm_with["id"] in ids
    assert len(data) == 1


def test_list_sight_marks_filter_by_setup(
    client, auth_headers, create_setup, create_sight_mark,
):
    """GET /api/v1/sight-marks?setup_id= filters correctly."""
    setup = create_setup()
    sm_with = create_sight_mark(setup_id=setup["id"], distance="40m")
    create_sight_mark(distance="50m")  # no setup

    resp = client.get(
        "/api/v1/sight-marks",
        params={"setup_id": setup["id"]},
        headers=auth_headers,
    )
    assert resp.status_code == 200
    data = resp.json()
    ids = [s["id"] for s in data]
    assert sm_with["id"] in ids
    assert len(data) == 1


# ── Create Sight Mark ─────────────────────────────────────────────────


def test_create_sight_mark(client, auth_headers, unique):
    """POST /api/v1/sight-marks creates a sight mark with all fields."""
    payload = {
        "distance": "70m",
        "setting": "4.25 turns",
        "notes": "Outdoor field, slight wind",
        "date_recorded": "2025-06-15T14:30:00Z",
    }
    resp = client.post("/api/v1/sight-marks", json=payload, headers=auth_headers)
    assert resp.status_code == 201
    data = resp.json()
    assert data["distance"] == "70m"
    assert data["setting"] == "4.25 turns"
    assert data["notes"] == "Outdoor field, slight wind"
    assert "id" in data
    assert "created_at" in data
    assert "date_recorded" in data

    # Cleanup
    client.delete(f"/api/v1/sight-marks/{data['id']}", headers=auth_headers)


def test_create_sight_mark_minimal(client, auth_headers):
    """POST /api/v1/sight-marks with only required fields."""
    payload = {
        "distance": "18m",
        "setting": "2.0 turns",
        "date_recorded": "2025-06-01T12:00:00Z",
    }
    resp = client.post("/api/v1/sight-marks", json=payload, headers=auth_headers)
    assert resp.status_code == 201
    data = resp.json()
    assert data["distance"] == "18m"
    assert data["setting"] == "2.0 turns"
    assert data["notes"] is None
    assert data["equipment_id"] is None
    assert data["setup_id"] is None

    # Cleanup
    client.delete(f"/api/v1/sight-marks/{data['id']}", headers=auth_headers)


def test_create_sight_mark_with_equipment(
    client, auth_headers, create_equipment,
):
    """POST /api/v1/sight-marks with equipment_id links equipment."""
    eq = create_equipment()
    payload = {
        "distance": "50m",
        "setting": "3.75 turns",
        "date_recorded": "2025-07-01T10:00:00Z",
        "equipment_id": eq["id"],
    }
    resp = client.post("/api/v1/sight-marks", json=payload, headers=auth_headers)
    assert resp.status_code == 201
    data = resp.json()
    assert data["equipment_id"] == eq["id"]

    # Cleanup
    client.delete(f"/api/v1/sight-marks/{data['id']}", headers=auth_headers)


def test_create_sight_mark_with_setup(
    client, auth_headers, create_setup,
):
    """POST /api/v1/sight-marks with setup_id links setup."""
    setup = create_setup()
    payload = {
        "distance": "60m",
        "setting": "4.0 turns",
        "date_recorded": "2025-07-01T10:00:00Z",
        "setup_id": setup["id"],
    }
    resp = client.post("/api/v1/sight-marks", json=payload, headers=auth_headers)
    assert resp.status_code == 201
    data = resp.json()
    assert data["setup_id"] == setup["id"]

    # Cleanup
    client.delete(f"/api/v1/sight-marks/{data['id']}", headers=auth_headers)


def test_create_sight_mark_unauthenticated(client):
    """POST /api/v1/sight-marks without auth returns 401."""
    resp = client.post("/api/v1/sight-marks", json={
        "distance": "18m",
        "setting": "2.0",
        "date_recorded": "2025-06-01T12:00:00Z",
    })
    assert resp.status_code == 401


def test_create_sight_mark_missing_required(client, auth_headers):
    """POST /api/v1/sight-marks missing required fields returns 422."""
    resp = client.post("/api/v1/sight-marks", json={
        "distance": "18m",
        # missing setting and date_recorded
    }, headers=auth_headers)
    assert resp.status_code == 422


# ── Update Sight Mark ─────────────────────────────────────────────────


def test_update_sight_mark(client, auth_headers, create_sight_mark):
    """PUT /api/v1/sight-marks/{id} updates fields."""
    sm = create_sight_mark()
    resp = client.put(f"/api/v1/sight-marks/{sm['id']}", json={
        "distance": "20m",
        "setting": "3.0 turns",
        "notes": "Updated note",
    }, headers=auth_headers)
    assert resp.status_code == 200
    data = resp.json()
    assert data["distance"] == "20m"
    assert data["setting"] == "3.0 turns"
    assert data["notes"] == "Updated note"


def test_update_sight_mark_partial(client, auth_headers, create_sight_mark):
    """PUT /api/v1/sight-marks/{id} with partial update preserves other fields."""
    sm = create_sight_mark(distance="18m", setting="3.5 turns")
    resp = client.put(f"/api/v1/sight-marks/{sm['id']}", json={
        "notes": "Just a note",
    }, headers=auth_headers)
    assert resp.status_code == 200
    data = resp.json()
    assert data["notes"] == "Just a note"
    # Unchanged fields preserved
    assert data["distance"] == "18m"
    assert data["setting"] == "3.5 turns"


def test_update_sight_mark_not_found(client, auth_headers):
    """PUT /api/v1/sight-marks/{id} with bogus ID returns 404."""
    fake_id = str(uuid.uuid4())
    resp = client.put(f"/api/v1/sight-marks/{fake_id}", json={
        "distance": "20m",
    }, headers=auth_headers)
    assert resp.status_code == 404


# ── Delete Sight Mark ─────────────────────────────────────────────────


def test_delete_sight_mark(client, auth_headers, create_sight_mark):
    """DELETE /api/v1/sight-marks/{id} removes the item."""
    sm = create_sight_mark()
    resp = client.delete(f"/api/v1/sight-marks/{sm['id']}", headers=auth_headers)
    assert resp.status_code == 204

    # Confirm gone — list should not contain it
    resp = client.get("/api/v1/sight-marks", headers=auth_headers)
    ids = [s["id"] for s in resp.json()]
    assert sm["id"] not in ids


def test_delete_sight_mark_not_found(client, auth_headers):
    """DELETE /api/v1/sight-marks/{id} with bogus ID returns 404."""
    fake_id = str(uuid.uuid4())
    resp = client.delete(f"/api/v1/sight-marks/{fake_id}", headers=auth_headers)
    assert resp.status_code == 404


def test_delete_sight_mark_unauthenticated(client, create_sight_mark):
    """DELETE /api/v1/sight-marks/{id} without auth returns 401."""
    sm = create_sight_mark()
    resp = client.delete(f"/api/v1/sight-marks/{sm['id']}")
    assert resp.status_code == 401
