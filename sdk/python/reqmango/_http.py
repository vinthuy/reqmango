"""Core HTTP client with typed error handling."""

from __future__ import annotations

from typing import Any

import httpx

DEFAULT_BASE_URL = "http://localhost:8000/api/v1"
DEFAULT_TIMEOUT = 30.0
LONG_TIMEOUT = 300.0  # 5 min for AI/agent ops


class APIError(Exception):
    """Typed error carrying the backend's {"message": ...} body."""

    def __init__(self, status_code: int, message: str, body: dict[str, Any] | None = None):
        self.status_code = status_code
        self.message = message
        self.body = body or {}
        super().__init__(f"api error {status_code}: {message}")


class HTTPClient:
    """Low-level HTTP client shared by all resource mixins."""

    def __init__(self, base_url: str = "", token: str = "", timeout: float = DEFAULT_TIMEOUT):
        self._base_url = (base_url or DEFAULT_BASE_URL).rstrip("/")
        self._token = token
        self._client = httpx.Client(
            timeout=timeout,
            headers={"Authorization": f"Bearer {token}"},
        )

    @property
    def base_url(self) -> str:
        return self._base_url

    def _do(
        self,
        method: str,
        path: str,
        *,
        query: dict[str, Any] | None = None,
        body: Any = None,
        out_type: type | None = None,
        timeout: float | None = None,
    ) -> Any:
        url = f"{self._base_url}{path}"
        kwargs: dict[str, Any] = {}
        if query:
            kwargs["params"] = {k: v for k, v in query.items() if v is not None and v != "" and v != 0}
        if body is not None:
            kwargs["json"] = body
        if timeout:
            kwargs["timeout"] = timeout

        resp = self._client.request(method, url, **kwargs)

        if resp.status_code < 200 or resp.status_code >= 300:
            body_data: dict[str, Any] = {}
            try:
                body_data = resp.json()
            except Exception:
                pass
            msg = body_data.get("message", resp.text.strip())
            raise APIError(resp.status_code, msg, body_data)

        if resp.status_code == 204 or not resp.content:
            return None
        data = resp.json()
        return data

    def get_json(self, path: str, query: dict[str, Any] | None = None, *, timeout: float | None = None) -> Any:
        return self._do("GET", path, query=query, timeout=timeout)

    def post_json(self, path: str, body: Any = None, query: dict[str, Any] | None = None, *, timeout: float | None = None) -> Any:
        return self._do("POST", path, query=query, body=body, timeout=timeout)

    def put_json(self, path: str, body: Any = None, query: dict[str, Any] | None = None, *, timeout: float | None = None) -> Any:
        return self._do("PUT", path, query=query, body=body, timeout=timeout)

    def delete_json(self, path: str, query: dict[str, Any] | None = None, *, timeout: float | None = None) -> Any:
        return self._do("DELETE", path, query=query, timeout=timeout)

    def close(self) -> None:
        self._client.close()

    def __enter__(self) -> HTTPClient:
        return self

    def __exit__(self, *args: Any) -> None:
        self.close()
