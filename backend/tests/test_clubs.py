import uuid
from datetime import datetime, timedelta, timezone

import pytest

from app.seed.round_templates import seed_round_templates


# ── Helpers ──────────────────────────────────────────────────────────────


async def _register(client, email, username):
    """Register a user and return (token, headers)."""
    resp = await client.post("/api/v1/auth/register", json={
        "email": email, "username": username, "password": "pass1234",
    })
    token = resp.json()["access_token"]
    return token, {"Authorization": f"Bearer {token}"}


async def _create_club(client, headers, name="Test Club"):
    resp = await client.post("/api/v1/clubs", json={"name": name, "description": "A test club"}, headers=headers)
    assert resp.status_code == 201
    return resp.json()


async def _create_invite(client, headers, club_id, **kwargs):
    resp = await client.post(f"/api/v1/clubs/{club_id}/invites", json=kwargs, headers=headers)
    assert resp.status_code == 201
    return resp.json()


async def _get_user_id(client, headers):
    resp = await client.get("/api/v1/users/me", headers=headers)
    return resp.json()["id"]


# ── Club CRUD ────────────────────────────────────────────────────────────


@pytest.mark.asyncio
async def test_create_club(client):
    _, headers = await _register(client, "owner@club.com", "clubowner")
    club = await _create_club(client, headers)
    assert club["name"] == "Test Club"
    assert club["my_role"] == "owner"
    assert club["member_count"] == 1


@pytest.mark.asyncio
async def test_list_my_clubs(client):
    _, headers = await _register(client, "lister@club.com", "clublister")
    await _create_club(client, headers, "Club A")
    await _create_club(client, headers, "Club B")

    resp = await client.get("/api/v1/clubs", headers=headers)
    assert resp.status_code == 200
    assert len(resp.json()) == 2


@pytest.mark.asyncio
async def test_get_club_detail(client):
    _, headers = await _register(client, "detail@club.com", "clubdetail")
    club = await _create_club(client, headers)

    resp = await client.get(f"/api/v1/clubs/{club['id']}", headers=headers)
    assert resp.status_code == 200
    data = resp.json()
    assert data["name"] == "Test Club"
    assert len(data["members"]) == 1
    assert data["members"][0]["role"] == "owner"


@pytest.mark.asyncio
async def test_update_club(client):
    _, headers = await _register(client, "updater@club.com", "clubupdater")
    club = await _create_club(client, headers)

    resp = await client.patch(f"/api/v1/clubs/{club['id']}", json={"name": "Renamed Club"}, headers=headers)
    assert resp.status_code == 200
    assert resp.json()["name"] == "Renamed Club"


@pytest.mark.asyncio
async def test_update_club_non_owner_forbidden(client):
    _, owner_h = await _register(client, "own@up.com", "ownup")
    _, member_h = await _register(client, "mem@up.com", "memup")
    club = await _create_club(client, owner_h)

    # Join as member
    invite = await _create_invite(client, owner_h, club["id"])
    await client.post(f"/api/v1/clubs/join/{invite['code']}", headers=member_h)

    resp = await client.patch(f"/api/v1/clubs/{club['id']}", json={"name": "Nope"}, headers=member_h)
    assert resp.status_code == 401


@pytest.mark.asyncio
async def test_delete_club(client):
    _, headers = await _register(client, "deleter@club.com", "clubdeleter")
    club = await _create_club(client, headers)

    resp = await client.delete(f"/api/v1/clubs/{club['id']}", headers=headers)
    assert resp.status_code == 204

    # Gone
    resp = await client.get(f"/api/v1/clubs/{club['id']}", headers=headers)
    assert resp.status_code == 404


@pytest.mark.asyncio
async def test_delete_club_non_owner_forbidden(client):
    _, owner_h = await _register(client, "own@del.com", "owndel")
    _, member_h = await _register(client, "mem@del.com", "memdel")
    club = await _create_club(client, owner_h)

    invite = await _create_invite(client, owner_h, club["id"])
    await client.post(f"/api/v1/clubs/join/{invite['code']}", headers=member_h)

    resp = await client.delete(f"/api/v1/clubs/{club['id']}", headers=member_h)
    assert resp.status_code == 401


# ── Invites ──────────────────────────────────────────────────────────────


