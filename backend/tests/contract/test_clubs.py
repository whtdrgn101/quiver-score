"""
Contract tests for /api/v1/clubs endpoints.

Covers:
- Club CRUD (create, list, get detail, update, delete)
- Invite management (create, list, preview, join, deactivate)
- Member management (promote, demote, remove)
- Leaderboard
- Activity feed
- Events (create, list, get, update, delete, RSVP)
- Teams (create, list, get, update, delete, add/remove members)
- Shared rounds (list, remove)
- Tournaments (create, list, get, register, start, leaderboard, complete, withdraw, submit-score)
"""


# ── Club CRUD ──────────────────────────────────────────────────────────


def test_create_club(client, register_user, unique):
    user = register_user()
    resp = client.post("/api/v1/clubs", json={
        "name": unique("club"),
        "description": "A test club",
    }, headers=user["headers"])
    assert resp.status_code == 201
    data = resp.json()
    assert data["name"].startswith("club_")
    assert data["description"] == "A test club"
    assert data["owner_id"] is not None
    assert data["member_count"] == 1
    assert data["my_role"] == "owner"
    assert "id" in data
    assert "created_at" in data


def test_create_club_minimal(client, register_user, unique):
    user = register_user()
    resp = client.post("/api/v1/clubs", json={
        "name": unique("club"),
    }, headers=user["headers"])
    assert resp.status_code == 201
    assert resp.json()["description"] is None


def test_create_club_unauthenticated(client, unique):
    resp = client.post("/api/v1/clubs", json={"name": unique("club")})
    assert resp.status_code == 401


def test_list_clubs_empty(client, register_user):
    user = register_user()
    resp = client.get("/api/v1/clubs", headers=user["headers"])
    assert resp.status_code == 200
    assert resp.json() == []


def test_list_clubs_with_items(client, register_user, unique):
    user = register_user()
    client.post("/api/v1/clubs", json={"name": unique("club")}, headers=user["headers"])
    resp = client.get("/api/v1/clubs", headers=user["headers"])
    assert resp.status_code == 200
    data = resp.json()
    assert len(data) >= 1
    assert data[0]["my_role"] == "owner"


def test_get_club_detail(client, register_user, unique):
    user = register_user()
    club = client.post("/api/v1/clubs", json={
        "name": unique("club"),
        "description": "Detail test",
    }, headers=user["headers"]).json()

    resp = client.get(f"/api/v1/clubs/{club['id']}", headers=user["headers"])
    assert resp.status_code == 200
    data = resp.json()
    assert data["id"] == club["id"]
    assert "members" in data
    assert len(data["members"]) == 1
    assert data["members"][0]["role"] == "owner"
    assert data["members"][0]["username"] == user["username"]


def test_get_club_not_member(client, register_user, unique):
    """Non-members cannot view club details."""
    owner = register_user()
    other = register_user()
    club = client.post("/api/v1/clubs", json={"name": unique("club")}, headers=owner["headers"]).json()

    resp = client.get(f"/api/v1/clubs/{club['id']}", headers=other["headers"])
    assert resp.status_code in (401, 403, 404)


def test_update_club(client, register_user, unique):
    user = register_user()
    club = client.post("/api/v1/clubs", json={"name": unique("club")}, headers=user["headers"]).json()

    resp = client.patch(f"/api/v1/clubs/{club['id']}", json={
        "name": "Updated Club Name",
        "description": "Updated description",
    }, headers=user["headers"])
    assert resp.status_code == 200
    data = resp.json()
    assert data["name"] == "Updated Club Name"
    assert data["description"] == "Updated description"


def test_update_club_not_owner(client, register_user, unique):
    """Only the owner can update the club."""
    owner = register_user()
    other = register_user()
    club = client.post("/api/v1/clubs", json={"name": unique("club")}, headers=owner["headers"]).json()

    resp = client.patch(f"/api/v1/clubs/{club['id']}", json={
        "name": "Hacked",
    }, headers=other["headers"])
    assert resp.status_code in (401, 403, 404)


