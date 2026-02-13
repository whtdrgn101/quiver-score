import pytest

from app.seed.round_templates import seed_round_templates


# ── Helpers ──────────────────────────────────────────────────────────────


async def _register(client, email, username):
    resp = await client.post("/api/v1/auth/register", json={
        "email": email, "username": username, "password": "pass1234",
    })
    token = resp.json()["access_token"]
    return token, {"Authorization": f"Bearer {token}"}


CUSTOM_ROUND = {
    "name": "Field Course v1",
    "organization": "Custom",
    "description": "12 targets, 3 arrows each",
    "stages": [{
        "name": "Stage 1",
        "distance": "Various",
        "num_ends": 12,
        "arrows_per_end": 3,
        "allowed_values": ["X", "10", "9", "8", "7", "6", "5", "4", "3", "2", "1", "M"],
        "value_score_map": {"X": 10, "10": 10, "9": 9, "8": 8, "7": 7, "6": 6, "5": 5, "4": 4, "3": 3, "2": 2, "1": 1, "M": 0},
        "max_score_per_arrow": 10,
    }],
}

UPDATED_ROUND = {
    "name": "Field Course v2",
    "organization": "Custom",
    "description": "14 targets now",
    "stages": [{
        "name": "Stage 1",
        "distance": "Various",
        "num_ends": 14,
        "arrows_per_end": 3,
        "allowed_values": ["X", "10", "9", "8", "7", "6", "5", "4", "3", "2", "1", "M"],
        "value_score_map": {"X": 10, "10": 10, "9": 9, "8": 8, "7": 7, "6": 6, "5": 5, "4": 4, "3": 3, "2": 2, "1": 1, "M": 0},
        "max_score_per_arrow": 10,
    }],
}


async def _create_custom_round(client, headers, data=None):
    resp = await client.post("/api/v1/rounds", json=data or CUSTOM_ROUND, headers=headers)
    assert resp.status_code == 201
    return resp.json()


async def _submit_end(client, headers, session_id, stage_id, arrows=None):
    if arrows is None:
        arrows = [{"score_value": "10"}, {"score_value": "9"}, {"score_value": "8"}]
    resp = await client.post(
        f"/api/v1/sessions/{session_id}/ends",
        json={"stage_id": stage_id, "arrows": arrows},
        headers=headers,
    )
    assert resp.status_code == 201
    return resp.json()


# ── Tests ────────────────────────────────────────────────────────────────


@pytest.mark.asyncio
async def test_edit_custom_round(client):
    _, headers = await _register(client, "edit1@test.com", "edituser1")
    custom = await _create_custom_round(client, headers)
    round_id = custom["id"]

    resp = await client.put(f"/api/v1/rounds/{round_id}", json=UPDATED_ROUND, headers=headers)
    assert resp.status_code == 200
    data = resp.json()
    assert data["name"] == "Field Course v2"
    assert data["description"] == "14 targets now"
    assert data["stages"][0]["num_ends"] == 14
    assert data["id"] == round_id  # same template ID


@pytest.mark.asyncio
async def test_edit_round_not_found(client):
    _, headers = await _register(client, "edit2@test.com", "edituser2")
    fake_id = "00000000-0000-0000-0000-000000000000"
    resp = await client.put(f"/api/v1/rounds/{fake_id}", json=UPDATED_ROUND, headers=headers)
    assert resp.status_code == 404


@pytest.mark.asyncio
async def test_edit_official_round_forbidden(client, db_session):
    await seed_round_templates(db_session)
    await db_session.commit()
    _, headers = await _register(client, "edit3@test.com", "edituser3")

    resp = await client.get("/api/v1/rounds", headers=headers)
    official = next(r for r in resp.json() if r["is_official"])

    resp = await client.put(f"/api/v1/rounds/{official['id']}", json=UPDATED_ROUND, headers=headers)
    assert resp.status_code == 403


@pytest.mark.asyncio
async def test_edit_not_creator_forbidden(client):
    _, creator_h = await _register(client, "edit4@test.com", "editcreator4")
    _, other_h = await _register(client, "other4@test.com", "editother4")
    custom = await _create_custom_round(client, creator_h)

    resp = await client.put(f"/api/v1/rounds/{custom['id']}", json=UPDATED_ROUND, headers=other_h)
    assert resp.status_code == 403


@pytest.mark.asyncio
async def test_edit_blocked_by_in_progress_session(client):
    _, headers = await _register(client, "edit5@test.com", "edituser5")
    custom = await _create_custom_round(client, headers)

    # Start a scoring session
    resp = await client.post("/api/v1/sessions", json={"template_id": custom["id"]}, headers=headers)
    assert resp.status_code == 201

    # Try to edit while session is in progress
    resp = await client.put(f"/api/v1/rounds/{custom['id']}", json=UPDATED_ROUND, headers=headers)
    assert resp.status_code == 422


