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


def _sample_stages():
    """Return a minimal valid stages array for round creation."""
    return [
        {
            "name": "Stage 1",
            "distance": "18m",
            "num_ends": 10,
            "arrows_per_end": 3,
            "allowed_values": ["X", "10", "9", "8", "7", "6", "5", "4", "3", "2", "1", "M"],
            "value_score_map": {
                "X": 10, "10": 10, "9": 9, "8": 8, "7": 7,
                "6": 6, "5": 5, "4": 4, "3": 3, "2": 2, "1": 1, "M": 0,
            },
            "max_score_per_arrow": 10,
        }
    ]


@pytest.fixture
def create_round(client, auth_headers, unique):
    """
    Factory fixture: create a custom round template and return its data.

    Cleanup deletes the round after the test.
    """
    created = []

    def _create(headers=None, **overrides):
        headers = headers or auth_headers
        payload = {
            "name": overrides.get("name", unique("round")),
            "organization": overrides.get("organization", "Test"),
            "description": overrides.get("description", "A test round"),
            "stages": overrides.get("stages", _sample_stages()),
        }
        resp = client.post("/api/v1/rounds", json=payload, headers=headers)
        assert resp.status_code == 201, f"Round creation failed: {resp.text}"
        data = resp.json()
        created.append((data["id"], headers))
        return data

    yield _create

    for round_id, headers in created:
        try:
            client.delete(f"/api/v1/rounds/{round_id}", headers=headers)
        except Exception:
            pass


@pytest.fixture
def create_equipment(client, auth_headers, unique):
    """
    Factory fixture: create an equipment item and return its data.

    Cleanup deletes the equipment after the test.
    """
    created = []

    def _create(headers=None, **overrides):
        headers = headers or auth_headers
        payload = {
            "category": overrides.get("category", "riser"),
            "name": overrides.get("name", unique("equip")),
            "brand": overrides.get("brand", "TestBrand"),
            "model": overrides.get("model", "TestModel"),
        }
        resp = client.post("/api/v1/equipment", json=payload, headers=headers)
        assert resp.status_code == 201, f"Equipment creation failed: {resp.text}"
        data = resp.json()
        created.append((data["id"], headers))
        return data

    yield _create

    for eq_id, headers in created:
        try:
            client.delete(f"/api/v1/equipment/{eq_id}", headers=headers)
        except Exception:
            pass


@pytest.fixture
def create_setup(client, auth_headers, unique):
    """
    Factory fixture: create a setup profile and return its data.

    Cleanup deletes the setup after the test.
    """
    created = []

    def _create(headers=None, **overrides):
        headers = headers or auth_headers
        payload = {
            "name": overrides.get("name", unique("setup")),
            "description": overrides.get("description", "A test setup"),
        }
        resp = client.post("/api/v1/setups", json=payload, headers=headers)
        assert resp.status_code == 201, f"Setup creation failed: {resp.text}"
        data = resp.json()
        created.append((data["id"], headers))
        return data

    yield _create

    for setup_id, headers in created:
        try:
            client.delete(f"/api/v1/setups/{setup_id}", headers=headers)
        except Exception:
            pass


@pytest.fixture
def create_sight_mark(client, auth_headers, unique):
    """
    Factory fixture: create a sight mark and return its data.

    Cleanup deletes the sight mark after the test.
    """
    created = []

    def _create(headers=None, **overrides):
        headers = headers or auth_headers
        payload = {
            "distance": overrides.get("distance", "18m"),
            "setting": overrides.get("setting", "3.5 turns"),
            "notes": overrides.get("notes", None),
            "date_recorded": overrides.get("date_recorded", "2025-06-01T12:00:00Z"),
            "equipment_id": overrides.get("equipment_id", None),
            "setup_id": overrides.get("setup_id", None),
        }
        resp = client.post("/api/v1/sight-marks", json=payload, headers=headers)
        assert resp.status_code == 201, f"Sight mark creation failed: {resp.text}"
        data = resp.json()
        created.append((data["id"], headers))
        return data

    yield _create

    for sm_id, headers in created:
        try:
            client.delete(f"/api/v1/sight-marks/{sm_id}", headers=headers)
        except Exception:
            pass


@pytest.fixture
def create_session(client, auth_headers, create_round):
    """
    Factory fixture: create a scoring session and return its data.

    Automatically creates a round template if none provided.
    Cleanup abandons then deletes the session after the test.
    """
    created = []

    def _create(headers=None, template_id=None, **overrides):
        headers = headers or auth_headers
        if template_id is None:
            rnd = create_round(headers=headers)
            template_id = rnd["id"]
        payload = {
            "template_id": template_id,
            "setup_profile_id": overrides.get("setup_profile_id"),
            "notes": overrides.get("notes"),
            "location": overrides.get("location"),
            "weather": overrides.get("weather"),
        }
        resp = client.post("/api/v1/sessions", json=payload, headers=headers)
        assert resp.status_code == 201, f"Session creation failed: {resp.text}"
        data = resp.json()
        created.append((data["id"], headers))
        return data

    yield _create

    for sid, headers in created:
        try:
            # Try to abandon first (only works if in_progress)
            client.post(f"/api/v1/sessions/{sid}/abandon", headers=headers)
            # Then delete (only works if abandoned)
            client.delete(f"/api/v1/sessions/{sid}", headers=headers)
        except Exception:
            pass
