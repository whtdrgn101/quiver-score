import pytest

from app.services.classification import calculate_classification

_user_counter = 0


async def _register_and_get_token(client, email=None, username=None):
    global _user_counter
    _user_counter += 1
    email = email or f"class{_user_counter}@test.com"
    username = username or f"class{_user_counter}"
    resp = await client.post("/api/v1/auth/register", json={
        "email": email, "username": username, "password": "pass1234",
    })
    return resp.json()["access_token"]


def test_calculate_classification_known_round():
    """calculate_classification returns correct classification for known round."""
    result = calculate_classification(580, "WA 720 (70m)")
    assert result is not None
    system, classification = result
    assert system == "ArcheryGB"
    assert classification == "Master Bowman"


def test_calculate_classification_unknown_round():
    """calculate_classification returns None for unknown round types."""
    result = calculate_classification(500, "Unknown Round")
    assert result is None


@pytest.mark.asyncio
async def test_list_classifications_empty(client, db_session):
    """GET /users/me/classifications returns empty list initially."""
    token = await _register_and_get_token(client)
    headers = {"Authorization": f"Bearer {token}"}
    resp = await client.get("/api/v1/users/me/classifications", headers=headers)
    assert resp.status_code == 200
    assert resp.json() == []


@pytest.mark.asyncio
async def test_current_classifications_empty(client, db_session):
    """GET /users/me/classifications/current returns empty list initially."""
    token = await _register_and_get_token(client)
    headers = {"Authorization": f"Bearer {token}"}
    resp = await client.get("/api/v1/users/me/classifications/current", headers=headers)
    assert resp.status_code == 200
    assert resp.json() == []
