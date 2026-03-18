"""
Contract tests for /api/v1/users/me/classifications endpoints.

These validate status codes and response shapes against a live server.

Note: Classification records are created automatically by the scoring engine
when a session is completed. There is no manual CRUD — these tests verify
the read-only endpoints return correct shapes and enforce auth.
"""


# ── List All Classifications ──────────────────────────────────────────


def test_list_classifications_empty(client, auth_headers):
    """GET /api/v1/users/me/classifications for a fresh user returns empty list."""
    resp = client.get("/api/v1/users/me/classifications", headers=auth_headers)
    assert resp.status_code == 200
    assert resp.json() == []


def test_list_classifications_unauthenticated(client):
    """GET /api/v1/users/me/classifications without auth returns 401."""
    resp = client.get("/api/v1/users/me/classifications")
    assert resp.status_code == 401


# ── Current Classifications ───────────────────────────────────────────


def test_current_classifications_empty(client, auth_headers):
    """GET /api/v1/users/me/classifications/current for a fresh user returns empty list."""
    resp = client.get("/api/v1/users/me/classifications/current", headers=auth_headers)
    assert resp.status_code == 200
    assert resp.json() == []


def test_current_classifications_unauthenticated(client):
    """GET /api/v1/users/me/classifications/current without auth returns 401."""
    resp = client.get("/api/v1/users/me/classifications/current")
    assert resp.status_code == 401