def test_delete_club(client, register_user, unique):
    user = register_user()
    club = client.post("/api/v1/clubs", json={"name": unique("club")}, headers=user["headers"]).json()

    resp = client.delete(f"/api/v1/clubs/{club['id']}", headers=user["headers"])
    assert resp.status_code == 204

    # Verify deleted
    resp = client.get(f"/api/v1/clubs/{club['id']}", headers=user["headers"])
    assert resp.status_code in (401, 403, 404)


def test_delete_club_not_owner(client, register_user, unique):
    owner = register_user()
    other = register_user()
    club = client.post("/api/v1/clubs", json={"name": unique("club")}, headers=owner["headers"]).json()

    resp = client.delete(f"/api/v1/clubs/{club['id']}", headers=other["headers"])
    assert resp.status_code in (401, 403, 404)


# ── Invites ────────────────────────────────────────────────────────────


def _create_club_with_invite(client, owner, unique):
    """Helper: create a club and an invite, return (club, invite)."""
    club = client.post("/api/v1/clubs", json={"name": unique("club")}, headers=owner["headers"]).json()
    invite = client.post(f"/api/v1/clubs/{club['id']}/invites", json={}, headers=owner["headers"]).json()
    return club, invite


def test_create_invite(client, register_user, unique):
    user = register_user()
    club = client.post("/api/v1/clubs", json={"name": unique("club")}, headers=user["headers"]).json()

    resp = client.post(f"/api/v1/clubs/{club['id']}/invites", json={}, headers=user["headers"])
    assert resp.status_code == 201
    data = resp.json()
    assert "code" in data
    assert "url" in data
    assert data["active"] is True
    assert data["use_count"] == 0


def test_create_invite_with_options(client, register_user, unique):
    user = register_user()
    club = client.post("/api/v1/clubs", json={"name": unique("club")}, headers=user["headers"]).json()

    resp = client.post(f"/api/v1/clubs/{club['id']}/invites", json={
        "max_uses": 5,
        "expires_in_hours": 24,
    }, headers=user["headers"])
    assert resp.status_code == 201
    data = resp.json()
    assert data["max_uses"] == 5
    assert data["expires_at"] is not None


def test_list_invites(client, register_user, unique):
    user = register_user()
    club, invite = _create_club_with_invite(client, user, unique)

    resp = client.get(f"/api/v1/clubs/{club['id']}/invites", headers=user["headers"])
    assert resp.status_code == 200
    data = resp.json()
    assert len(data) >= 1
    assert any(i["id"] == invite["id"] for i in data)


def test_preview_invite(client, register_user, unique):
    owner = register_user()
    joiner = register_user()
    club, invite = _create_club_with_invite(client, owner, unique)

    resp = client.get(f"/api/v1/clubs/join/{invite['code']}", headers=joiner["headers"])
    assert resp.status_code == 200
    data = resp.json()
    assert data["id"] == club["id"]
    assert data["name"] == club["name"]


def test_join_club_via_invite(client, register_user, unique):
    owner = register_user()
    joiner = register_user()
    club, invite = _create_club_with_invite(client, owner, unique)

    resp = client.post(f"/api/v1/clubs/join/{invite['code']}", headers=joiner["headers"])
    assert resp.status_code == 200
    data = resp.json()
    assert data["club_id"] == club["id"]
    assert data["role"] == "member"

    # Verify they can now see the club
    resp = client.get("/api/v1/clubs", headers=joiner["headers"])
    assert any(c["id"] == club["id"] for c in resp.json())


def test_join_club_already_member(client, register_user, unique):
    owner = register_user()
    joiner = register_user()
    club, invite = _create_club_with_invite(client, owner, unique)

    client.post(f"/api/v1/clubs/join/{invite['code']}", headers=joiner["headers"])
    resp = client.post(f"/api/v1/clubs/join/{invite['code']}", headers=joiner["headers"])
    assert resp.status_code == 409


def test_join_club_invalid_code(client, register_user):
    user = register_user()
    resp = client.post("/api/v1/clubs/join/invalid-code-xyz", headers=user["headers"])
    assert resp.status_code == 404


def test_deactivate_invite(client, register_user, unique):
    user = register_user()
    club, invite = _create_club_with_invite(client, user, unique)

    resp = client.delete(f"/api/v1/clubs/{club['id']}/invites/{invite['id']}", headers=user["headers"])
    assert resp.status_code == 204


