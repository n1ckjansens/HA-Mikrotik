#!/usr/bin/env python3
"""Mock RouterOS v7 REST API for local development."""

from __future__ import annotations

import json
import os
from http import HTTPStatus
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from threading import Lock
from urllib.parse import parse_qs, urlparse

LOCK = Lock()
STATE = {
    "scenario": os.getenv("MOCK_ROUTEROS_SCENARIO", "online"),
}

ONLINE_PAYLOADS = {
    "/rest/ip/dhcp-server/lease": [
        {
            ".id": "*1",
            "mac-address": "AA:BB:CC:DD:EE:01",
            "address": "192.168.88.10",
            "host-name": "iphone-anna",
            "status": "bound",
            "last-seen": "5s",
        },
        {
            ".id": "*2",
            "mac-address": "AA:BB:CC:DD:EE:02",
            "address": "192.168.88.20",
            "host-name": "laptop-max",
            "status": "bound",
            "last-seen": "18s",
        },
    ],
    "/rest/interface/wifi/registration-table": [
        {
            ".id": "*3",
            "mac-address": "AA:BB:CC:DD:EE:01",
            "interface": "wifi1",
            "uptime": "3h2m10s",
            "last-activity": "4s",
        }
    ],
    "/rest/interface/bridge/host": [
        {
            ".id": "*4",
            "mac-address": "AA:BB:CC:DD:EE:01",
            "on-interface": "bridge-lan",
        },
        {
            ".id": "*5",
            "mac-address": "AA:BB:CC:DD:EE:02",
            "on-interface": "bridge-lan",
        },
    ],
    "/rest/ip/arp": [
        {
            ".id": "*6",
            "mac-address": "AA:BB:CC:DD:EE:01",
            "address": "192.168.88.10",
            "interface": "bridge-lan",
        },
        {
            ".id": "*7",
            "mac-address": "AA:BB:CC:DD:EE:02",
            "address": "192.168.88.20",
            "interface": "bridge-lan",
        },
    ],
    "/rest/ip/address": [
        {
            ".id": "*8",
            "address": "192.168.88.1/24",
            "interface": "bridge-lan",
        },
        {
            ".id": "*9",
            "address": "10.10.0.1/16",
            "interface": "vlan-iot",
        },
    ],
}

OFFLINE_PAYLOADS = {
    "/rest/ip/dhcp-server/lease": [],
    "/rest/interface/wifi/registration-table": [],
    "/rest/interface/bridge/host": [],
    "/rest/ip/arp": [],
    "/rest/ip/address": ONLINE_PAYLOADS["/rest/ip/address"],
}


class Handler(BaseHTTPRequestHandler):
    server_version = "MockRouterOS/1.0"

    def _json(self, status: HTTPStatus, payload: object) -> None:
        body = json.dumps(payload).encode("utf-8")
        self.send_response(status)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def do_GET(self) -> None:  # noqa: N802
        parsed = urlparse(self.path)

        if parsed.path == "/healthz":
            with LOCK:
                scenario = STATE["scenario"]
            self._json(HTTPStatus.OK, {"status": "ok", "scenario": scenario})
            return

        if parsed.path == "/admin/scenario":
            query = parse_qs(parsed.query)
            desired = query.get("state", [None])[0]
            with LOCK:
                if desired in {"online", "offline"}:
                    STATE["scenario"] = desired
                scenario = STATE["scenario"]
            self._json(HTTPStatus.OK, {"scenario": scenario})
            return

        with LOCK:
            scenario = STATE["scenario"]

        data_source = ONLINE_PAYLOADS if scenario == "online" else OFFLINE_PAYLOADS
        if parsed.path not in data_source:
            self._json(HTTPStatus.NOT_FOUND, {"error": "not_found", "path": parsed.path})
            return

        self._json(HTTPStatus.OK, data_source[parsed.path])

    def log_message(self, fmt: str, *args: object) -> None:
        print(f"[mock-routeros] {self.address_string()} - {fmt % args}")


def main() -> None:
    bind = os.getenv("MOCK_ROUTEROS_BIND", "0.0.0.0")
    port = int(os.getenv("MOCK_ROUTEROS_PORT", "18080"))
    server = ThreadingHTTPServer((bind, port), Handler)
    print(f"Mock RouterOS listening on http://{bind}:{port}")
    server.serve_forever()


if __name__ == "__main__":
    main()
