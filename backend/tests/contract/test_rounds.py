"""
Contract tests for /api/v1/rounds endpoints.

These validate status codes and response shapes against a live server.
They do NOT reach into the database or import app internals — that's
what makes them portable across Python and Go implementations.
"""

import uuid


# ── List Rounds ────────────────────────────────────────────────────────


def test_list_rounds_unauthenticated(client):
    """GET /api/v1/rounds without auth returns only official rounds."""
    resp = client.get("/api/v1/rounds")
    assert resp.status_code == 200
    data = resp.json()
    assert isinstance(data, list)
    # All returned rounds should be official
    for r in data:
        assert r["is_official"] is True


def test_list_rounds_authenticated(client, auth_headers, create_round):
    """GET /api/v1/rounds with auth includes user's custom rounds."""
    custom = create_round()
    resp = client.get("/api/v1/rounds", headers=auth_headers)
    assert resp.status_code == 200
    data = resp.json()
    ids = [r["id"] for r in data]
    assert custom["id"] in ids


# ── Create Round ───────────────────────────────────────────────────────


def test_create_round(client, auth_headers, unique):
    """POST /api/v1/rounds creates a custom round with stages."""
    payload = {
        "name": unique("create"),
        "organization": "WA",
        "description": "Test round",
        "stages": [
            {
                "name": "Half 1",
                "distance": "70m",
                "num_ends": 6,
                "arrows_per_end": 6,
                "allowed_values": ["X", "10", "9", "8", "7", "6", "5", "4", "3", "2", "1", "M"],
                "value_score_map": {
                    "X": 10, "10": 10, "9": 9, "8": 8, "7": 7,
                    "6": 6, "5": 5, "4": 4, "3": 3, "2": 2, "1": 1, "M": 0,
                },
                "max_score_per_arrow": 10,
            },
        ],
    }
    resp = client.post("/api/v1/rounds", json=payload, headers=auth_headers)
    assert resp.status_code == 201
    data = resp.json()
    assert data["name"] == payload["name"]
    assert data["organization"] == "WA"
    assert data["is_official"] is False
    assert len(data["stages"]) == 1
    stage = data["stages"][0]
    assert stage["name"] == "Half 1"
    assert stage["num_ends"] == 6
    assert stage["arrows_per_end"] == 6
    assert "id" in stage

    # Cleanup
    client.delete(f"/api/v1/rounds/{data['id']}", headers=auth_headers)


def test_create_round_missing_name(client, auth_headers):
    """POST /api/v1/rounds without name should fail with 422."""
    payload = {
        "organization": "WA",
        "stages": [
            {
                "name": "S1",
                "distance": "18m",
                "num_ends": 10,
                "arrows_per_end": 3,
                "allowed_values": ["10", "M"],
                "value_score_map": {"10": 10, "M": 0},
                "max_score_per_arrow": 10,
            },
        ],
    }
    resp = client.post("/api/v1/rounds", json=payload, headers=auth_headers)
    assert resp.status_code == 422


def test_create_round_unauthenticated(client):
    """POST /api/v1/rounds without auth should fail with 401."""
    payload = {
        "name": "Sneaky Round",
        "organization": "WA",
        "stages": [
            {
                "name": "S1",
                "distance": "18m",
                "num_ends": 10,
                "arrows_per_end": 3,
                "allowed_values": ["10", "M"],
                "value_score_map": {"10": 10, "M": 0},
                "max_score_per_arrow": 10,
            },
        ],
    }
    resp = client.post("/api/v1/rounds", json=payload)
    assert resp.status_code == 401


# ── Get Round ──────────────────────────────────────────────────────────


def test_get_round(client, auth_headers, create_round):
    """GET /api/v1/rounds/{id} returns round with stages."""
    custom = create_round()
    resp = client.get(f"/api/v1/rounds/{custom['id']}")
    assert resp.status_code == 200
    data = resp.json()
    assert data["id"] == custom["id"]
    assert data["name"] == custom["name"]
    assert isinstance(data["stages"], list)
    assert len(data["stages"]) >= 1


def test_get_round_not_found(client):
    """GET /api/v1/rounds/{id} with bogus ID returns 404."""
    fake_id = str(uuid.uuid4())
    resp = client.get(f"/api/v1/rounds/{fake_id}")
    assert resp.status_code == 404


# ── Update Round ───────────────────────────────────────────────────────


def test_update_round(client, auth_headers, create_round):
    """PUT /api/v1/rounds/{id} updates the round."""
    custom = create_round()
    updated_payload = {
        "name": custom["name"] + " Updated",
        "organization": custom["organization"],
        "description": "Updated description",
        "stages": [
            {
                "name": "New Stage",
                "distance": "20yd",
                "num_ends": 5,
                "arrows_per_end": 5,
                "allowed_values": ["X", "10", "9", "8", "7", "6", "5", "4", "3", "2", "1", "M"],
                "value_score_map": {
                    "X": 10, "10": 10, "9": 9, "8": 8, "7": 7,
                    "6": 6, "5": 5, "4": 4, "3": 3, "2": 2, "1": 1, "M": 0,
                },
                "max_score_per_arrow": 10,
            },
        ],
    }
    resp = client.put(
        f"/api/v1/rounds/{custom['id']}",
        json=updated_payload,
        headers=auth_headers,
    )
    assert resp.status_code == 200
    data = resp.json()
    assert data["name"].endswith("Updated")
    assert data["description"] == "Updated description"
    assert len(data["stages"]) == 1
    assert data["stages"][0]["name"] == "New Stage"


def test_update_official_round_forbidden(client, auth_headers):
    """PUT on an official round should return 403."""
    # Fetch an official round to try to update
    resp = client.get("/api/v1/rounds")
    assert resp.status_code == 200
    officials = [r for r in resp.json() if r["is_official"]]
    if not officials:
        # No official rounds seeded — skip
        return
    official = officials[0]
    updated_payload = {
        "name": "Hacked Name",
        "organization": official["organization"],
        "stages": [
            {
                "name": "S1",
                "distance": "18m",
                "num_ends": 10,
                "arrows_per_end": 3,
                "allowed_values": ["10", "M"],
                "value_score_map": {"10": 10, "M": 0},
                "max_score_per_arrow": 10,
            },
        ],
    }
    resp = client.put(
        f"/api/v1/rounds/{official['id']}",
        json=updated_payload,
        headers=auth_headers,
    )
    assert resp.status_code == 403


# ── Delete Round ───────────────────────────────────────────────────────


def test_delete_round(client, auth_headers, create_round):
    """DELETE /api/v1/rounds/{id} removes the user's custom round."""
    custom = create_round()
    resp = client.delete(f"/api/v1/rounds/{custom['id']}", headers=auth_headers)
    assert resp.status_code == 204

    # Confirm it's gone
    resp = client.get(f"/api/v1/rounds/{custom['id']}")
    assert resp.status_code == 404


def test_delete_official_round_forbidden(client, auth_headers):
    """DELETE on an official round should return 403."""
    resp = client.get("/api/v1/rounds")
    assert resp.status_code == 200
    officials = [r for r in resp.json() if r["is_official"]]
    if not officials:
        return
    official = officials[0]
    resp = client.delete(
        f"/api/v1/rounds/{official['id']}",
        headers=auth_headers,
    )
    assert resp.status_code == 403


def test_delete_round_unauthenticated(client, create_round):
    """DELETE without auth should return 401."""
    custom = create_round()
    resp = client.delete(f"/api/v1/rounds/{custom['id']}")
    assert resp.status_code == 401