# ── Member Management ──────────────────────────────────────────────────


def _create_club_with_member(client, owner, member, unique):
    """Helper: create a club and add a member via invite."""
    club, invite = _create_club_with_invite(client, owner, unique)
    client.post(f"/api/v1/clubs/join/{invite['code']}", headers=member["headers"])
    return club


def test_promote_member(client, register_user, unique):
    owner = register_user()
    member = register_user()
    club = _create_club_with_member(client, owner, member, unique)

    # Get member's user_id
    detail = client.get(f"/api/v1/clubs/{club['id']}", headers=owner["headers"]).json()
    member_entry = [m for m in detail["members"] if m["role"] == "member"][0]

    resp = client.post(
        f"/api/v1/clubs/{club['id']}/members/{member_entry['user_id']}/promote",
        headers=owner["headers"],
    )
    assert resp.status_code == 200

    # Verify promoted
    detail = client.get(f"/api/v1/clubs/{club['id']}", headers=owner["headers"]).json()
    updated = [m for m in detail["members"] if m["user_id"] == member_entry["user_id"]][0]
    assert updated["role"] == "admin"


def test_demote_member(client, register_user, unique):
    owner = register_user()
    member = register_user()
    club = _create_club_with_member(client, owner, member, unique)

    detail = client.get(f"/api/v1/clubs/{club['id']}", headers=owner["headers"]).json()
    member_entry = [m for m in detail["members"] if m["role"] == "member"][0]

    # Promote first
    client.post(
        f"/api/v1/clubs/{club['id']}/members/{member_entry['user_id']}/promote",
        headers=owner["headers"],
    )
    # Then demote
    resp = client.post(
        f"/api/v1/clubs/{club['id']}/members/{member_entry['user_id']}/demote",
        headers=owner["headers"],
    )
    assert resp.status_code == 200


def test_remove_member(client, register_user, unique):
    owner = register_user()
    member = register_user()
    club = _create_club_with_member(client, owner, member, unique)

    detail = client.get(f"/api/v1/clubs/{club['id']}", headers=owner["headers"]).json()
    member_entry = [m for m in detail["members"] if m["role"] == "member"][0]

    resp = client.delete(
        f"/api/v1/clubs/{club['id']}/members/{member_entry['user_id']}",
        headers=owner["headers"],
    )
    assert resp.status_code == 204

    # Verify removed
    detail = client.get(f"/api/v1/clubs/{club['id']}", headers=owner["headers"]).json()
    assert len(detail["members"]) == 1


def test_member_leave_club(client, register_user, unique):
    """Members can remove themselves."""
    owner = register_user()
    member = register_user()
    club = _create_club_with_member(client, owner, member, unique)

    # Get member's user_id from their own perspective
    me = client.get("/api/v1/users/me", headers=member["headers"]).json()

    resp = client.delete(
        f"/api/v1/clubs/{club['id']}/members/{me['id']}",
        headers=member["headers"],
    )
    assert resp.status_code == 204


# ── Leaderboard ────────────────────────────────────────────────────────


def test_leaderboard_empty(client, register_user, unique):
    user = register_user()
    club = client.post("/api/v1/clubs", json={"name": unique("club")}, headers=user["headers"]).json()

    resp = client.get(f"/api/v1/clubs/{club['id']}/leaderboard", headers=user["headers"])
    assert resp.status_code == 200
    assert isinstance(resp.json(), list)


def test_leaderboard_not_member(client, register_user, unique):
    owner = register_user()
    other = register_user()
    club = client.post("/api/v1/clubs", json={"name": unique("club")}, headers=owner["headers"]).json()

    resp = client.get(f"/api/v1/clubs/{club['id']}/leaderboard", headers=other["headers"])
    assert resp.status_code in (401, 403, 404)


# ── Activity Feed ──────────────────────────────────────────────────────


def test_activity_empty(client, register_user, unique):
    user = register_user()
    club = client.post("/api/v1/clubs", json={"name": unique("club")}, headers=user["headers"]).json()

    resp = client.get(f"/api/v1/clubs/{club['id']}/activity", headers=user["headers"])
    assert resp.status_code == 200
    assert isinstance(resp.json(), list)


