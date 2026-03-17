"""Smoke test — verifies the API is reachable."""


def test_health(client):
    resp = client.get("/health")
    assert resp.status_code == 200
