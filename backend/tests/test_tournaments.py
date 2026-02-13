import secrets
import pytest

from app.seed.round_templates import seed_round_templates

_user_counter = 0


async def _register_and_get_token(client, email=None, username=None):
    global _user_counter
    _user_counter += 1
    email = email or f"tourney{_user_counter}@test.com"
    username = username or f"tourney{_user_counter}"
    resp = await client.post("/api/v1/auth/register", json={
        "email": email, "username": username, "password": "pass1234",
    })
    return resp.json()["access_token"]


async def _create_club(client, token, name="Test Club"):
    resp = await client.post("/api/v1/clubs", json={"name": name},
                             headers={"Authorization": f"Bearer {token}"})
    assert resp.status_code == 201
    return resp.json()["id"]


async def _create_invite(client, token, club_id):
    resp = await client.post(f"/api/v1/clubs/{club_id}/invites", json={},
                             headers={"Authorization": f"Bearer {token}"})
    assert resp.status_code == 201
    return resp.json()["code"]


async def _join_club(client, token, code):
    resp = await client.post(f"/api/v1/clubs/join/{code}",
                             headers={"Authorization": f"Bearer {token}"})
    assert resp.status_code == 200


async def _get_template_and_stage(client, token):
    """Get WA Indoor 18m (Recurve) template."""
    resp = await client.get("/api/v1/rounds", headers={"Authorization": f"Bearer {token}"})
    template = next(r for r in resp.json() if r["name"] == "WA Indoor 18m (Recurve)")
    stage = template["stages"][0]
    return template["id"], stage["id"], stage["arrows_per_end"]


async def _create_tournament(client, token, club_id, template_id, **kwargs):
    data = {
        "name": kwargs.get("name", "Test Tournament"),
        "template_id": template_id,
        **kwargs,
    }
    resp = await client.post(f"/api/v1/clubs/{club_id}/tournaments", json=data,
                             headers={"Authorization": f"Bearer {token}"})
    return resp


async def _score_session(client, token, template_id, stage_id, arrows_per_end, score_value="10"):
    """Create, score one end, and complete a session. Returns session_id."""
    headers = {"Authorization": f"Bearer {token}"}
    s_resp = await client.post("/api/v1/sessions", json={"template_id": template_id}, headers=headers)
    assert s_resp.status_code == 201, f"Session create failed: {s_resp.json()}"
    session_id = s_resp.json()["id"]
    arrows = [{"score_value": score_value} for _ in range(arrows_per_end)]
    end_resp = await client.post(f"/api/v1/sessions/{session_id}/ends", json={
        "stage_id": stage_id, "arrows": arrows,
    }, headers=headers)
    assert end_resp.status_code == 201, f"End submit failed: {end_resp.json()}"
    c_resp = await client.post(f"/api/v1/sessions/{session_id}/complete", headers=headers)
    assert c_resp.status_code == 200, f"Complete failed: {c_resp.json()}"
    return session_id


@pytest.mark.asyncio
async def test_create_tournament(client, db_session):
    """POST /clubs/{club_id}/tournaments creates a tournament."""
    await seed_round_templates(db_session)
    token = await _register_and_get_token(client)
    club_id = await _create_club(client, token)
    template_id, _, _ = await _get_template_and_stage(client, token)

    resp = await _create_tournament(client, token, club_id, template_id, name="Spring Shoot")
    assert resp.status_code == 201
    data = resp.json()
    assert data["name"] == "Spring Shoot"
    assert data["status"] == "registration"
    assert data["participant_count"] == 0
    assert data["club_id"] == club_id


@pytest.mark.asyncio
async def test_register_for_tournament(client, db_session):
    """POST /clubs/{club_id}/tournaments/{id}/register adds participant."""
    await seed_round_templates(db_session)
    organizer_token = await _register_and_get_token(client)
    archer_token = await _register_and_get_token(client)

    club_id = await _create_club(client, organizer_token)
    code = await _create_invite(client, organizer_token, club_id)
    await _join_club(client, archer_token, code)

    template_id, _, _ = await _get_template_and_stage(client, organizer_token)
    t_resp = await _create_tournament(client, organizer_token, club_id, template_id)
    t_id = t_resp.json()["id"]

    resp = await client.post(f"/api/v1/clubs/{club_id}/tournaments/{t_id}/register",
                             headers={"Authorization": f"Bearer {archer_token}"})
    assert resp.status_code == 201

    detail = await client.get(f"/api/v1/clubs/{club_id}/tournaments/{t_id}",
                              headers={"Authorization": f"Bearer {archer_token}"})
    assert len(detail.json()["participants"]) == 1


