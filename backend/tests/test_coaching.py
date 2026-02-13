import pytest

from app.seed.round_templates import seed_round_templates

_user_counter = 0


async def _register_and_get_token(client, email=None, username=None):
    global _user_counter
    _user_counter += 1
    email = email or f"coach{_user_counter}@test.com"
    username = username or f"coach{_user_counter}"
    resp = await client.post("/api/v1/auth/register", json={
        "email": email, "username": username, "password": "pass1234",
    })
    return resp.json()["access_token"], username


@pytest.mark.asyncio
async def test_invite_athlete(client, db_session):
    """POST /coaching/invite creates a pending link."""
    coach_token, _ = await _register_and_get_token(client)
    _, athlete_username = await _register_and_get_token(client)

    resp = await client.post("/api/v1/coaching/invite", json={
        "athlete_username": athlete_username,
    }, headers={"Authorization": f"Bearer {coach_token}"})
    assert resp.status_code == 201
    assert resp.json()["status"] == "pending"
    assert resp.json()["athlete_username"] == athlete_username


@pytest.mark.asyncio
async def test_accept_invite(client, db_session):
    """POST /coaching/respond accepts invite."""
    coach_token, _ = await _register_and_get_token(client)
    athlete_token, athlete_username = await _register_and_get_token(client)

    invite = await client.post("/api/v1/coaching/invite", json={
        "athlete_username": athlete_username,
    }, headers={"Authorization": f"Bearer {coach_token}"})
    link_id = invite.json()["id"]

    resp = await client.post("/api/v1/coaching/respond", json={
        "link_id": link_id, "accept": True,
    }, headers={"Authorization": f"Bearer {athlete_token}"})
    assert resp.status_code == 200
    assert resp.json()["status"] == "active"


@pytest.mark.asyncio
async def test_reject_invite(client, db_session):
    """POST /coaching/respond rejects invite."""
    coach_token, _ = await _register_and_get_token(client)
    athlete_token, athlete_username = await _register_and_get_token(client)

    invite = await client.post("/api/v1/coaching/invite", json={
        "athlete_username": athlete_username,
    }, headers={"Authorization": f"Bearer {coach_token}"})
    link_id = invite.json()["id"]

    resp = await client.post("/api/v1/coaching/respond", json={
        "link_id": link_id, "accept": False,
    }, headers={"Authorization": f"Bearer {athlete_token}"})
    assert resp.status_code == 200
    assert resp.json()["status"] == "revoked"


@pytest.mark.asyncio
async def test_view_athlete_sessions(client, db_session):
    """Coach can view athlete's sessions after accepting."""
    await seed_round_templates(db_session)
    coach_token, _ = await _register_and_get_token(client)
    athlete_token, athlete_username = await _register_and_get_token(client)

    # Create coaching link
    invite = await client.post("/api/v1/coaching/invite", json={
        "athlete_username": athlete_username,
    }, headers={"Authorization": f"Bearer {coach_token}"})
    link_id = invite.json()["id"]
    await client.post("/api/v1/coaching/respond", json={
        "link_id": link_id, "accept": True,
    }, headers={"Authorization": f"Bearer {athlete_token}"})

    # Athlete creates and completes a session
    rounds_resp = await client.get("/api/v1/rounds", headers={"Authorization": f"Bearer {athlete_token}"})
    template = next(r for r in rounds_resp.json() if r["name"] == "WA Indoor 18m (Recurve)")
    template_id = template["id"]
    stage_id = template["stages"][0]["id"]

    s_resp = await client.post("/api/v1/sessions", json={"template_id": template_id},
                               headers={"Authorization": f"Bearer {athlete_token}"})
    session_id = s_resp.json()["id"]
    await client.post(f"/api/v1/sessions/{session_id}/ends", json={
        "stage_id": stage_id,
        "arrows": [{"score_value": "10"}, {"score_value": "9"}, {"score_value": "8"}],
    }, headers={"Authorization": f"Bearer {athlete_token}"})
    await client.post(f"/api/v1/sessions/{session_id}/complete",
                      headers={"Authorization": f"Bearer {athlete_token}"})

    # Coach views athlete sessions
    athlete_id = invite.json()["athlete_id"]
    resp = await client.get(f"/api/v1/coaching/athletes/{athlete_id}/sessions",
                            headers={"Authorization": f"Bearer {coach_token}"})
    assert resp.status_code == 200
    assert len(resp.json()) == 1
    assert resp.json()[0]["total_score"] == 27


@pytest.mark.asyncio
async def test_add_annotation(client, db_session):
    """Coach can annotate athlete session."""
    await seed_round_templates(db_session)
    coach_token, _ = await _register_and_get_token(client)
    athlete_token, athlete_username = await _register_and_get_token(client)

    # Setup coaching link
    invite = await client.post("/api/v1/coaching/invite", json={
        "athlete_username": athlete_username,
    }, headers={"Authorization": f"Bearer {coach_token}"})
    link_id = invite.json()["id"]
    await client.post("/api/v1/coaching/respond", json={
        "link_id": link_id, "accept": True,
    }, headers={"Authorization": f"Bearer {athlete_token}"})

    # Create session
    rounds_resp = await client.get("/api/v1/rounds", headers={"Authorization": f"Bearer {athlete_token}"})
    template = next(r for r in rounds_resp.json() if r["name"] == "WA Indoor 18m (Recurve)")
    s_resp = await client.post("/api/v1/sessions", json={"template_id": template["id"]},
                               headers={"Authorization": f"Bearer {athlete_token}"})
    session_id = s_resp.json()["id"]

    # Coach adds annotation
    resp = await client.post(f"/api/v1/coaching/sessions/{session_id}/annotations", json={
        "text": "Good grouping on this end!",
        "end_number": 1,
    }, headers={"Authorization": f"Bearer {coach_token}"})
    assert resp.status_code == 201
    assert resp.json()["text"] == "Good grouping on this end!"
    assert resp.json()["end_number"] == 1


@pytest.mark.asyncio
async def test_list_annotations(client, db_session):
    """List annotations for a session."""
    await seed_round_templates(db_session)
    coach_token, _ = await _register_and_get_token(client)
    athlete_token, athlete_username = await _register_and_get_token(client)

    invite = await client.post("/api/v1/coaching/invite", json={
        "athlete_username": athlete_username,
    }, headers={"Authorization": f"Bearer {coach_token}"})
    link_id = invite.json()["id"]
    await client.post("/api/v1/coaching/respond", json={
        "link_id": link_id, "accept": True,
    }, headers={"Authorization": f"Bearer {athlete_token}"})

    rounds_resp = await client.get("/api/v1/rounds", headers={"Authorization": f"Bearer {athlete_token}"})
    template = next(r for r in rounds_resp.json() if r["name"] == "WA Indoor 18m (Recurve)")
    s_resp = await client.post("/api/v1/sessions", json={"template_id": template["id"]},
                               headers={"Authorization": f"Bearer {athlete_token}"})
    session_id = s_resp.json()["id"]

    await client.post(f"/api/v1/coaching/sessions/{session_id}/annotations", json={
        "text": "Note 1",
    }, headers={"Authorization": f"Bearer {coach_token}"})
    await client.post(f"/api/v1/coaching/sessions/{session_id}/annotations", json={
        "text": "Note 2",
    }, headers={"Authorization": f"Bearer {athlete_token}"})

    resp = await client.get(f"/api/v1/coaching/sessions/{session_id}/annotations",
                            headers={"Authorization": f"Bearer {coach_token}"})
    assert resp.status_code == 200
    assert len(resp.json()) == 2
