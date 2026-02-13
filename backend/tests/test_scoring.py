import pytest

from app.seed.round_templates import seed_round_templates


async def _register_and_get_token(client):
    resp = await client.post("/api/v1/auth/register", json={
        "email": "scorer@test.com", "username": "scorer", "password": "pass1234",
    })
    return resp.json()["access_token"]


async def _seed_templates(db_session):
    await seed_round_templates(db_session)


@pytest.mark.asyncio
async def test_list_rounds(client, db_session):
    await _seed_templates(db_session)
    resp = await client.get("/api/v1/rounds")
    assert resp.status_code == 200
    rounds = resp.json()
    assert len(rounds) == 6
    names = [r["name"] for r in rounds]
    assert "NFAA Indoor 300" in names


@pytest.mark.asyncio
async def test_get_round_by_id(client, db_session):
    await _seed_templates(db_session)
    rounds = (await client.get("/api/v1/rounds")).json()
    round_id = rounds[0]["id"]

    resp = await client.get(f"/api/v1/rounds/{round_id}")
    assert resp.status_code == 200
    assert resp.json()["id"] == round_id
    assert "stages" in resp.json()


@pytest.mark.asyncio
async def test_get_round_not_found(client, db_session):
    await _seed_templates(db_session)
    import uuid
    fake_id = str(uuid.uuid4())
    resp = await client.get(f"/api/v1/rounds/{fake_id}")
    assert resp.status_code == 404


@pytest.mark.asyncio
async def test_create_session_and_submit_end(client, db_session):
    await _seed_templates(db_session)
    token = await _register_and_get_token(client)
    headers = {"Authorization": f"Bearer {token}"}

    # Get a round template
    rounds = (await client.get("/api/v1/rounds")).json()
    wa_indoor = next(r for r in rounds if r["name"] == "WA Indoor 18m (Recurve)")
    stage_id = wa_indoor["stages"][0]["id"]

    # Create session
    resp = await client.post("/api/v1/sessions", json={
        "template_id": wa_indoor["id"],
        "location": "Indoor Range",
    }, headers=headers)
    assert resp.status_code == 201
    session = resp.json()
    session_id = session["id"]
    assert session["status"] == "in_progress"

    # Submit an end
    resp = await client.post(f"/api/v1/sessions/{session_id}/ends", json={
        "stage_id": stage_id,
        "arrows": [
            {"score_value": "X"},
            {"score_value": "10"},
            {"score_value": "9"},
        ],
    }, headers=headers)
    assert resp.status_code == 201
    end = resp.json()
    assert end["end_total"] == 29
    assert end["end_number"] == 1
    assert len(end["arrows"]) == 3

    # Check session totals
    resp = await client.get(f"/api/v1/sessions/{session_id}", headers=headers)
    session = resp.json()
    assert session["total_score"] == 29
    assert session["total_x_count"] == 1
    assert session["total_arrows"] == 3


@pytest.mark.asyncio
async def test_invalid_arrow_value(client, db_session):
    await _seed_templates(db_session)
    token = await _register_and_get_token(client)
    headers = {"Authorization": f"Bearer {token}"}

    rounds = (await client.get("/api/v1/rounds")).json()
    nfaa = next(r for r in rounds if r["name"] == "NFAA Indoor 300")
    stage_id = nfaa["stages"][0]["id"]

    resp = await client.post("/api/v1/sessions", json={
        "template_id": nfaa["id"],
    }, headers=headers)
    session_id = resp.json()["id"]

    # Try submitting a "10" to NFAA (not allowed - max is 5)
    resp = await client.post(f"/api/v1/sessions/{session_id}/ends", json={
        "stage_id": stage_id,
        "arrows": [
            {"score_value": "X"},
            {"score_value": "5"},
            {"score_value": "10"},  # invalid for NFAA
            {"score_value": "3"},
            {"score_value": "2"},
        ],
    }, headers=headers)
    assert resp.status_code == 422


@pytest.mark.asyncio
async def test_wrong_arrow_count(client, db_session):
    await _seed_templates(db_session)
    token = await _register_and_get_token(client)
    headers = {"Authorization": f"Bearer {token}"}

    rounds = (await client.get("/api/v1/rounds")).json()
    wa = next(r for r in rounds if r["name"] == "WA Indoor 18m (Recurve)")
    stage_id = wa["stages"][0]["id"]

    resp = await client.post("/api/v1/sessions", json={"template_id": wa["id"]}, headers=headers)
    session_id = resp.json()["id"]

    # Submit 2 arrows when 3 expected
    resp = await client.post(f"/api/v1/sessions/{session_id}/ends", json={
        "stage_id": stage_id,
        "arrows": [{"score_value": "10"}, {"score_value": "9"}],
    }, headers=headers)
    assert resp.status_code == 422


@pytest.mark.asyncio
async def test_complete_session(client, db_session):
    await _seed_templates(db_session)
    token = await _register_and_get_token(client)
    headers = {"Authorization": f"Bearer {token}"}

    rounds = (await client.get("/api/v1/rounds")).json()
    vegas = next(r for r in rounds if r["name"] == "Vegas 300")
    stage_id = vegas["stages"][0]["id"]

    resp = await client.post("/api/v1/sessions", json={"template_id": vegas["id"]}, headers=headers)
    session_id = resp.json()["id"]

    # Submit one end then complete
    await client.post(f"/api/v1/sessions/{session_id}/ends", json={
        "stage_id": stage_id,
        "arrows": [{"score_value": "X"}, {"score_value": "X"}, {"score_value": "X"}],
    }, headers=headers)

    resp = await client.post(f"/api/v1/sessions/{session_id}/complete", headers=headers)
    assert resp.status_code == 200
    assert resp.json()["status"] == "completed"
    assert resp.json()["total_score"] == 30
    assert resp.json()["total_x_count"] == 3