@pytest.mark.asyncio
async def test_start_tournament(client, db_session):
    """POST /clubs/{club_id}/tournaments/{id}/start changes status."""
    await seed_round_templates(db_session)
    organizer_token = await _register_and_get_token(client)
    archer_token = await _register_and_get_token(client)

    club_id = await _create_club(client, organizer_token)
    code = await _create_invite(client, organizer_token, club_id)
    await _join_club(client, archer_token, code)

    template_id, _, _ = await _get_template_and_stage(client, organizer_token)
    t_resp = await _create_tournament(client, organizer_token, club_id, template_id)
    t_id = t_resp.json()["id"]

    await client.post(f"/api/v1/clubs/{club_id}/tournaments/{t_id}/register",
                      headers={"Authorization": f"Bearer {archer_token}"})

    resp = await client.post(f"/api/v1/clubs/{club_id}/tournaments/{t_id}/start",
                             headers={"Authorization": f"Bearer {organizer_token}"})
    assert resp.status_code == 200
    assert resp.json()["status"] == "in_progress"


@pytest.mark.asyncio
async def test_submit_score_and_leaderboard(client, db_session):
    """Submit a score and verify leaderboard."""
    await seed_round_templates(db_session)
    organizer_token = await _register_and_get_token(client)
    archer_token = await _register_and_get_token(client)

    club_id = await _create_club(client, organizer_token)
    code = await _create_invite(client, organizer_token, club_id)
    await _join_club(client, archer_token, code)

    template_id, stage_id, arrows_per_end = await _get_template_and_stage(client, organizer_token)
    t_resp = await _create_tournament(client, organizer_token, club_id, template_id)
    t_id = t_resp.json()["id"]

    await client.post(f"/api/v1/clubs/{club_id}/tournaments/{t_id}/register",
                      headers={"Authorization": f"Bearer {archer_token}"})
    await client.post(f"/api/v1/clubs/{club_id}/tournaments/{t_id}/start",
                      headers={"Authorization": f"Bearer {organizer_token}"})

    session_id = await _score_session(client, archer_token, template_id, stage_id, arrows_per_end, "10")

    resp = await client.post(
        f"/api/v1/clubs/{club_id}/tournaments/{t_id}/submit-score?session_id={session_id}",
        headers={"Authorization": f"Bearer {archer_token}"},
    )
    assert resp.status_code == 200
    assert resp.json()["final_score"] == 30

    lb = await client.get(f"/api/v1/clubs/{club_id}/tournaments/{t_id}/leaderboard",
                          headers={"Authorization": f"Bearer {archer_token}"})
    assert lb.status_code == 200
    entries = lb.json()
    assert len(entries) == 1
    assert entries[0]["final_score"] == 30
    assert entries[0]["rank"] == 1


@pytest.mark.asyncio
async def test_complete_tournament(client, db_session):
    """POST /clubs/{club_id}/tournaments/{id}/complete finalizes ranks."""
    await seed_round_templates(db_session)
    organizer_token = await _register_and_get_token(client)
    archer1_token = await _register_and_get_token(client)
    archer2_token = await _register_and_get_token(client)

    club_id = await _create_club(client, organizer_token)
    code = await _create_invite(client, organizer_token, club_id)
    await _join_club(client, archer1_token, code)
    await _join_club(client, archer2_token, code)

    template_id, stage_id, arrows_per_end = await _get_template_and_stage(client, organizer_token)
    t_resp = await _create_tournament(client, organizer_token, club_id, template_id)
    t_id = t_resp.json()["id"]

    await client.post(f"/api/v1/clubs/{club_id}/tournaments/{t_id}/register",
                      headers={"Authorization": f"Bearer {archer1_token}"})
    await client.post(f"/api/v1/clubs/{club_id}/tournaments/{t_id}/register",
                      headers={"Authorization": f"Bearer {archer2_token}"})

    await client.post(f"/api/v1/clubs/{club_id}/tournaments/{t_id}/start",
                      headers={"Authorization": f"Bearer {organizer_token}"})

    s1_id = await _score_session(client, archer1_token, template_id, stage_id, arrows_per_end, "10")
    await client.post(f"/api/v1/clubs/{club_id}/tournaments/{t_id}/submit-score?session_id={s1_id}",
                      headers={"Authorization": f"Bearer {archer1_token}"})

    s2_id = await _score_session(client, archer2_token, template_id, stage_id, arrows_per_end, "5")
    await client.post(f"/api/v1/clubs/{club_id}/tournaments/{t_id}/submit-score?session_id={s2_id}",
                      headers={"Authorization": f"Bearer {archer2_token}"})

    resp = await client.post(f"/api/v1/clubs/{club_id}/tournaments/{t_id}/complete",
                             headers={"Authorization": f"Bearer {organizer_token}"})
    assert resp.status_code == 200
    assert resp.json()["status"] == "completed"

    lb = await client.get(f"/api/v1/clubs/{club_id}/tournaments/{t_id}/leaderboard",
                          headers={"Authorization": f"Bearer {organizer_token}"})
    entries = lb.json()
    assert entries[0]["rank"] == 1
    assert entries[0]["final_score"] == 30
    assert entries[1]["rank"] == 2
    assert entries[1]["final_score"] == 15