@pytest.mark.asyncio
async def test_invite_flow(client):
    _, owner_h = await _register(client, "own@inv.com", "owninv")
    _, joiner_h = await _register(client, "join@inv.com", "joininv")
    club = await _create_club(client, owner_h)

    invite = await _create_invite(client, owner_h, club["id"])
    assert invite["active"] is True
    assert invite["code"]

    # List invites
    resp = await client.get(f"/api/v1/clubs/{club['id']}/invites", headers=owner_h)
    assert resp.status_code == 200
    assert len(resp.json()) == 1

    # Preview invite
    resp = await client.get(f"/api/v1/clubs/join/{invite['code']}", headers=joiner_h)
    assert resp.status_code == 200
    assert resp.json()["name"] == "Test Club"

    # Join via invite
    resp = await client.post(f"/api/v1/clubs/join/{invite['code']}", headers=joiner_h)
    assert resp.status_code == 200
    assert resp.json()["role"] == "member"

    # Duplicate join
    resp = await client.post(f"/api/v1/clubs/join/{invite['code']}", headers=joiner_h)
    assert resp.status_code == 409


@pytest.mark.asyncio
async def test_invite_invalid_code(client):
    _, headers = await _register(client, "bad@inv.com", "badinv")
    resp = await client.get("/api/v1/clubs/join/nonexistentcode", headers=headers)
    assert resp.status_code == 404


@pytest.mark.asyncio
async def test_deactivate_invite(client):
    _, headers = await _register(client, "deact@inv.com", "deactinv")
    club = await _create_club(client, headers)
    invite = await _create_invite(client, headers, club["id"])

    resp = await client.delete(f"/api/v1/clubs/{club['id']}/invites/{invite['id']}", headers=headers)
    assert resp.status_code == 204

    # Invite no longer valid
    resp = await client.get(f"/api/v1/clubs/join/{invite['code']}", headers=headers)
    assert resp.status_code == 404


@pytest.mark.asyncio
async def test_invite_member_cannot_create(client):
    _, owner_h = await _register(client, "own@perm.com", "ownperm")
    _, member_h = await _register(client, "mem@perm.com", "memperm")
    club = await _create_club(client, owner_h)

    invite = await _create_invite(client, owner_h, club["id"])
    await client.post(f"/api/v1/clubs/join/{invite['code']}", headers=member_h)

    resp = await client.post(f"/api/v1/clubs/{club['id']}/invites", json={}, headers=member_h)
    assert resp.status_code == 401


# ── Member management ────────────────────────────────────────────────────


@pytest.mark.asyncio
async def test_promote_and_demote_member(client):
    _, owner_h = await _register(client, "own@promo.com", "ownpromo")
    _, member_h = await _register(client, "mem@promo.com", "mempromo")
    club = await _create_club(client, owner_h)

    invite = await _create_invite(client, owner_h, club["id"])
    await client.post(f"/api/v1/clubs/join/{invite['code']}", headers=member_h)

    member_id = await _get_user_id(client, member_h)

    # Promote
    resp = await client.post(f"/api/v1/clubs/{club['id']}/members/{member_id}/promote", headers=owner_h)
    assert resp.status_code == 200

    # Verify admin in club detail
    resp = await client.get(f"/api/v1/clubs/{club['id']}", headers=owner_h)
    members = {m["user_id"]: m for m in resp.json()["members"]}
    assert members[member_id]["role"] == "admin"

    # Demote
    resp = await client.post(f"/api/v1/clubs/{club['id']}/members/{member_id}/demote", headers=owner_h)
    assert resp.status_code == 200


@pytest.mark.asyncio
async def test_promote_requires_owner(client):
    _, owner_h = await _register(client, "own@preq.com", "ownpreq")
    _, admin_h = await _register(client, "adm@preq.com", "admpreq")
    _, member_h = await _register(client, "mem@preq.com", "mempreq")
    club = await _create_club(client, owner_h)

    invite = await _create_invite(client, owner_h, club["id"])
    await client.post(f"/api/v1/clubs/join/{invite['code']}", headers=admin_h)
    await client.post(f"/api/v1/clubs/join/{invite['code']}", headers=member_h)

    admin_id = await _get_user_id(client, admin_h)
    member_id = await _get_user_id(client, member_h)

    # Promote admin_h first
    await client.post(f"/api/v1/clubs/{club['id']}/members/{admin_id}/promote", headers=owner_h)

    # Admin cannot promote
    resp = await client.post(f"/api/v1/clubs/{club['id']}/members/{member_id}/promote", headers=admin_h)
    assert resp.status_code == 401


