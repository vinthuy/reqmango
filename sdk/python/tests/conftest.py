"""Shared test fixtures."""

import pytest
import respx
import httpx

from reqmango import ReqMangoClient


@pytest.fixture
def mock_api():
    """respx mock targeting the default base URL."""
    with respx.mock(base_url="http://localhost:8000/api/v1") as rsps:
        yield rsps


@pytest.fixture
def client():
    """ReqMangoClient with a test token."""
    return ReqMangoClient(token="reqmango_pat_test")
