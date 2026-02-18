#!/usr/bin/env bash
set -euo pipefail

backend_url="${BACKEND_URL:-http://127.0.0.1:8099}"
mock_router_url="${MOCK_ROUTER_URL:-http://127.0.0.1:18080}"

echo "[smoke] Health check"
curl -fsS "${backend_url}/healthz" | tee /tmp/mikrotik-health.json

echo "[smoke] Force refresh"
curl -fsS -X POST "${backend_url}/api/refresh" -H 'Content-Type: application/json' -d '{}' >/dev/null
sleep 2

echo "[smoke] Load devices"
DEVICES_JSON="$(curl -fsS "${backend_url}/api/devices")"
printf '%s' "$DEVICES_JSON" > /tmp/mikrotik-devices.json
python3 - <<'PY'
import json
from pathlib import Path
payload = json.loads(Path('/tmp/mikrotik-devices.json').read_text())
items = payload.get('items', [])
print(f"devices_count={len(items)}")
if not items:
    raise SystemExit('no devices returned')
print(f"first_device={items[0]['mac']} status={items[0]['status']} online={items[0]['online']}")
PY

MAC="$(python3 - <<'PY'
import json
from pathlib import Path
payload = json.loads(Path('/tmp/mikrotik-devices.json').read_text())
print(payload['items'][0]['mac'])
PY
)"

echo "[smoke] Register first device ${MAC}"
curl -fsS -X POST "${backend_url}/api/devices/${MAC}/register" \
  -H 'Content-Type: application/json' \
  -d '{"name":"Smoke Device","comment":"registered by smoke test"}' >/dev/null
sleep 1

echo "[smoke] Verify registered filter"
REGISTERED_JSON="$(curl -fsS "${backend_url}/api/devices?status=registered")"
printf '%s' "$REGISTERED_JSON" > /tmp/mikrotik-registered.json
python3 - <<'PY'
import json
from pathlib import Path
payload = json.loads(Path('/tmp/mikrotik-registered.json').read_text())
items = payload.get('items', [])
print(f"registered_count={len(items)}")
if not any(item.get('name') == 'Smoke Device' for item in items):
    raise SystemExit('registered device not found')
PY

echo "[smoke] Toggle mock router offline"
curl -fsS "${mock_router_url}/admin/scenario?state=offline" >/dev/null
sleep 6

OFFLINE_JSON="$(curl -fsS "${backend_url}/api/devices?online=false")"
printf '%s' "$OFFLINE_JSON" > /tmp/mikrotik-offline.json
python3 - <<'PY'
import json
from pathlib import Path
payload = json.loads(Path('/tmp/mikrotik-offline.json').read_text())
items = payload.get('items', [])
print(f"offline_count={len(items)}")
if len(items) == 0:
    raise SystemExit('expected offline devices after scenario switch')
PY

echo "[smoke] Restore mock router online"
curl -fsS "${mock_router_url}/admin/scenario?state=online" >/dev/null

echo "[smoke] OK"
