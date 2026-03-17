"""
Contract test fixtures.

These tests run against a live API server (Python or Go) specified by
the API_BASE_URL environment variable. They validate HTTP status codes
and response shapes to ensure API compatibility across implementations.

Usage:
    # Against local Python API
    API_BASE_URL=http://localhost:8000 pytest tests/contract/ -v

    # Against local Go API
    API_BASE_URL=http://localhost:8080 pytest tests/contract/ -v

    # Against production
    API_BASE_URL=https://api.quiverscore.com pytest tests/contract/ -v
"""

import os
import uuid

import httpx
import pytest


API_BASE_URL = os.environ.get("API_BASE_URL", "")

# Skip all contract tests if no API_BASE_URL is set
if not API_BASE_URL:
    pytest.skip("API_BASE_URL not set — skipping contract tests", allow_module_level=True)


@pytest.fixture(scope="session")
def base_url():
    return API_BASE_URL.rstrip("/")


@pytest.fixture(scope="session")
def client(base_url):
    """Synchronous httpx client pointed at the live API."""
    with httpx.Client(base_url=base_url, timeout=30.0) as c:
        yield c


def _unique(prefix: str) -> str:
    """Generate a unique string to avoid collisions between test runs."""
    return f"{prefix}_{uuid.uuid4().hex[:8]}"


@pytest.fixture
def unique():
    """Returns a helper to generate unique test identifiers."""
    return _unique


@pytest.fixture
def register_user(client, unique):
    """
    Factory fixture: register a fresh user and return (tokens_dict, headers).

    The user is unique per call so tests don't interfere with each other.
    Cleanup is best-effort via delete-account.
    """
    created = []

    def _register(email=None, username=None, password="testpass1234"):
        email = email or f"{unique('u')}@test.com"
        username = username or unique("user")
        resp = client.post("/api/v1/auth/register", json={
            "email": email,
            "username": username,
            "password": password,
        })
        assert resp.status_code == 201, f"Registration failed: {resp.text}"
        tokens = resp.json()
        headers = {"Authorization": f"Bearer {tokens['access_token']}"}
        created.append(headers)
        return {
            "email": email,
            "username": username,
            "password": password,
            "access_token": tokens["access_token"],
            "refresh_token": tokens["refresh_token"],
            "headers": headers,
        }

    yield _register

    # Best-effort cleanup: delete all users we created
    for headers in created:
        try:
            client.post("/api/v1/auth/delete-account", json={
                "confirmation": "Yes, delete ALL of my data",
            }, headers=headers)
        except Exception:
            pass


@pytest.fixture
def auth_headers(register_user):
    """Convenience: register one user and return just the auth headers."""
    user = register_user()
    return user["headers"]