@pytest.mark.asyncio
async def test_edit_allowed_after_session_completed(client):
    _, headers = await _register(client, "edit6@test.com", "edituser6")
    custom = await _create_custom_round(client, headers)
    stage_id = custom["stages"][0]["id"]

    # Create and complete a session (submit all 12 ends)
    resp = await client.post("/api/v1/sessions", json={"template_id": custom["id"]}, headers=headers)
    session_id = resp.json()["id"]
    for _ in range(12):
        await _submit_end(client, headers, session_id, stage_id)
    resp = await client.post(f"/api/v1/sessions/{session_id}/complete", headers=headers)
    assert resp.status_code == 200

    # Now edit should work
    resp = await client.put(f"/api/v1/rounds/{custom['id']}", json=UPDATED_ROUND, headers=headers)
    assert resp.status_code == 200
    assert resp.json()["name"] == "Field Course v2"


@pytest.mark.asyncio
async def test_edit_preserves_historical_session_scores(client):
    _, headers = await _register(client, "edit7@test.com", "edituser7")
    custom = await _create_custom_round(client, headers)
    stage_id = custom["stages"][0]["id"]

    # Complete a session
    resp = await client.post("/api/v1/sessions", json={"template_id": custom["id"]}, headers=headers)
    session_id = resp.json()["id"]
    for _ in range(12):
        await _submit_end(client, headers, session_id, stage_id)
    await client.post(f"/api/v1/sessions/{session_id}/complete", headers=headers)

    # Get session score before edit
    resp = await client.get(f"/api/v1/sessions/{session_id}", headers=headers)
    score_before = resp.json()["total_score"]
    arrows_before = resp.json()["total_arrows"]

    # Edit the round template
    resp = await client.put(f"/api/v1/rounds/{custom['id']}", json=UPDATED_ROUND, headers=headers)
    assert resp.status_code == 200

    # Session scores should be unchanged
    resp = await client.get(f"/api/v1/sessions/{session_id}", headers=headers)
    assert resp.json()["total_score"] == score_before
    assert resp.json()["total_arrows"] == arrows_before


@pytest.mark.asyncio
async def test_edit_stages_replaced(client):
    _, headers = await _register(client, "edit8@test.com", "edituser8")
    custom = await _create_custom_round(client, headers)
    old_stage_id = custom["stages"][0]["id"]

    # Edit with two stages
    two_stage_round = {
        "name": "Two Stage Field",
        "organization": "Custom",
        "description": "Two distances",
        "stages": [
            {
                "name": "Close",
                "distance": "10m",
                "num_ends": 6,
                "arrows_per_end": 3,
                "allowed_values": ["X", "10", "9", "8", "7", "6", "5", "4", "3", "2", "1", "M"],
                "value_score_map": {"X": 10, "10": 10, "9": 9, "8": 8, "7": 7, "6": 6, "5": 5, "4": 4, "3": 3, "2": 2, "1": 1, "M": 0},
                "max_score_per_arrow": 10,
            },
            {
                "name": "Far",
                "distance": "30m",
                "num_ends": 6,
                "arrows_per_end": 3,
                "allowed_values": ["X", "10", "9", "8", "7", "6", "5", "4", "3", "2", "1", "M"],
                "value_score_map": {"X": 10, "10": 10, "9": 9, "8": 8, "7": 7, "6": 6, "5": 5, "4": 4, "3": 3, "2": 2, "1": 1, "M": 0},
                "max_score_per_arrow": 10,
            },
        ],
    }
    resp = await client.put(f"/api/v1/rounds/{custom['id']}", json=two_stage_round, headers=headers)
    assert resp.status_code == 200
    data = resp.json()
    assert len(data["stages"]) == 2
    assert data["stages"][0]["name"] == "Close"
    assert data["stages"][1]["name"] == "Far"
    # Old stage ID should no longer be present
    new_stage_ids = [s["id"] for s in data["stages"]]
    assert old_stage_id not in new_stage_ids


@pytest.mark.asyncio
async def test_edit_round_get_reflects_changes(client):
    _, headers = await _register(client, "edit9@test.com", "edituser9")
    custom = await _create_custom_round(client, headers)
    round_id = custom["id"]

    await client.put(f"/api/v1/rounds/{round_id}", json=UPDATED_ROUND, headers=headers)

    # GET should return updated data
    resp = await client.get(f"/api/v1/rounds/{round_id}", headers=headers)
    assert resp.status_code == 200
    data = resp.json()
    assert data["name"] == "Field Course v2"
    assert data["stages"][0]["num_ends"] == 14


@pytest.mark.asyncio
async def test_edit_unauthenticated(client):
    resp = await client.put("/api/v1/rounds/00000000-0000-0000-0000-000000000000", json=UPDATED_ROUND)
    assert resp.status_code == 401