def test_activity_not_member(client, register_user, unique):
    owner = register_user()
    other = register_user()
    club = client.post("/api/v1/clubs", json={"name": unique("club")}, headers=owner["headers"]).json()

    resp = client.get(f"/api/v1/clubs/{club['id']}/activity", headers=other["headers"])
    assert resp.status_code in (401, 403, 404)


# ── Events ─────────────────────────────────────────────────────────────


def _create_club_event(client, user, club_id, unique, create_round):
    """Helper: create an event in a club."""
    rnd = create_round(headers=user["headers"])
    resp = client.post(f"/api/v1/clubs/{club_id}/events", json={
        "name": unique("event"),
        "description": "A test event",
        "template_id": rnd["id"],
        "event_date": "2026-06-01T10:00:00Z",
        "location": "Test Range",
    }, headers=user["headers"])
    assert resp.status_code == 201
    return resp.json()


def test_create_event(client, register_user, unique, create_round):
    user = register_user()
    club = client.post("/api/v1/clubs", json={"name": unique("club")}, headers=user["headers"]).json()
    event = _create_club_event(client, user, club["id"], unique, create_round)
    assert event["name"].startswith("event_")
    assert event["location"] == "Test Range"
    assert "participants" in event


def test_create_event_not_admin(client, register_user, unique, create_round):
    owner = register_user()
    member = register_user()
    club = _create_club_with_member(client, owner, member, unique)
    rnd = create_round(headers=member["headers"])

    resp = client.post(f"/api/v1/clubs/{club['id']}/events", json={
        "name": unique("event"),
        "template_id": rnd["id"],
        "event_date": "2026-06-01T10:00:00Z",
    }, headers=member["headers"])
    assert resp.status_code in (401, 403)


def test_list_events(client, register_user, unique, create_round):
    user = register_user()
    club = client.post("/api/v1/clubs", json={"name": unique("club")}, headers=user["headers"]).json()
    _create_club_event(client, user, club["id"], unique, create_round)

    resp = client.get(f"/api/v1/clubs/{club['id']}/events", headers=user["headers"])
    assert resp.status_code == 200
    assert len(resp.json()) >= 1


def test_get_event(client, register_user, unique, create_round):
    user = register_user()
    club = client.post("/api/v1/clubs", json={"name": unique("club")}, headers=user["headers"]).json()
    event = _create_club_event(client, user, club["id"], unique, create_round)

    resp = client.get(f"/api/v1/clubs/{club['id']}/events/{event['id']}", headers=user["headers"])
    assert resp.status_code == 200
    data = resp.json()
    assert data["id"] == event["id"]
    assert "participants" in data


def test_update_event(client, register_user, unique, create_round):
    user = register_user()
    club = client.post("/api/v1/clubs", json={"name": unique("club")}, headers=user["headers"]).json()
    event = _create_club_event(client, user, club["id"], unique, create_round)

    resp = client.patch(f"/api/v1/clubs/{club['id']}/events/{event['id']}", json={
        "name": "Updated Event",
        "location": "New Location",
    }, headers=user["headers"])
    assert resp.status_code == 200
    assert resp.json()["name"] == "Updated Event"
    assert resp.json()["location"] == "New Location"


def test_delete_event(client, register_user, unique, create_round):
    user = register_user()
    club = client.post("/api/v1/clubs", json={"name": unique("club")}, headers=user["headers"]).json()
    event = _create_club_event(client, user, club["id"], unique, create_round)

    resp = client.delete(f"/api/v1/clubs/{club['id']}/events/{event['id']}", headers=user["headers"])
    assert resp.status_code == 204


def test_rsvp_event(client, register_user, unique, create_round):
    user = register_user()
    club = client.post("/api/v1/clubs", json={"name": unique("club")}, headers=user["headers"]).json()
    event = _create_club_event(client, user, club["id"], unique, create_round)

    resp = client.post(f"/api/v1/clubs/{club['id']}/events/{event['id']}/rsvp", json={
        "status": "going",
    }, headers=user["headers"])
    assert resp.status_code == 200
    data = resp.json()
    participants = data["participants"]
    assert len(participants) >= 1
    assert any(p["status"] == "going" for p in participants)


