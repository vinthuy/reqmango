"""Tests for issues client methods."""

import pytest
import respx

from reqmango import IssueListResult


def test_list_issues_reads_total_header(mock_api, client):
    mock_api.get("/issues").respond(
        200,
        json=[{"id": 1, "name": "bug", "sequence_id": 5}],
        headers={"X-Total-Count": "7"},
    )
    result = client.list_issues(project_id=1, rql='priority = "high"')
    assert isinstance(result, IssueListResult)
    assert result.total == 7
    assert len(result.items) == 1
    assert result.items[0].name == "bug"


def test_create_issue_with_optional_fields(mock_api, client):
    def _check_request(request):
        import json
        body = json.loads(request.content)
        assert body["name"] == "Login broken"
        assert body["priority"] == "high"
        assert body["state_id"] == 3
        return httpx.Response(201, json={"id": 11, "name": "Login broken", "sequence_id": 42, "priority": "high"})

    mock_api.post("/issues").mock(side_effect=_check_request)
    import httpx
    issue = client.create_issue(project_id=5, workspace_id=2, name="Login broken", priority="high", state_id=3)
    assert issue.id == 11
    assert issue.sequence_id == 42


def test_get_issue(mock_api, client):
    mock_api.get("/issues/11").respond(200, json={"id": 11, "name": "Login broken", "sequence_id": 42})
    issue = client.get_issue(11)
    assert issue.id == 11
    assert issue.name == "Login broken"


def test_search_issues(mock_api, client):
    mock_api.get("/issues/search").respond(200, json=[
        {"id": 1, "name": "auth bug", "sequence_id": 5, "project_identifier": "DEMO", "project_id": 1},
    ])
    results = client.search_issues(workspace_id=2, query="auth")
    assert len(results) == 1
    assert results[0].project_identifier == "DEMO"


def test_add_comment(mock_api, client):
    mock_api.post("/comments").respond(200, json={
        "id": 1, "issue_id": 11, "body": "looks good", "author_id": 1,
    })
    c = client.add_comment(issue_id=11, body="looks good")
    assert c.id == 1
    assert c.body == "looks good"


def test_list_comments(mock_api, client):
    mock_api.get("/comments/issue/11").respond(200, json={
        "comments": [{"id": 1, "body": "first"}, {"id": 2, "body": "second"}],
        "total": 2,
    })
    comments, total = client.list_comments(issue_id=11)
    assert total == 2
    assert len(comments) == 2
    assert comments[0].body == "first"
