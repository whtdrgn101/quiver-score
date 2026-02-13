import uuid

import pytest

from app.seed.round_templates import seed_round_templates


# ── Helpers ──────────────────────────────────────────────────────────────


async def _register(client, email, username):
    resp = await client.post("/api/v1/auth/register", json={
        "email": email, "username": username, "password": "pass1234",
    })
    token = resp.json()["access_token"]
    return token, {"Authorization": f"Bearer {token}"}


async def _create_club(client, headers, name="Test Club"):
    resp = await client.post("/api/v1/clubs", json={"name": name, "description": "A test club"}, headers=headers)
    assert resp.status_code == 201
    return resp.json()


async def _create_invite(client, headers, club_id):
    resp = await client.post(f"/api/v1/clubs/{club_id}/invites", json={}, headers=headers)
    assert resp.status_code == 201
    return resp.json()


async def _join_club(client, headers, code):
    resp = await client.post(f"/api/v1/clubs/join/{code}", headers=headers)
    assert resp.status_code == 200
    return resp.json()


async def _create_custom_round(client, headers, name="My Custom Round"):
    resp = await client.post("/api/v1/rounds", json={
        "name": name,
        "organization": "Custom",
        "description": "A custom round",
        "stages": [{
            "name": "Stage 1",
            "distance": "20m",
            "num_ends": 6,
            "arrows_per_end": 3,
            "allowed_values": ["X", "10", "9", "8", "7", "6", "5", "4", "3", "2", "1", "M"],
            "value_score_map": {"X": 10, "10": 10, "9": 9, "8": 8, "7": 7, "6": 6, "5": 5, "4": 4, "3": 3, "2": 2, "1": 1, "M": 0},
            "max_score_per_arrow": 10,
        }],
    }, headers=headers)
    assert resp.status_code == 201
    return resp.json()


# ── Tests ────────────────────────────────────────────────────────────────


@pytest.mark.asyncio
async def test_share_round_with_club(client):
    _, owner_h = await _register(client, "share1@test.com", "shareowner1")
    club = await _create_club(client, owner_h)
    custom = await _create_custom_round(client, owner_h)

    resp = await client.post(
        f"/api/v1/rounds/{custom['id']}/share",
        json={"club_id": club["id"]},
        headers=owner_h,
    )
    assert resp.status_code == 201

    # Verify it appears in club rounds
    resp = await client.get(f"/api/v1/clubs/{club['id']}/rounds", headers=owner_h)
    assert resp.status_code == 200
    rounds = resp.json()
    assert len(rounds) == 1
    assert rounds[0]["template_name"] == "My Custom Round"


@pytest.mark.asyncio
async def test_shared_round_appears_in_round_list(client):
    _, owner_h = await _register(client, "share2@test.com", "shareowner2")
    _, member_h = await _register(client, "member2@test.com", "sharemember2")

    club = await _create_club(client, owner_h)
    invite = await _create_invite(client, owner_h, club["id"])
    await _join_club(client, member_h, invite["code"])

    custom = await _create_custom_round(client, owner_h, "Shared Field Round")
    await client.post(
        f"/api/v1/rounds/{custom['id']}/share",
        json={"club_id": club["id"]},
        headers=owner_h,
    )

    # Member should see the shared round in their round list
    resp = await client.get("/api/v1/rounds", headers=member_h)
    assert resp.status_code == 200
    names = [r["name"] for r in resp.json()]
    assert "Shared Field Round" in names


@pytest.mark.asyncio
async def test_share_duplicate_conflict(client):
    _, owner_h = await _register(client, "share3@test.com", "shareowner3")
    club = await _create_club(client, owner_h)
    custom = await _create_custom_round(client, owner_h)

    await client.post(
        f"/api/v1/rounds/{custom['id']}/share",
        json={"club_id": club["id"]},
        headers=owner_h,
    )
    resp = await client.post(
        f"/api/v1/rounds/{custom['id']}/share",
        json={"club_id": club["id"]},
        headers=owner_h,
    )
    assert resp.status_code == 409


@pytest.mark.asyncio
async def test_share_official_round_forbidden(client, db_session):
    await seed_round_templates(db_session)
    await db_session.commit()

    _, owner_h = await _register(client, "share4@test.com", "shareowner4")
    club = await _create_club(client, owner_h)

    # Get an official round
    resp = await client.get("/api/v1/rounds", headers=owner_h)
    official = next(r for r in resp.json() if r["is_official"])

    resp = await client.post(
        f"/api/v1/rounds/{official['id']}/share",
        json={"club_id": club["id"]},
        headers=owner_h,
    )
    assert resp.status_code == 403