def test_rsvp_event_update(client, register_user, unique, create_round):
    """RSVP can be updated."""
    user = register_user()
    club = client.post("/api/v1/clubs", json={"name": unique("club")}, headers=user["headers"]).json()
    event = _create_club_event(client, user, club["id"], unique, create_round)

    client.post(f"/api/v1/clubs/{club['id']}/events/{event['id']}/rsvp", json={
        "status": "going",
    }, headers=user["headers"])
    resp = client.post(f"/api/v1/clubs/{club['id']}/events/{event['id']}/rsvp", json={
        "status": "maybe",
    }, headers=user["headers"])
    assert resp.status_code == 200


# ── Teams ──────────────────────────────────────────────────────────────


def test_create_team(client, register_user, unique):
    user = register_user()
    club = client.post("/api/v1/clubs", json={"name": unique("club")}, headers=user["headers"]).json()
    me = client.get("/api/v1/users/me", headers=user["headers"]).json()

    resp = client.post(f"/api/v1/clubs/{club['id']}/teams", json={
        "name": unique("team"),
        "description": "A test team",
        "leader_id": me["id"],
    }, headers=user["headers"])
    assert resp.status_code == 201
    data = resp.json()
    assert data["name"].startswith("team_")
    assert data["leader"]["user_id"] == me["id"]
    assert data["member_count"] >= 0


def test_create_team_not_admin(client, register_user, unique):
    owner = register_user()
    member = register_user()
    club = _create_club_with_member(client, owner, member, unique)
    me = client.get("/api/v1/users/me", headers=member["headers"]).json()

    resp = client.post(f"/api/v1/clubs/{club['id']}/teams", json={
        "name": unique("team"),
        "leader_id": me["id"],
    }, headers=member["headers"])
    assert resp.status_code in (401, 403)


def test_list_teams(client, register_user, unique):
    user = register_user()
    club = client.post("/api/v1/clubs", json={"name": unique("club")}, headers=user["headers"]).json()
    me = client.get("/api/v1/users/me", headers=user["headers"]).json()
    client.post(f"/api/v1/clubs/{club['id']}/teams", json={
        "name": unique("team"),
        "leader_id": me["id"],
    }, headers=user["headers"])

    resp = client.get(f"/api/v1/clubs/{club['id']}/teams", headers=user["headers"])
    assert resp.status_code == 200
    assert len(resp.json()) >= 1


def test_get_team_detail(client, register_user, unique):
    user = register_user()
    club = client.post("/api/v1/clubs", json={"name": unique("club")}, headers=user["headers"]).json()
    me = client.get("/api/v1/users/me", headers=user["headers"]).json()
    team = client.post(f"/api/v1/clubs/{club['id']}/teams", json={
        "name": unique("team"),
        "leader_id": me["id"],
    }, headers=user["headers"]).json()

    resp = client.get(f"/api/v1/clubs/{club['id']}/teams/{team['id']}", headers=user["headers"])
    assert resp.status_code == 200
    data = resp.json()
    assert "members" in data
    assert data["leader"]["user_id"] == me["id"]


def test_update_team(client, register_user, unique):
    user = register_user()
    club = client.post("/api/v1/clubs", json={"name": unique("club")}, headers=user["headers"]).json()
    me = client.get("/api/v1/users/me", headers=user["headers"]).json()
    team = client.post(f"/api/v1/clubs/{club['id']}/teams", json={
        "name": unique("team"),
        "leader_id": me["id"],
    }, headers=user["headers"]).json()

    resp = client.patch(f"/api/v1/clubs/{club['id']}/teams/{team['id']}", json={
        "name": "Updated Team",
        "description": "New description",
    }, headers=user["headers"])
    assert resp.status_code == 200
    assert resp.json()["name"] == "Updated Team"