@pytest.mark.asyncio
async def test_remove_member(client):
    _, owner_h = await _register(client, "own@rem.com", "ownrem")
    _, member_h = await _register(client, "mem@rem.com", "memrem")
    club = await _create_club(client, owner_h)

    invite = await _create_invite(client, owner_h, club["id"])
    await client.post(f"/api/v1/clubs/join/{invite['code']}", headers=member_h)

    member_id = await _get_user_id(client, member_h)

    resp = await client.delete(f"/api/v1/clubs/{club['id']}/members/{member_id}", headers=owner_h)
    assert resp.status_code == 204


@pytest.mark.asyncio
async def test_member_leave_club(client):
    _, owner_h = await _register(client, "own@leave.com", "ownleave")
    _, member_h = await _register(client, "mem@leave.com", "memleave")
    club = await _create_club(client, owner_h)

    invite = await _create_invite(client, owner_h, club["id"])
    await client.post(f"/api/v1/clubs/join/{invite['code']}", headers=member_h)

    member_id = await _get_user_id(client, member_h)
    resp = await client.delete(f"/api/v1/clubs/{club['id']}/members/{member_id}", headers=member_h)
    assert resp.status_code == 204


@pytest.mark.asyncio
async def test_owner_cannot_leave(client):
    _, owner_h = await _register(client, "own@stay.com", "ownstay")
    club = await _create_club(client, owner_h)
    owner_id = await _get_user_id(client, owner_h)

    resp = await client.delete(f"/api/v1/clubs/{club['id']}/members/{owner_id}", headers=owner_h)
    assert resp.status_code == 422


@pytest.mark.asyncio
async def test_non_member_access_denied(client):
    _, owner_h = await _register(client, "own@acc.com", "ownacc")
    _, outsider_h = await _register(client, "out@acc.com", "outacc")
    club = await _create_club(client, owner_h)

    resp = await client.get(f"/api/v1/clubs/{club['id']}", headers=outsider_h)
    assert resp.status_code == 401


# ── Leaderboard & Activity ──────────────────────────────────────────────


@pytest.mark.asyncio
async def test_leaderboard_empty(client, db_session):
    await seed_round_templates(db_session)
    _, headers = await _register(client, "lb@club.com", "lbuser")
    club = await _create_club(client, headers)

    resp = await client.get(f"/api/v1/clubs/{club['id']}/leaderboard", headers=headers)
    assert resp.status_code == 200
    assert resp.json() == []


@pytest.mark.asyncio
async def test_leaderboard_with_scores(client, db_session):
    await seed_round_templates(db_session)
    _, owner_h = await _register(client, "lbo@club.com", "lbouser")
    _, member_h = await _register(client, "lbm@club.com", "lbmuser")
    club = await _create_club(client, owner_h)

    invite = await _create_invite(client, owner_h, club["id"])
    await client.post(f"/api/v1/clubs/join/{invite['code']}", headers=member_h)

    # Owner completes a session
    rounds = (await client.get("/api/v1/rounds")).json()
    vegas = next(r for r in rounds if r["name"] == "Vegas 300")
    stage_id = vegas["stages"][0]["id"]

    resp = await client.post("/api/v1/sessions", json={"template_id": vegas["id"]}, headers=owner_h)
    sid = resp.json()["id"]
    await client.post(f"/api/v1/sessions/{sid}/ends", json={
        "stage_id": stage_id,
        "arrows": [{"score_value": "X"}, {"score_value": "X"}, {"score_value": "X"}],
    }, headers=owner_h)
    await client.post(f"/api/v1/sessions/{sid}/complete", headers=owner_h)

    resp = await client.get(f"/api/v1/clubs/{club['id']}/leaderboard", headers=owner_h)
    assert resp.status_code == 200
    data = resp.json()
    assert len(data) == 1
    assert data[0]["template_name"] == "Vegas 300"
    assert len(data[0]["entries"]) == 1
    assert data[0]["entries"][0]["best_score"] == 30


@pytest.mark.asyncio
async def test_activity_feed(client, db_session):
    await seed_round_templates(db_session)
    _, headers = await _register(client, "act@club.com", "actuser")
    club = await _create_club(client, headers)

    resp = await client.get(f"/api/v1/clubs/{club['id']}/activity", headers=headers)
    assert resp.status_code == 200
    assert isinstance(resp.json(), list)


# ── Events ───────────────────────────────────────────────────────────────


