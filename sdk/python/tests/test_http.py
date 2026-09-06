"""Tests for the HTTP client error mapping."""

import pytest
import respx
import httpx

from reqmango import APIError, ReqMangoClient


def test_get_json_ok(mock_api, client):
    mock_api.get("/workspaces").respond(200, json=[{"id": 1, "name": "Acme", "slug": "acme"}])
    ws = client.list_workspaces()
    assert len(ws) == 1
    assert ws[0].name == "Acme"


def test_api_error_401(mock_api, client):
    mock_api.get("/workspaces").respond(401, json={"message": "token expired"})
    with pytest.raises(APIError) as exc_info:
        client.list_workspaces()
    assert exc_info.value.status_code == 401
    assert "token expired" in exc_info.value.message


def test_api_error_409_with_body(mock_api, client):
    mock_api.put("/issues/1").respond(409, json={
        "message": "approval_required",
        "transition_id": 9,
    })
    with pytest.raises(APIError) as exc_info:
        client.update_issue(1, name="new")
    assert exc_info.value.status_code == 409
    assert exc_info.value.body.get("transition_id") == 9


def test_post_json_created(mock_api, client):
    mock_api.post("/issues").respond(201, json={"id": 11, "name": "Bug", "sequence_id": 42})
    issue = client.create_issue(project_id=5, workspace_id=2, name="Bug")
    assert issue.id == 11
    assert issue.sequence_id == 42


def test_delete_json(mock_api, client):
    mock_api.delete("/auth/tokens/3").respond(200, json={"message": "revoked"})
    client.revoke_pat(3)  # should not raise
