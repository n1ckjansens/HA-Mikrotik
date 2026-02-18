#!/usr/bin/env python3
"""Mock Home Assistant internal API for add-on local development."""

from __future__ import annotations

import json
import os
from dataclasses import dataclass, field
from datetime import datetime, timezone
from http import HTTPStatus
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from threading import Lock
from typing import Any


@dataclass
class ConfigState:
    version: int = 1
    host: str = os.getenv("ROUTER_HOST", "mock-routeros:18080")
    username: str = os.getenv("ROUTER_USERNAME", "admin")
    password: str = os.getenv("ROUTER_PASSWORD", "admin")
    ssl: bool = os.getenv("ROUTER_SSL", "false").lower() == "true"
    verify_tls: bool = os.getenv("ROUTER_VERIFY_TLS", "false").lower() == "true"
    poll_interval_sec: int = int(os.getenv("ROUTER_POLL_INTERVAL_SEC", "5"))
    roles: list[str] = field(default_factory=list)


LOCK = Lock()
STATE = ConfigState()


class Handler(BaseHTTPRequestHandler):
    server_version = "MockHA/1.0"

    def _json(self, status: HTTPStatus, payload: Any) -> None:
        body = json.dumps(payload).encode("utf-8")
        self.send_response(status)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def do_GET(self) -> None:  # noqa: N802
        if self.path == "/healthz":
            self._json(HTTPStatus.OK, {"status": "ok"})
            return

        if self.path != "/api/mikrotik_presence/config":
            self._json(HTTPStatus.NOT_FOUND, {"error": "not_found"})
            return

        with LOCK:
            payload = {
                "configured": True,
                "version": STATE.version,
                "updated_at": datetime.now(timezone.utc).isoformat(),
                "host": STATE.host,
                "username": STATE.username,
                "password": STATE.password,
                "ssl": STATE.ssl,
                "verify_tls": STATE.verify_tls,
                "poll_interval_sec": max(5, STATE.poll_interval_sec),
                "roles": STATE.roles,
            }

        self._json(HTTPStatus.OK, payload)

    def do_POST(self) -> None:  # noqa: N802
        if self.path != "/admin/config":
            self._json(HTTPStatus.NOT_FOUND, {"error": "not_found"})
            return

        size = int(self.headers.get("Content-Length", "0"))
        raw = self.rfile.read(size) if size > 0 else b"{}"
        try:
            patch = json.loads(raw.decode("utf-8"))
        except json.JSONDecodeError:
            self._json(HTTPStatus.BAD_REQUEST, {"error": "invalid_json"})
            return

        with LOCK:
            for key, value in patch.items():
                if hasattr(STATE, key):
                    setattr(STATE, key, value)
            STATE.version += 1
            version = STATE.version

        self._json(HTTPStatus.OK, {"ok": True, "version": version})

    def log_message(self, fmt: str, *args: object) -> None:
        print(f"[mock-ha] {self.address_string()} - {fmt % args}")


def main() -> None:
    bind = os.getenv("MOCK_HA_BIND", "0.0.0.0")
    port = int(os.getenv("MOCK_HA_PORT", "8123"))
    server = ThreadingHTTPServer((bind, port), Handler)
    print(f"Mock HA listening on http://{bind}:{port}")
    print(
        "Config source:",
        {
            "host": STATE.host,
            "ssl": STATE.ssl,
            "verify_tls": STATE.verify_tls,
            "poll_interval_sec": STATE.poll_interval_sec,
        },
    )
    server.serve_forever()


if __name__ == "__main__":
    main()