@pytest.mark.asyncio
async def test_withdraw_from_tournament(client, db_session):
    """POST /clubs/{club_id}/tournaments/{id}/withdraw sets status to withdrawn."""
    await seed_round_templates(db_session)
    organizer_token = await _register_and_get_token(client)
    archer_token = await _register_and_get_token(client)

    club_id = await _create_club(client, organizer_token)
    code = await _create_invite(client, organizer_token, club_id)
    await _join_club(client, archer_token, code)

    template_id, _, _ = await _get_template_and_stage(client, organizer_token)
    t_resp = await _create_tournament(client, organizer_token, club_id, template_id)
    t_id = t_resp.json()["id"]

    await client.post(f"/api/v1/clubs/{club_id}/tournaments/{t_id}/register",
                      headers={"Authorization": f"Bearer {archer_token}"})

    resp = await client.post(f"/api/v1/clubs/{club_id}/tournaments/{t_id}/withdraw",
                             headers={"Authorization": f"Bearer {archer_token}"})
    assert resp.status_code == 200

    detail = await client.get(f"/api/v1/clubs/{club_id}/tournaments/{t_id}",
                              headers={"Authorization": f"Bearer {archer_token}"})
    p = detail.json()["participants"][0]
    assert p["status"] == "withdrawn"


@pytest.mark.asyncio
async def test_cannot_register_twice(client, db_session):
    """Duplicate registration returns 409."""
    await seed_round_templates(db_session)
    organizer_token = await _register_and_get_token(client)
    archer_token = await _register_and_get_token(client)

    club_id = await _create_club(client, organizer_token)
    code = await _create_invite(client, organizer_token, club_id)
    await _join_club(client, archer_token, code)

    template_id, _, _ = await _get_template_and_stage(client, organizer_token)
    t_resp = await _create_tournament(client, organizer_token, club_id, template_id)
    t_id = t_resp.json()["id"]

    await client.post(f"/api/v1/clubs/{club_id}/tournaments/{t_id}/register",
                      headers={"Authorization": f"Bearer {archer_token}"})
    resp = await client.post(f"/api/v1/clubs/{club_id}/tournaments/{t_id}/register",
                             headers={"Authorization": f"Bearer {archer_token}"})
    assert resp.status_code == 409


@pytest.mark.asyncio
async def test_list_tournaments(client, db_session):
    """GET /clubs/{club_id}/tournaments lists tournaments."""
    await seed_round_templates(db_session)
    token = await _register_and_get_token(client)
    club_id = await _create_club(client, token)
    template_id, _, _ = await _get_template_and_stage(client, token)
    headers = {"Authorization": f"Bearer {token}"}

    await _create_tournament(client, token, club_id, template_id, name="T1")
    await _create_tournament(client, token, club_id, template_id, name="T2")

    resp = await client.get(f"/api/v1/clubs/{club_id}/tournaments", headers=headers)
    assert resp.status_code == 200
    assert len(resp.json()) == 2


@pytest.mark.asyncio
async def test_non_member_cannot_create_tournament(client, db_session):
    """Non-member cannot create a tournament in a club."""
    await seed_round_templates(db_session)
    owner_token = await _register_and_get_token(client)
    outsider_token = await _register_and_get_token(client)

    club_id = await _create_club(client, owner_token)
    template_id, _, _ = await _get_template_and_stage(client, owner_token)

    resp = await _create_tournament(client, outsider_token, club_id, template_id)
    assert resp.status_code in (401, 403)