@pytest.mark.asyncio
async def test_event_crud(client, db_session):
    await seed_round_templates(db_session)
    _, owner_h = await _register(client, "evtown@club.com", "evtown")
    club = await _create_club(client, owner_h)

    rounds = (await client.get("/api/v1/rounds")).json()
    template_id = rounds[0]["id"]
    event_date = (datetime.now(timezone.utc) + timedelta(days=7)).isoformat()

    # Create event
    resp = await client.post(f"/api/v1/clubs/{club['id']}/events", json={
        "name": "Weekend Shoot",
        "template_id": template_id,
        "event_date": event_date,
        "location": "Range A",
    }, headers=owner_h)
    assert resp.status_code == 201
    event = resp.json()
    assert event["name"] == "Weekend Shoot"
    assert event["location"] == "Range A"
    event_id = event["id"]

    # List events
    resp = await client.get(f"/api/v1/clubs/{club['id']}/events", headers=owner_h)
    assert resp.status_code == 200
    assert len(resp.json()) == 1

    # Get event
    resp = await client.get(f"/api/v1/clubs/{club['id']}/events/{event_id}", headers=owner_h)
    assert resp.status_code == 200
    assert resp.json()["name"] == "Weekend Shoot"

    # Update event
    resp = await client.patch(f"/api/v1/clubs/{club['id']}/events/{event_id}", json={
        "name": "Updated Shoot",
    }, headers=owner_h)
    assert resp.status_code == 200
    assert resp.json()["name"] == "Updated Shoot"

    # Delete event
    resp = await client.delete(f"/api/v1/clubs/{club['id']}/events/{event_id}", headers=owner_h)
    assert resp.status_code == 204


@pytest.mark.asyncio
async def test_event_rsvp(client, db_session):
    await seed_round_templates(db_session)
    _, owner_h = await _register(client, "rsvpo@club.com", "rsvpown")
    _, member_h = await _register(client, "rsvpm@club.com", "rsvpmem")
    club = await _create_club(client, owner_h)

    invite = await _create_invite(client, owner_h, club["id"])
    await client.post(f"/api/v1/clubs/join/{invite['code']}", headers=member_h)

    rounds = (await client.get("/api/v1/rounds")).json()
    template_id = rounds[0]["id"]
    event_date = (datetime.now(timezone.utc) + timedelta(days=7)).isoformat()

    resp = await client.post(f"/api/v1/clubs/{club['id']}/events", json={
        "name": "RSVP Test", "template_id": template_id, "event_date": event_date,
    }, headers=owner_h)
    event_id = resp.json()["id"]

    # RSVP as going
    resp = await client.post(f"/api/v1/clubs/{club['id']}/events/{event_id}/rsvp", json={
        "status": "going",
    }, headers=member_h)
    assert resp.status_code == 200
    participants = resp.json()["participants"]
    assert len(participants) == 1
    assert participants[0]["status"] == "going"

    # Change RSVP to maybe
    resp = await client.post(f"/api/v1/clubs/{club['id']}/events/{event_id}/rsvp", json={
        "status": "maybe",
    }, headers=member_h)
    assert resp.status_code == 200
    assert resp.json()["participants"][0]["status"] == "maybe"


@pytest.mark.asyncio
async def test_event_member_cannot_create(client, db_session):
    await seed_round_templates(db_session)
    _, owner_h = await _register(client, "evtno@club.com", "evtnoown")
    _, member_h = await _register(client, "evtnm@club.com", "evtnomem")
    club = await _create_club(client, owner_h)

    invite = await _create_invite(client, owner_h, club["id"])
    await client.post(f"/api/v1/clubs/join/{invite['code']}", headers=member_h)

    rounds = (await client.get("/api/v1/rounds")).json()
    event_date = (datetime.now(timezone.utc) + timedelta(days=7)).isoformat()

    resp = await client.post(f"/api/v1/clubs/{club['id']}/events", json={
        "name": "Blocked", "template_id": rounds[0]["id"], "event_date": event_date,
    }, headers=member_h)
    assert resp.status_code == 401


# ── Teams ────────────────────────────────────────────────────────────────