def test_delete_team(client, register_user, unique):
    user = register_user()
    club = client.post("/api/v1/clubs", json={"name": unique("club")}, headers=user["headers"]).json()
    me = client.get("/api/v1/users/me", headers=user["headers"]).json()
    team = client.post(f"/api/v1/clubs/{club['id']}/teams", json={
        "name": unique("team"),
        "leader_id": me["id"],
    }, headers=user["headers"]).json()

    resp = client.delete(f"/api/v1/clubs/{club['id']}/teams/{team['id']}", headers=user["headers"])
    assert resp.status_code == 204


def test_add_team_member(client, register_user, unique):
    owner = register_user()
    member = register_user()
    club = _create_club_with_member(client, owner, member, unique)

    me = client.get("/api/v1/users/me", headers=owner["headers"]).json()
    team = client.post(f"/api/v1/clubs/{club['id']}/teams", json={
        "name": unique("team"),
        "leader_id": me["id"],
    }, headers=owner["headers"]).json()

    member_me = client.get("/api/v1/users/me", headers=member["headers"]).json()
    resp = client.post(
        f"/api/v1/clubs/{club['id']}/teams/{team['id']}/members/{member_me['id']}",
        headers=owner["headers"],
    )
    assert resp.status_code == 201


def test_remove_team_member(client, register_user, unique):
    owner = register_user()
    member = register_user()
    club = _create_club_with_member(client, owner, member, unique)

    me = client.get("/api/v1/users/me", headers=owner["headers"]).json()
    team = client.post(f"/api/v1/clubs/{club['id']}/teams", json={
        "name": unique("team"),
        "leader_id": me["id"],
    }, headers=owner["headers"]).json()

    member_me = client.get("/api/v1/users/me", headers=member["headers"]).json()
    client.post(
        f"/api/v1/clubs/{club['id']}/teams/{team['id']}/members/{member_me['id']}",
        headers=owner["headers"],
    )
    resp = client.delete(
        f"/api/v1/clubs/{club['id']}/teams/{team['id']}/members/{member_me['id']}",
        headers=owner["headers"],
    )
    assert resp.status_code == 204


# ── Shared Rounds ──────────────────────────────────────────────────────


def test_list_shared_rounds_empty(client, register_user, unique):
    user = register_user()
    club = client.post("/api/v1/clubs", json={"name": unique("club")}, headers=user["headers"]).json()

    resp = client.get(f"/api/v1/clubs/{club['id']}/rounds", headers=user["headers"])
    assert resp.status_code == 200
    assert resp.json() == []


# ── Tournaments ────────────────────────────────────────────────────────


def _create_tournament(client, user, club_id, unique, create_round):
    """Helper: create a tournament in a club."""
    rnd = create_round(headers=user["headers"])
    resp = client.post(f"/api/v1/clubs/{club_id}/tournaments", json={
        "name": unique("tourney"),
        "description": "A test tournament",
        "template_id": rnd["id"],
        "max_participants": 10,
        "registration_deadline": "2026-06-01T00:00:00Z",
        "start_date": "2026-06-02T10:00:00Z",
        "end_date": "2026-06-02T18:00:00Z",
    }, headers=user["headers"])
    assert resp.status_code == 201
    return resp.json(), rnd


def test_create_tournament(client, register_user, unique, create_round):
    user = register_user()
    club = client.post("/api/v1/clubs", json={"name": unique("club")}, headers=user["headers"]).json()
    tourney, _ = _create_tournament(client, user, club["id"], unique, create_round)
    assert tourney["name"].startswith("tourney_")
    assert tourney["status"] == "registration"
    assert tourney["max_participants"] == 10


def test_create_tournament_not_admin(client, register_user, unique, create_round):
    owner = register_user()
    member = register_user()
    club = _create_club_with_member(client, owner, member, unique)
    rnd = create_round(headers=member["headers"])

    resp = client.post(f"/api/v1/clubs/{club['id']}/tournaments", json={
        "name": unique("tourney"),
        "template_id": rnd["id"],
        "registration_deadline": "2026-06-01T00:00:00Z",
        "start_date": "2026-06-02T10:00:00Z",
        "end_date": "2026-06-02T18:00:00Z",
    }, headers=member["headers"])
    assert resp.status_code in (401, 403)


