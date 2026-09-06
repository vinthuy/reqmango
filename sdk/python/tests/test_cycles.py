"""Tests for cycles client methods."""

import pytest

from reqmango import CycleListResult, CycleProgress, BurndownData


def test_list_cycles_wrapped_shape(mock_api, client):
    mock_api.get("/projects/5/cycles").respond(200, json={
        "items": [{"id": 1, "name": "Sprint 1", "status": "active", "progress": 40.5}],
        "total": 1, "limit": 50, "offset": 0,
    })
    result = client.list_cycles(project_id=5, status="active")
    assert isinstance(result, CycleListResult)
    assert result.total == 1
    assert len(result.items) == 1
    assert result.items[0].progress == 40.5


def test_get_cycle(mock_api, client):
    mock_api.get("/cycles/3").respond(200, json={"id": 3, "name": "Sprint 1", "status": "active"})
    cycle = client.get_cycle(3)
    assert cycle.id == 3
    assert cycle.name == "Sprint 1"


def test_get_cycle_progress(mock_api, client):
    mock_api.get("/cycles/3/progress").respond(200, json={
        "cycle_id": 3, "cycle_name": "S1", "total_issues": 10,
        "completed_issues": 5, "progress": 50.0,
        "state_breakdown": [{"state": "Done", "group": "completed", "count": 5}],
    })
    p = client.get_cycle_progress(3)
    assert isinstance(p, CycleProgress)
    assert p.progress == 50.0
    assert len(p.state_breakdown) == 1


def test_get_cycle_burndown(mock_api, client):
    mock_api.get("/cycles/3/burndown").respond(200, json={
        "cycle_id": 3, "cycle_name": "S1", "total_issues": 10, "is_on_track": True,
        "daily_points": [{"day_index": 0, "actual_remaining": 9.0}],
    })
    b = client.get_cycle_burndown(3)
    assert isinstance(b, BurndownData)
    assert b.is_on_track is True
    assert len(b.daily_points) == 1


def test_add_issue_to_cycle(mock_api, client):
    mock_api.post("/cycles/3/issues").respond(200, json={"cycle_id": 3, "issue_id": 11, "action": "added"})
    client.add_issue_to_cycle(cycle_id=3, issue_id=11)  # should not raise