@pytest.mark.asyncio
async def test_team_crud(client):
    _, owner_h = await _register(client, "tmown@club.com", "tmown")
    club = await _create_club(client, owner_h)
    owner_id = await _get_user_id(client, owner_h)

    # Create team
    resp = await client.post(f"/api/v1/clubs/{club['id']}/teams", json={
        "name": "Alpha Team", "leader_id": owner_id,
    }, headers=owner_h)
    assert resp.status_code == 201
    team = resp.json()
    assert team["name"] == "Alpha Team"
    assert team["leader"]["user_id"] == owner_id
    team_id = team["id"]

    # List teams
    resp = await client.get(f"/api/v1/clubs/{club['id']}/teams", headers=owner_h)
    assert resp.status_code == 200
    assert len(resp.json()) == 1

    # Get team detail
    resp = await client.get(f"/api/v1/clubs/{club['id']}/teams/{team_id}", headers=owner_h)
    assert resp.status_code == 200
    assert resp.json()["name"] == "Alpha Team"
    assert "members" in resp.json()

    # Update team
    resp = await client.patch(f"/api/v1/clubs/{club['id']}/teams/{team_id}", json={
        "name": "Beta Team",
    }, headers=owner_h)
    assert resp.status_code == 200
    assert resp.json()["name"] == "Beta Team"

    # Delete team
    resp = await client.delete(f"/api/v1/clubs/{club['id']}/teams/{team_id}", headers=owner_h)
    assert resp.status_code == 204


@pytest.mark.asyncio
async def test_team_add_and_remove_member(client):
    _, owner_h = await _register(client, "tmao@club.com", "tmaoown")
    _, member_h = await _register(client, "tmam@club.com", "tmaomem")
    club = await _create_club(client, owner_h)
    owner_id = await _get_user_id(client, owner_h)
    member_id = await _get_user_id(client, member_h)

    invite = await _create_invite(client, owner_h, club["id"])
    await client.post(f"/api/v1/clubs/join/{invite['code']}", headers=member_h)

    # Create team
    resp = await client.post(f"/api/v1/clubs/{club['id']}/teams", json={
        "name": "Team X", "leader_id": owner_id,
    }, headers=owner_h)
    team_id = resp.json()["id"]

    # Add member to team
    resp = await client.post(
        f"/api/v1/clubs/{club['id']}/teams/{team_id}/members/{member_id}", headers=owner_h,
    )
    assert resp.status_code == 201

    # Verify in team detail
    resp = await client.get(f"/api/v1/clubs/{club['id']}/teams/{team_id}", headers=owner_h)
    assert len(resp.json()["members"]) == 1

    # Duplicate add
    resp = await client.post(
        f"/api/v1/clubs/{club['id']}/teams/{team_id}/members/{member_id}", headers=owner_h,
    )
    assert resp.status_code == 409

    # Remove from team
    resp = await client.delete(
        f"/api/v1/clubs/{club['id']}/teams/{team_id}/members/{member_id}", headers=owner_h,
    )
    assert resp.status_code == 204

    # Remove again → 404
    resp = await client.delete(
        f"/api/v1/clubs/{club['id']}/teams/{team_id}/members/{member_id}", headers=owner_h,
    )
    assert resp.status_code == 404


@pytest.mark.asyncio
async def test_team_leader_must_be_member(client):
    _, owner_h = await _register(client, "tmldr@club.com", "tmldr")
    club = await _create_club(client, owner_h)

    fake_id = str(uuid.uuid4())
    resp = await client.post(f"/api/v1/clubs/{club['id']}/teams", json={
        "name": "Bad Team", "leader_id": fake_id,
    }, headers=owner_h)
    assert resp.status_code == 422


@pytest.mark.asyncio
async def test_team_non_club_member_denied(client):
    _, owner_h = await _register(client, "tmnd@club.com", "tmndown")
    _, outsider_h = await _register(client, "tmno@club.com", "tmnoout")
    club = await _create_club(client, owner_h)
    owner_id = await _get_user_id(client, owner_h)

    resp = await client.post(f"/api/v1/clubs/{club['id']}/teams", json={
        "name": "Nope", "leader_id": owner_id,
    }, headers=outsider_h)
    assert resp.status_code == 401


@pytest.mark.asyncio
async def test_team_member_cannot_create(client):
    _, owner_h = await _register(client, "tmmc@club.com", "tmmcown")
    _, member_h = await _register(client, "tmme@club.com", "tmmcmem")
    club = await _create_club(client, owner_h)
    owner_id = await _get_user_id(client, owner_h)

    invite = await _create_invite(client, owner_h, club["id"])
    await client.post(f"/api/v1/clubs/join/{invite['code']}", headers=member_h)

    resp = await client.post(f"/api/v1/clubs/{club['id']}/teams", json={
        "name": "Blocked", "leader_id": owner_id,
    }, headers=member_h)
    assert resp.status_code == 401
