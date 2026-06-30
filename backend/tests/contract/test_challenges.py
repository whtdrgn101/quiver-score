"""
Contract tests for Challenges endpoints.
"""

import os
import pytest

API_BASE_URL = os.environ.get("API_BASE_URL", "")
if not API_BASE_URL:
    pytest.skip("API_BASE_URL not set — skipping contract tests", allow_module_level=True)


def test_challenges_flow(client, register_user, unique, create_round):
    # 1. Register challenger and challengee
    challenger = register_user()
    challengee = register_user()

    # 2. Get a round template (create_round is a fixture that registers a template)
    rnd = create_round(name=unique("template"))

    # 3. Create a challenge
    resp = client.post(
        "/api/v1/challenges",
        json={
            "challengee_id": challengee["id"],
            "template_id": rnd["id"],
            "expires_in_hours": 24,
        },
        headers=challenger["headers"],
    )
    assert resp.status_code == 201
    challenge = resp.json()
    assert challenge["id"] is not None
    assert challenge["status"] == "pending"
    assert challenge["challenger_username"] == challenger["username"]
    assert challenge["challengee_username"] == challengee["username"]

    # 4. List challenges
    resp = client.get("/api/v1/challenges", headers=challenger["headers"])
    assert resp.status_code == 200
    challenges = resp.json()
    assert len(challenges) >= 1
    assert any(c["id"] == challenge["id"] for c in challenges)

    # 5. Accept challenge
    resp = client.post(
        f"/api/v1/challenges/{challenge['id']}/accept",
        headers=challengee["headers"],
    )
    assert resp.status_code == 200
    challenge = resp.json()
    assert challenge["status"] == "accepted"

    # 6. Decline verification (cannot decline accepted challenge or accept again)
    resp = client.post(
        f"/api/v1/challenges/{challenge['id']}/decline",
        headers=challengee["headers"],
    )
    assert resp.status_code == 400