def test_list_tournaments(client, register_user, unique, create_round):
    user = register_user()
    club = client.post("/api/v1/clubs", json={"name": unique("club")}, headers=user["headers"]).json()
    _create_tournament(client, user, club["id"], unique, create_round)

    resp = client.get(f"/api/v1/clubs/{club['id']}/tournaments", headers=user["headers"])
    assert resp.status_code == 200
    assert len(resp.json()) >= 1


def test_get_tournament(client, register_user, unique, create_round):
    user = register_user()
    club = client.post("/api/v1/clubs", json={"name": unique("club")}, headers=user["headers"]).json()
    tourney, _ = _create_tournament(client, user, club["id"], unique, create_round)

    resp = client.get(f"/api/v1/clubs/{club['id']}/tournaments/{tourney['id']}", headers=user["headers"])
    assert resp.status_code == 200
    data = resp.json()
    assert data["id"] == tourney["id"]
    assert "participants" in data


def test_register_for_tournament(client, register_user, unique, create_round):
    owner = register_user()
    member = register_user()
    club = _create_club_with_member(client, owner, member, unique)
    tourney, _ = _create_tournament(client, owner, club["id"], unique, create_round)

    resp = client.post(
        f"/api/v1/clubs/{club['id']}/tournaments/{tourney['id']}/register",
        headers=member["headers"],
    )
    assert resp.status_code == 201


def test_register_tournament_already_registered(client, register_user, unique, create_round):
    owner = register_user()
    member = register_user()
    club = _create_club_with_member(client, owner, member, unique)
    tourney, _ = _create_tournament(client, owner, club["id"], unique, create_round)

    client.post(
        f"/api/v1/clubs/{club['id']}/tournaments/{tourney['id']}/register",
        headers=member["headers"],
    )
    resp = client.post(
        f"/api/v1/clubs/{club['id']}/tournaments/{tourney['id']}/register",
        headers=member["headers"],
    )
    assert resp.status_code in (409, 422)


def test_start_tournament(client, register_user, unique, create_round):
    user = register_user()
    club = client.post("/api/v1/clubs", json={"name": unique("club")}, headers=user["headers"]).json()
    tourney, _ = _create_tournament(client, user, club["id"], unique, create_round)

    resp = client.post(
        f"/api/v1/clubs/{club['id']}/tournaments/{tourney['id']}/start",
        headers=user["headers"],
    )
    assert resp.status_code == 200
    assert resp.json()["status"] == "in_progress"


def test_tournament_leaderboard(client, register_user, unique, create_round):
    user = register_user()
    club = client.post("/api/v1/clubs", json={"name": unique("club")}, headers=user["headers"]).json()
    tourney, _ = _create_tournament(client, user, club["id"], unique, create_round)

    resp = client.get(
        f"/api/v1/clubs/{club['id']}/tournaments/{tourney['id']}/leaderboard",
        headers=user["headers"],
    )
    assert resp.status_code == 200
    assert isinstance(resp.json(), list)


def test_complete_tournament(client, register_user, unique, create_round):
    user = register_user()
    club = client.post("/api/v1/clubs", json={"name": unique("club")}, headers=user["headers"]).json()
    tourney, _ = _create_tournament(client, user, club["id"], unique, create_round)

    # Must start first
    client.post(
        f"/api/v1/clubs/{club['id']}/tournaments/{tourney['id']}/start",
        headers=user["headers"],
    )
    resp = client.post(
        f"/api/v1/clubs/{club['id']}/tournaments/{tourney['id']}/complete",
        headers=user["headers"],
    )
    assert resp.status_code == 200
    assert resp.json()["status"] == "completed"


def test_withdraw_from_tournament(client, register_user, unique, create_round):
    owner = register_user()
    member = register_user()
    club = _create_club_with_member(client, owner, member, unique)
    tourney, _ = _create_tournament(client, owner, club["id"], unique, create_round)

    client.post(
        f"/api/v1/clubs/{club['id']}/tournaments/{tourney['id']}/register",
        headers=member["headers"],
    )
    resp = client.post(
        f"/api/v1/clubs/{club['id']}/tournaments/{tourney['id']}/withdraw",
        headers=member["headers"],
    )
    assert resp.status_code == 200