@pytest.mark.asyncio
async def test_share_not_creator_forbidden(client):
    _, creator_h = await _register(client, "share5@test.com", "sharecreator5")
    _, other_h = await _register(client, "other5@test.com", "shareother5")

    club = await _create_club(client, other_h)
    custom = await _create_custom_round(client, creator_h)

    # other user tries to share creator's round
    resp = await client.post(
        f"/api/v1/rounds/{custom['id']}/share",
        json={"club_id": club["id"]},
        headers=other_h,
    )
    assert resp.status_code == 403


@pytest.mark.asyncio
async def test_share_not_club_member(client):
    _, owner_h = await _register(client, "share6@test.com", "shareowner6")
    _, outsider_h = await _register(client, "outsider6@test.com", "shareoutsider6")

    club = await _create_club(client, owner_h)
    custom = await _create_custom_round(client, outsider_h)

    resp = await client.post(
        f"/api/v1/rounds/{custom['id']}/share",
        json={"club_id": club["id"]},
        headers=outsider_h,
    )
    assert resp.status_code == 401


@pytest.mark.asyncio
async def test_creator_unshare(client):
    _, owner_h = await _register(client, "share7@test.com", "shareowner7")
    club = await _create_club(client, owner_h)
    custom = await _create_custom_round(client, owner_h)

    await client.post(
        f"/api/v1/rounds/{custom['id']}/share",
        json={"club_id": club["id"]},
        headers=owner_h,
    )
    resp = await client.delete(
        f"/api/v1/rounds/{custom['id']}/share/{club['id']}",
        headers=owner_h,
    )
    assert resp.status_code == 204

    # Verify it's gone
    resp = await client.get(f"/api/v1/clubs/{club['id']}/rounds", headers=owner_h)
    assert resp.status_code == 200
    assert len(resp.json()) == 0


@pytest.mark.asyncio
async def test_club_admin_remove_shared_round(client):
    _, owner_h = await _register(client, "share8@test.com", "shareowner8")
    _, admin_h = await _register(client, "admin8@test.com", "shareadmin8")

    club = await _create_club(client, owner_h)
    invite = await _create_invite(client, owner_h, club["id"])
    await _join_club(client, admin_h, invite["code"])

    # Promote to admin
    resp = await client.get("/api/v1/users/me", headers=admin_h)
    admin_id = resp.json()["id"]
    await client.post(f"/api/v1/clubs/{club['id']}/members/{admin_id}/promote", headers=owner_h)

    custom = await _create_custom_round(client, owner_h)
    await client.post(
        f"/api/v1/rounds/{custom['id']}/share",
        json={"club_id": club["id"]},
        headers=owner_h,
    )

    # Admin removes the shared round
    resp = await client.delete(
        f"/api/v1/clubs/{club['id']}/rounds/{custom['id']}",
        headers=admin_h,
    )
    assert resp.status_code == 204


@pytest.mark.asyncio
async def test_regular_member_cannot_remove(client):
    _, owner_h = await _register(client, "share9@test.com", "shareowner9")
    _, member_h = await _register(client, "member9@test.com", "sharemember9")

    club = await _create_club(client, owner_h)
    invite = await _create_invite(client, owner_h, club["id"])
    await _join_club(client, member_h, invite["code"])

    custom = await _create_custom_round(client, owner_h)
    await client.post(
        f"/api/v1/rounds/{custom['id']}/share",
        json={"club_id": club["id"]},
        headers=owner_h,
    )

    # Regular member tries to remove
    resp = await client.delete(
        f"/api/v1/clubs/{club['id']}/rounds/{custom['id']}",
        headers=member_h,
    )
    assert resp.status_code == 401


@pytest.mark.asyncio
async def test_list_club_shared_rounds(client):
    _, owner_h = await _register(client, "share10@test.com", "shareowner10")
    club = await _create_club(client, owner_h, "Archery Club")
    custom1 = await _create_custom_round(client, owner_h, "Field Animal")
    custom2 = await _create_custom_round(client, owner_h, "3D Safari")

    await client.post(f"/api/v1/rounds/{custom1['id']}/share", json={"club_id": club["id"]}, headers=owner_h)
    await client.post(f"/api/v1/rounds/{custom2['id']}/share", json={"club_id": club["id"]}, headers=owner_h)

    resp = await client.get(f"/api/v1/clubs/{club['id']}/rounds", headers=owner_h)
    assert resp.status_code == 200
    data = resp.json()
    assert len(data) == 2
    names = {r["template_name"] for r in data}
    assert names == {"Field Animal", "3D Safari"}
    # Check metadata
    assert data[0]["club_name"] == "Archery Club"
    assert data[0]["shared_by_username"] == "shareowner10"
    assert "shared_at" in data[0]


@pytest.mark.asyncio
async def test_non_member_cannot_list(client):
    _, owner_h = await _register(client, "share11@test.com", "shareowner11")
    _, outsider_h = await _register(client, "outsider11@test.com", "shareoutsider11")

    club = await _create_club(client, owner_h)

    resp = await client.get(f"/api/v1/clubs/{club['id']}/rounds", headers=outsider_h)
    assert resp.status_code == 401