@pytest.mark.asyncio
async def test_session_with_setup_profile(client, db_session):
    await _seed_templates(db_session)
    token = await _register_and_get_token(client)
    headers = {"Authorization": f"Bearer {token}"}

    # Create a setup profile
    resp = await client.post("/api/v1/setups", json={
        "name": "Indoor Setup", "draw_weight": 38.0,
    }, headers=headers)
    assert resp.status_code == 201
    setup_id = resp.json()["id"]

    # Create session with setup
    rounds = (await client.get("/api/v1/rounds")).json()
    wa = next(r for r in rounds if r["name"] == "WA Indoor 18m (Recurve)")

    resp = await client.post("/api/v1/sessions", json={
        "template_id": wa["id"],
        "setup_profile_id": setup_id,
    }, headers=headers)
    assert resp.status_code == 201
    session = resp.json()
    assert session["setup_profile_id"] == setup_id
    assert session["setup_profile_name"] == "Indoor Setup"

    # Verify it shows in session detail
    resp = await client.get(f"/api/v1/sessions/{session['id']}", headers=headers)
    assert resp.json()["setup_profile_name"] == "Indoor Setup"

    # Session without setup still works
    resp = await client.post("/api/v1/sessions", json={
        "template_id": wa["id"],
    }, headers=headers)
    assert resp.status_code == 201
    assert resp.json()["setup_profile_id"] is None
    assert resp.json()["setup_profile_name"] is None


@pytest.mark.asyncio
async def test_session_stats(client, db_session):
    await _seed_templates(db_session)
    token = await _register_and_get_token(client)
    headers = {"Authorization": f"Bearer {token}"}

    # Stats with no sessions
    resp = await client.get("/api/v1/sessions/stats", headers=headers)
    assert resp.status_code == 200
    stats = resp.json()
    assert stats["total_sessions"] == 0
    assert stats["personal_best_score"] is None

    # Create and complete a session
    rounds = (await client.get("/api/v1/rounds")).json()
    vegas = next(r for r in rounds if r["name"] == "Vegas 300")
    stage_id = vegas["stages"][0]["id"]

    resp = await client.post("/api/v1/sessions", json={"template_id": vegas["id"]}, headers=headers)
    session_id = resp.json()["id"]

    await client.post(f"/api/v1/sessions/{session_id}/ends", json={
        "stage_id": stage_id,
        "arrows": [{"score_value": "X"}, {"score_value": "10"}, {"score_value": "9"}],
    }, headers=headers)

    await client.post(f"/api/v1/sessions/{session_id}/complete", headers=headers)

    # Stats should reflect the completed session
    resp = await client.get("/api/v1/sessions/stats", headers=headers)
    assert resp.status_code == 200
    stats = resp.json()
    assert stats["total_sessions"] == 1
    assert stats["completed_sessions"] == 1
    assert stats["total_arrows"] == 3
    assert stats["total_x_count"] == 1
    assert stats["personal_best_score"] == 29
    assert stats["personal_best_template"] == "Vegas 300"
    assert len(stats["avg_by_round_type"]) == 1
    assert stats["avg_by_round_type"][0]["template_name"] == "Vegas 300"
    assert len(stats["recent_trend"]) == 1


@pytest.mark.asyncio
async def test_multi_stage_session(client, db_session):
    await _seed_templates(db_session)
    token = await _register_and_get_token(client)
    headers = {"Authorization": f"Bearer {token}"}

    rounds = (await client.get("/api/v1/rounds")).json()
    wa1440 = next(r for r in rounds if r["name"] == "WA 1440 (Recurve)")
    assert len(wa1440["stages"]) == 4

    stage1 = wa1440["stages"][0]  # 90m
    stage2 = wa1440["stages"][1]  # 70m
    assert stage1["distance"] == "90m"
    assert stage2["distance"] == "70m"

    # Create session
    resp = await client.post("/api/v1/sessions", json={
        "template_id": wa1440["id"],
    }, headers=headers)
    assert resp.status_code == 201
    session_id = resp.json()["id"]

    # Submit an end to stage 1 (90m): 6 arrows
    resp = await client.post(f"/api/v1/sessions/{session_id}/ends", json={
        "stage_id": stage1["id"],
        "arrows": [
            {"score_value": "X"}, {"score_value": "10"}, {"score_value": "9"},
            {"score_value": "8"}, {"score_value": "7"}, {"score_value": "6"},
        ],
    }, headers=headers)
    assert resp.status_code == 201
    assert resp.json()["end_total"] == 50

    # Submit an end to stage 2 (70m): 6 arrows
    resp = await client.post(f"/api/v1/sessions/{session_id}/ends", json={
        "stage_id": stage2["id"],
        "arrows": [
            {"score_value": "X"}, {"score_value": "X"}, {"score_value": "10"},
            {"score_value": "10"}, {"score_value": "9"}, {"score_value": "9"},
        ],
    }, headers=headers)
    assert resp.status_code == 201
    assert resp.json()["end_total"] == 58

    # Verify session totals aggregate across stages
    resp = await client.get(f"/api/v1/sessions/{session_id}", headers=headers)
    session = resp.json()
    assert session["total_score"] == 108
    assert session["total_arrows"] == 12
    assert session["total_x_count"] == 3
