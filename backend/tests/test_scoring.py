import uuid

import pytest

from app.seed.round_templates import seed_round_templates

_user_counter = 0


async def _register_and_get_token(client, email=None, username=None):
    global _user_counter
    _user_counter += 1
    email = email or f"scorer{_user_counter}@test.com"
    username = username or f"scorer{_user_counter}"
    resp = await client.post("/api/v1/auth/register", json={
        "email": email, "username": username, "password": "pass1234",
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
async def test_complete_session_with_location_weather(client, db_session):
    await _seed_templates(db_session)
    token = await _register_and_get_token(client)
    headers = {"Authorization": f"Bearer {token}"}

    rounds = (await client.get("/api/v1/rounds")).json()
    vegas = next(r for r in rounds if r["name"] == "Vegas 300")
    stage_id = vegas["stages"][0]["id"]

    # Create session without location/weather
    resp = await client.post("/api/v1/sessions", json={"template_id": vegas["id"]}, headers=headers)
    session_id = resp.json()["id"]
    assert resp.json()["location"] is None
    assert resp.json()["weather"] is None

    # Submit one end
    await client.post(f"/api/v1/sessions/{session_id}/ends", json={
        "stage_id": stage_id,
        "arrows": [{"score_value": "X"}, {"score_value": "X"}, {"score_value": "X"}],
    }, headers=headers)

    # Complete with notes, location, and weather
    resp = await client.post(f"/api/v1/sessions/{session_id}/complete", json={
        "notes": "Great session",
        "location": "Indoor Range",
        "weather": "Sunny, 72F",
    }, headers=headers)
    assert resp.status_code == 200
    data = resp.json()
    assert data["status"] == "completed"
    assert data["notes"] == "Great session"
    assert data["location"] == "Indoor Range"
    assert data["weather"] == "Sunny, 72F"


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


@pytest.mark.asyncio
async def test_create_session_invalid_setup_profile(client, db_session):
    await _seed_templates(db_session)
    token = await _register_and_get_token(client)
    headers = {"Authorization": f"Bearer {token}"}

    rounds = (await client.get("/api/v1/rounds")).json()
    wa = next(r for r in rounds if r["name"] == "WA Indoor 18m (Recurve)")

    fake_id = str(uuid.uuid4())
    resp = await client.post("/api/v1/sessions", json={
        "template_id": wa["id"],
        "setup_profile_id": fake_id,
    }, headers=headers)
    assert resp.status_code == 404


@pytest.mark.asyncio
async def test_list_sessions_with_template_name(client, db_session):
    await _seed_templates(db_session)
    token = await _register_and_get_token(client)
    headers = {"Authorization": f"Bearer {token}"}

    rounds = (await client.get("/api/v1/rounds")).json()
    vegas = next(r for r in rounds if r["name"] == "Vegas 300")

    await client.post("/api/v1/sessions", json={"template_id": vegas["id"]}, headers=headers)

    resp = await client.get("/api/v1/sessions", headers=headers)
    assert resp.status_code == 200
    sessions = resp.json()
    assert len(sessions) >= 1
    assert sessions[0]["template_name"] == "Vegas 300"


@pytest.mark.asyncio
async def test_list_sessions_with_setup_profile_name(client, db_session):
    await _seed_templates(db_session)
    token = await _register_and_get_token(client)
    headers = {"Authorization": f"Bearer {token}"}

    # Create setup profile
    resp = await client.post("/api/v1/setups", json={"name": "My Setup"}, headers=headers)
    setup_id = resp.json()["id"]

    rounds = (await client.get("/api/v1/rounds")).json()
    vegas = next(r for r in rounds if r["name"] == "Vegas 300")

    await client.post("/api/v1/sessions", json={
        "template_id": vegas["id"],
        "setup_profile_id": setup_id,
    }, headers=headers)

    resp = await client.get("/api/v1/sessions", headers=headers)
    assert resp.status_code == 200
    sessions = resp.json()
    assert sessions[0]["setup_profile_name"] == "My Setup"


@pytest.mark.asyncio
async def test_personal_records_empty(client, db_session):
    await _seed_templates(db_session)
    token = await _register_and_get_token(client)
    headers = {"Authorization": f"Bearer {token}"}

    resp = await client.get("/api/v1/sessions/personal-records", headers=headers)
    assert resp.status_code == 200
    assert resp.json() == []


@pytest.mark.asyncio
async def test_personal_records_with_pr(client, db_session):
    await _seed_templates(db_session)
    token = await _register_and_get_token(client)
    headers = {"Authorization": f"Bearer {token}"}

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

    resp = await client.get("/api/v1/sessions/personal-records", headers=headers)
    assert resp.status_code == 200
    prs = resp.json()
    assert len(prs) == 1
    assert prs[0]["template_name"] == "Vegas 300"
    assert prs[0]["score"] == 29
    assert prs[0]["session_id"] == session_id


@pytest.mark.asyncio
async def test_get_session_not_found(client, db_session):
    await _seed_templates(db_session)
    token = await _register_and_get_token(client)
    headers = {"Authorization": f"Bearer {token}"}

    fake_id = str(uuid.uuid4())
    resp = await client.get(f"/api/v1/sessions/{fake_id}", headers=headers)
    assert resp.status_code == 404


@pytest.mark.asyncio
async def test_get_session_is_personal_best(client, db_session):
    await _seed_templates(db_session)
    token = await _register_and_get_token(client)
    headers = {"Authorization": f"Bearer {token}"}

    rounds = (await client.get("/api/v1/rounds")).json()
    vegas = next(r for r in rounds if r["name"] == "Vegas 300")
    stage_id = vegas["stages"][0]["id"]

    resp = await client.post("/api/v1/sessions", json={"template_id": vegas["id"]}, headers=headers)
    session_id = resp.json()["id"]

    await client.post(f"/api/v1/sessions/{session_id}/ends", json={
        "stage_id": stage_id,
        "arrows": [{"score_value": "X"}, {"score_value": "X"}, {"score_value": "X"}],
    }, headers=headers)
    await client.post(f"/api/v1/sessions/{session_id}/complete", headers=headers)

    resp = await client.get(f"/api/v1/sessions/{session_id}", headers=headers)
    assert resp.status_code == 200
    assert resp.json()["is_personal_best"] is True


@pytest.mark.asyncio
async def test_get_session_not_personal_best(client, db_session):
    await _seed_templates(db_session)
    token = await _register_and_get_token(client)
    headers = {"Authorization": f"Bearer {token}"}

    rounds = (await client.get("/api/v1/rounds")).json()
    vegas = next(r for r in rounds if r["name"] == "Vegas 300")
    stage_id = vegas["stages"][0]["id"]

    # First session — high score (this becomes PR)
    resp = await client.post("/api/v1/sessions", json={"template_id": vegas["id"]}, headers=headers)
    s1_id = resp.json()["id"]
    await client.post(f"/api/v1/sessions/{s1_id}/ends", json={
        "stage_id": stage_id,
        "arrows": [{"score_value": "X"}, {"score_value": "X"}, {"score_value": "X"}],
    }, headers=headers)
    await client.post(f"/api/v1/sessions/{s1_id}/complete", headers=headers)

    # Second session — lower score
    resp = await client.post("/api/v1/sessions", json={"template_id": vegas["id"]}, headers=headers)
    s2_id = resp.json()["id"]
    await client.post(f"/api/v1/sessions/{s2_id}/ends", json={
        "stage_id": stage_id,
        "arrows": [{"score_value": "9"}, {"score_value": "8"}, {"score_value": "7"}],
    }, headers=headers)
    await client.post(f"/api/v1/sessions/{s2_id}/complete", headers=headers)

    resp = await client.get(f"/api/v1/sessions/{s2_id}", headers=headers)
    assert resp.status_code == 200
    assert resp.json()["is_personal_best"] is False


@pytest.mark.asyncio
async def test_undo_last_end(client, db_session):
    await _seed_templates(db_session)
    token = await _register_and_get_token(client)
    headers = {"Authorization": f"Bearer {token}"}

    rounds = (await client.get("/api/v1/rounds")).json()
    vegas = next(r for r in rounds if r["name"] == "Vegas 300")
    stage_id = vegas["stages"][0]["id"]

    resp = await client.post("/api/v1/sessions", json={"template_id": vegas["id"]}, headers=headers)
    session_id = resp.json()["id"]

    # Submit two ends
    await client.post(f"/api/v1/sessions/{session_id}/ends", json={
        "stage_id": stage_id,
        "arrows": [{"score_value": "X"}, {"score_value": "10"}, {"score_value": "9"}],
    }, headers=headers)
    await client.post(f"/api/v1/sessions/{session_id}/ends", json={
        "stage_id": stage_id,
        "arrows": [{"score_value": "8"}, {"score_value": "7"}, {"score_value": "6"}],
    }, headers=headers)

    # Undo last end
    resp = await client.delete(f"/api/v1/sessions/{session_id}/ends/last", headers=headers)
    assert resp.status_code == 200
    session = resp.json()
    assert session["total_score"] == 29  # only first end remains
    assert session["total_arrows"] == 3
    assert session["total_x_count"] == 1
    assert len(session["ends"]) == 1


@pytest.mark.asyncio
async def test_undo_last_end_empty(client, db_session):
    await _seed_templates(db_session)
    token = await _register_and_get_token(client)
    headers = {"Authorization": f"Bearer {token}"}

    rounds = (await client.get("/api/v1/rounds")).json()
    vegas = next(r for r in rounds if r["name"] == "Vegas 300")

    resp = await client.post("/api/v1/sessions", json={"template_id": vegas["id"]}, headers=headers)
    session_id = resp.json()["id"]

    # Undo with no ends
    resp = await client.delete(f"/api/v1/sessions/{session_id}/ends/last", headers=headers)
    assert resp.status_code == 422


@pytest.mark.asyncio
async def test_undo_last_end_completed(client, db_session):
    await _seed_templates(db_session)
    token = await _register_and_get_token(client)
    headers = {"Authorization": f"Bearer {token}"}

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

    # Undo on completed session
    resp = await client.delete(f"/api/v1/sessions/{session_id}/ends/last", headers=headers)
    assert resp.status_code == 422


@pytest.mark.asyncio
async def test_submit_end_nonexistent_session(client, db_session):
    await _seed_templates(db_session)
    token = await _register_and_get_token(client)
    headers = {"Authorization": f"Bearer {token}"}

    rounds = (await client.get("/api/v1/rounds")).json()
    vegas = next(r for r in rounds if r["name"] == "Vegas 300")
    stage_id = vegas["stages"][0]["id"]

    fake_id = str(uuid.uuid4())
    resp = await client.post(f"/api/v1/sessions/{fake_id}/ends", json={
        "stage_id": stage_id,
        "arrows": [{"score_value": "X"}, {"score_value": "10"}, {"score_value": "9"}],
    }, headers=headers)
    assert resp.status_code == 404


@pytest.mark.asyncio
async def test_submit_end_completed_session(client, db_session):
    await _seed_templates(db_session)
    token = await _register_and_get_token(client)
    headers = {"Authorization": f"Bearer {token}"}

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

    # Try submitting end to completed session
    resp = await client.post(f"/api/v1/sessions/{session_id}/ends", json={
        "stage_id": stage_id,
        "arrows": [{"score_value": "X"}, {"score_value": "10"}, {"score_value": "9"}],
    }, headers=headers)
    assert resp.status_code == 422


@pytest.mark.asyncio
async def test_submit_end_invalid_stage(client, db_session):
    await _seed_templates(db_session)
    token = await _register_and_get_token(client)
    headers = {"Authorization": f"Bearer {token}"}

    rounds = (await client.get("/api/v1/rounds")).json()
    vegas = next(r for r in rounds if r["name"] == "Vegas 300")

    resp = await client.post("/api/v1/sessions", json={"template_id": vegas["id"]}, headers=headers)
    session_id = resp.json()["id"]

    fake_stage_id = str(uuid.uuid4())
    resp = await client.post(f"/api/v1/sessions/{session_id}/ends", json={
        "stage_id": fake_stage_id,
        "arrows": [{"score_value": "X"}, {"score_value": "10"}, {"score_value": "9"}],
    }, headers=headers)
    assert resp.status_code == 404


@pytest.mark.asyncio
async def test_complete_session_not_found(client, db_session):
    await _seed_templates(db_session)
    token = await _register_and_get_token(client)
    headers = {"Authorization": f"Bearer {token}"}

    fake_id = str(uuid.uuid4())
    resp = await client.post(f"/api/v1/sessions/{fake_id}/complete", headers=headers)
    assert resp.status_code == 404


@pytest.mark.asyncio
async def test_complete_session_updates_existing_pr(client, db_session):
    await _seed_templates(db_session)
    token = await _register_and_get_token(client)
    headers = {"Authorization": f"Bearer {token}"}

    rounds = (await client.get("/api/v1/rounds")).json()
    vegas = next(r for r in rounds if r["name"] == "Vegas 300")
    stage_id = vegas["stages"][0]["id"]

    # First session — lower score
    resp = await client.post("/api/v1/sessions", json={"template_id": vegas["id"]}, headers=headers)
    s1_id = resp.json()["id"]
    await client.post(f"/api/v1/sessions/{s1_id}/ends", json={
        "stage_id": stage_id,
        "arrows": [{"score_value": "9"}, {"score_value": "8"}, {"score_value": "7"}],
    }, headers=headers)
    resp = await client.post(f"/api/v1/sessions/{s1_id}/complete", headers=headers)
    assert resp.json()["is_personal_best"] is True

    # Second session — higher score should update PR
    resp = await client.post("/api/v1/sessions", json={"template_id": vegas["id"]}, headers=headers)
    s2_id = resp.json()["id"]
    await client.post(f"/api/v1/sessions/{s2_id}/ends", json={
        "stage_id": stage_id,
        "arrows": [{"score_value": "X"}, {"score_value": "X"}, {"score_value": "X"}],
    }, headers=headers)
    resp = await client.post(f"/api/v1/sessions/{s2_id}/complete", headers=headers)
    assert resp.json()["is_personal_best"] is True

    # Check PR updated to the higher score
    resp = await client.get("/api/v1/sessions/personal-records", headers=headers)
    prs = resp.json()
    assert len(prs) == 1
    assert prs[0]["score"] == 30
    assert prs[0]["session_id"] == s2_id


@pytest.mark.asyncio
async def test_list_sessions_filter_by_template(client, db_session):
    await _seed_templates(db_session)
    token = await _register_and_get_token(client)
    headers = {"Authorization": f"Bearer {token}"}

    rounds = (await client.get("/api/v1/rounds")).json()
    vegas = next(r for r in rounds if r["name"] == "Vegas 300")
    wa = next(r for r in rounds if r["name"] == "WA Indoor 18m (Recurve)")

    await client.post("/api/v1/sessions", json={"template_id": vegas["id"]}, headers=headers)
    await client.post("/api/v1/sessions", json={"template_id": wa["id"]}, headers=headers)

    # Filter by Vegas template
    resp = await client.get(f"/api/v1/sessions?template_id={vegas['id']}", headers=headers)
    assert resp.status_code == 200
    sessions = resp.json()
    assert len(sessions) == 1
    assert sessions[0]["template_name"] == "Vegas 300"


@pytest.mark.asyncio
async def test_list_sessions_filter_by_date_range(client, db_session):
    await _seed_templates(db_session)
    token = await _register_and_get_token(client)
    headers = {"Authorization": f"Bearer {token}"}

    rounds = (await client.get("/api/v1/rounds")).json()
    vegas = next(r for r in rounds if r["name"] == "Vegas 300")
    await client.post("/api/v1/sessions", json={"template_id": vegas["id"]}, headers=headers)

    # Filter with today's date range — should include it
    from datetime import date
    today = date.today().isoformat()
    resp = await client.get(f"/api/v1/sessions?date_from={today}&date_to={today}", headers=headers)
    assert resp.status_code == 200
    assert len(resp.json()) >= 1

    # Filter with a past date range — should exclude it
    resp = await client.get("/api/v1/sessions?date_from=2020-01-01&date_to=2020-12-31", headers=headers)
    assert resp.status_code == 200
    assert len(resp.json()) == 0


@pytest.mark.asyncio
async def test_list_sessions_search_notes(client, db_session):
    await _seed_templates(db_session)
    token = await _register_and_get_token(client)
    headers = {"Authorization": f"Bearer {token}"}

    rounds = (await client.get("/api/v1/rounds")).json()
    vegas = next(r for r in rounds if r["name"] == "Vegas 300")
    await client.post("/api/v1/sessions", json={
        "template_id": vegas["id"],
        "notes": "Windy afternoon practice",
    }, headers=headers)
    await client.post("/api/v1/sessions", json={
        "template_id": vegas["id"],
        "notes": "Calm morning session",
    }, headers=headers)

    resp = await client.get("/api/v1/sessions?search=windy", headers=headers)
    assert resp.status_code == 200
    assert len(resp.json()) == 1


@pytest.mark.asyncio
async def test_list_sessions_search_location(client, db_session):
    await _seed_templates(db_session)
    token = await _register_and_get_token(client)
    headers = {"Authorization": f"Bearer {token}"}

    rounds = (await client.get("/api/v1/rounds")).json()
    vegas = next(r for r in rounds if r["name"] == "Vegas 300")
    await client.post("/api/v1/sessions", json={
        "template_id": vegas["id"],
        "location": "City Archery Club",
    }, headers=headers)

    resp = await client.get("/api/v1/sessions?search=archery", headers=headers)
    assert resp.status_code == 200
    assert len(resp.json()) == 1


@pytest.mark.asyncio
async def test_export_sessions_csv(client, db_session):
    await _seed_templates(db_session)
    token = await _register_and_get_token(client)
    headers = {"Authorization": f"Bearer {token}"}

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

    resp = await client.get("/api/v1/sessions/export", headers=headers)
    assert resp.status_code == 200
    assert "text/csv" in resp.headers["content-type"]
    assert "Vegas 300" in resp.text


@pytest.mark.asyncio
async def test_export_single_session_csv(client, db_session):
    await _seed_templates(db_session)
    token = await _register_and_get_token(client)
    headers = {"Authorization": f"Bearer {token}"}

    rounds = (await client.get("/api/v1/rounds")).json()
    vegas = next(r for r in rounds if r["name"] == "Vegas 300")
    stage_id = vegas["stages"][0]["id"]

    resp = await client.post("/api/v1/sessions", json={"template_id": vegas["id"]}, headers=headers)
    session_id = resp.json()["id"]
    await client.post(f"/api/v1/sessions/{session_id}/ends", json={
        "stage_id": stage_id,
        "arrows": [{"score_value": "X"}, {"score_value": "10"}, {"score_value": "9"}],
    }, headers=headers)

    resp = await client.get(f"/api/v1/sessions/{session_id}/export?format=csv", headers=headers)
    assert resp.status_code == 200
    assert "text/csv" in resp.headers["content-type"]
    assert "X" in resp.text


@pytest.mark.asyncio
async def test_export_single_session_pdf(client, db_session):
    await _seed_templates(db_session)
    token = await _register_and_get_token(client)
    headers = {"Authorization": f"Bearer {token}"}

    rounds = (await client.get("/api/v1/rounds")).json()
    vegas = next(r for r in rounds if r["name"] == "Vegas 300")
    stage_id = vegas["stages"][0]["id"]

    resp = await client.post("/api/v1/sessions", json={"template_id": vegas["id"]}, headers=headers)
    session_id = resp.json()["id"]
    await client.post(f"/api/v1/sessions/{session_id}/ends", json={
        "stage_id": stage_id,
        "arrows": [{"score_value": "X"}, {"score_value": "10"}, {"score_value": "9"}],
    }, headers=headers)

    resp = await client.get(f"/api/v1/sessions/{session_id}/export?format=pdf", headers=headers)
    assert resp.status_code == 200
    assert "application/pdf" in resp.headers["content-type"]
    assert resp.content[:4] == b"%PDF"


@pytest.mark.asyncio
async def test_trends_empty(client, db_session):
    """GET /sessions/trends returns empty list when no completed sessions."""
    token = await _register_and_get_token(client)
    headers = {"Authorization": f"Bearer {token}"}
    resp = await client.get("/api/v1/sessions/trends", headers=headers)
    assert resp.status_code == 200
    assert resp.json() == []


@pytest.mark.asyncio
async def test_trends_with_data(client, db_session):
    """GET /sessions/trends returns trend data for completed sessions."""
    token = await _register_and_get_token(client)
    headers = {"Authorization": f"Bearer {token}"}
    await seed_round_templates(db_session)
    await db_session.commit()

    rounds = await client.get("/api/v1/rounds", headers=headers)
    template = rounds.json()[0]
    stage = template["stages"][0]

    # Create and complete a session
    session_resp = await client.post("/api/v1/sessions", json={"template_id": template["id"]}, headers=headers)
    session_id = session_resp.json()["id"]

    arrows = [{"score_value": stage["allowed_values"][0]}] * stage["arrows_per_end"]
    await client.post(f"/api/v1/sessions/{session_id}/ends", json={
        "stage_id": stage["id"],
        "arrows": arrows,
    }, headers=headers)
    await client.post(f"/api/v1/sessions/{session_id}/complete", headers=headers)

    resp = await client.get("/api/v1/sessions/trends", headers=headers)
    assert resp.status_code == 200
    data = resp.json()
    assert len(data) == 1
    item = data[0]
    assert item["session_id"] == session_id
    assert item["template_name"] == template["name"]
    assert item["total_score"] > 0
    assert item["max_score"] > 0
    assert item["percentage"] > 0
    assert "completed_at" in item


# --- Custom Round Templates ---

CUSTOM_ROUND = {
    "name": "My Practice Round",
    "organization": "Custom",
    "description": "A short practice round",
    "stages": [
        {
            "name": "Stage 1",
            "distance": "10m",
            "num_ends": 3,
            "arrows_per_end": 3,
            "allowed_values": ["X", "10", "9", "8", "7", "6", "5", "4", "3", "2", "1", "M"],
            "value_score_map": {"X": 10, "10": 10, "9": 9, "8": 8, "7": 7, "6": 6, "5": 5, "4": 4, "3": 3, "2": 2, "1": 1, "M": 0},
            "max_score_per_arrow": 10,
        }
    ],
}


@pytest.mark.asyncio
async def test_create_custom_round(client, db_session):
    """POST /rounds creates a custom round template."""
    token = await _register_and_get_token(client)
    headers = {"Authorization": f"Bearer {token}"}

    resp = await client.post("/api/v1/rounds", json=CUSTOM_ROUND, headers=headers)
    assert resp.status_code == 201
    data = resp.json()
    assert data["name"] == "My Practice Round"
    assert data["is_official"] is False
    assert len(data["stages"]) == 1
    assert data["stages"][0]["num_ends"] == 3


@pytest.mark.asyncio
async def test_delete_custom_round(client, db_session):
    """DELETE /rounds/{id} deletes user's own custom round."""
    token = await _register_and_get_token(client)
    headers = {"Authorization": f"Bearer {token}"}

    resp = await client.post("/api/v1/rounds", json=CUSTOM_ROUND, headers=headers)
    round_id = resp.json()["id"]

    resp = await client.delete(f"/api/v1/rounds/{round_id}", headers=headers)
    assert resp.status_code == 204

    resp = await client.get(f"/api/v1/rounds/{round_id}", headers=headers)
    assert resp.status_code == 404


@pytest.mark.asyncio
async def test_cannot_delete_official_round(client, db_session):
    """DELETE /rounds/{id} returns 403 for official round templates."""
    token = await _register_and_get_token(client)
    headers = {"Authorization": f"Bearer {token}"}
    await seed_round_templates(db_session)
    await db_session.commit()

    rounds = await client.get("/api/v1/rounds", headers=headers)
    official_id = rounds.json()[0]["id"]

    resp = await client.delete(f"/api/v1/rounds/{official_id}", headers=headers)
    assert resp.status_code == 403


@pytest.mark.asyncio
async def test_list_rounds_includes_custom(client, db_session):
    """GET /rounds includes user's custom rounds alongside official ones."""
    token = await _register_and_get_token(client)
    headers = {"Authorization": f"Bearer {token}"}
    await seed_round_templates(db_session)
    await db_session.commit()

    # Create a custom round
    await client.post("/api/v1/rounds", json=CUSTOM_ROUND, headers=headers)

    resp = await client.get("/api/v1/rounds", headers=headers)
    names = [r["name"] for r in resp.json()]
    assert "My Practice Round" in names
    # Official rounds should also be there
    assert any(r["is_official"] for r in resp.json())


@pytest.mark.asyncio
async def test_delete_abandoned_session(client, db_session):
    """DELETE /sessions/{id} deletes an abandoned session and its data."""
    await seed_round_templates(db_session)
    await db_session.commit()
    token = await _register_and_get_token(client)
    headers = {"Authorization": f"Bearer {token}"}

    rounds = (await client.get("/api/v1/rounds", headers=headers)).json()
    vegas = next(r for r in rounds if r["name"] == "Vegas 300")
    stage_id = vegas["stages"][0]["id"]

    # Create session and submit an end
    resp = await client.post("/api/v1/sessions", json={"template_id": vegas["id"]}, headers=headers)
    session_id = resp.json()["id"]
    await client.post(f"/api/v1/sessions/{session_id}/ends", json={
        "stage_id": stage_id,
        "arrows": [{"score_value": "X"}, {"score_value": "10"}, {"score_value": "9"}],
    }, headers=headers)

    # Abandon then delete
    await client.post(f"/api/v1/sessions/{session_id}/abandon", headers=headers)
    resp = await client.delete(f"/api/v1/sessions/{session_id}", headers=headers)
    assert resp.status_code == 204

    # Session should be gone
    resp = await client.get(f"/api/v1/sessions/{session_id}", headers=headers)
    assert resp.status_code == 404


@pytest.mark.asyncio
async def test_delete_completed_session_forbidden(client, db_session):
    """DELETE /sessions/{id} returns 422 for completed sessions."""
    await seed_round_templates(db_session)
    await db_session.commit()
    token = await _register_and_get_token(client)
    headers = {"Authorization": f"Bearer {token}"}

    rounds = (await client.get("/api/v1/rounds", headers=headers)).json()
    vegas = next(r for r in rounds if r["name"] == "Vegas 300")
    stage_id = vegas["stages"][0]["id"]

    resp = await client.post("/api/v1/sessions", json={"template_id": vegas["id"]}, headers=headers)
    session_id = resp.json()["id"]
    await client.post(f"/api/v1/sessions/{session_id}/ends", json={
        "stage_id": stage_id,
        "arrows": [{"score_value": "X"}, {"score_value": "10"}, {"score_value": "9"}],
    }, headers=headers)
    await client.post(f"/api/v1/sessions/{session_id}/complete", headers=headers)

    resp = await client.delete(f"/api/v1/sessions/{session_id}", headers=headers)
    assert resp.status_code == 422


@pytest.mark.asyncio
async def test_delete_in_progress_session_forbidden(client, db_session):
    """DELETE /sessions/{id} returns 422 for in-progress sessions."""
    await seed_round_templates(db_session)
    await db_session.commit()
    token = await _register_and_get_token(client)
    headers = {"Authorization": f"Bearer {token}"}

    rounds = (await client.get("/api/v1/rounds", headers=headers)).json()
    vegas = next(r for r in rounds if r["name"] == "Vegas 300")

    resp = await client.post("/api/v1/sessions", json={"template_id": vegas["id"]}, headers=headers)
    session_id = resp.json()["id"]

    resp = await client.delete(f"/api/v1/sessions/{session_id}", headers=headers)
    assert resp.status_code == 422


@pytest.mark.asyncio
async def test_delete_other_users_session_forbidden(client, db_session):
    """DELETE /sessions/{id} returns 404 for another user's session."""
    await seed_round_templates(db_session)
    await db_session.commit()
    token1 = await _register_and_get_token(client, "del1@test.com", "deluser1")
    token2 = await _register_and_get_token(client, "del2@test.com", "deluser2")
    headers1 = {"Authorization": f"Bearer {token1}"}
    headers2 = {"Authorization": f"Bearer {token2}"}

    rounds = (await client.get("/api/v1/rounds", headers=headers1)).json()
    vegas = next(r for r in rounds if r["name"] == "Vegas 300")

    resp = await client.post("/api/v1/sessions", json={"template_id": vegas["id"]}, headers=headers1)
    session_id = resp.json()["id"]
    await client.post(f"/api/v1/sessions/{session_id}/abandon", headers=headers1)

    # Other user tries to delete
    resp = await client.delete(f"/api/v1/sessions/{session_id}", headers=headers2)
    assert resp.status_code == 404
