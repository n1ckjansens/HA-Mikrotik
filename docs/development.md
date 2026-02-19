# Development Guide

This guide explains how to run MikroTik Presence locally without Home Assistant, with either:
- Mock RouterOS (fully local)
- Real RouterOS device

## Prerequisites

- Docker + Docker Compose v2
- `curl`
- `python3` (for smoke test script)

## Quick Start (Mock mode)

1. Prepare env file:

```bash
cp env.dev.example .env.dev
```

2. Start full dev stack (backend + frontend + mock HA + mock RouterOS):

```bash
make dev-up-d
```

3. Open UI:

- Frontend dev server: `http://localhost:5173`
- Backend API: `http://localhost:8099`

4. Run smoke test:

```bash
make dev-smoke
```

5. Watch logs:

```bash
make dev-logs
```

6. Stop and clean:

```bash
make dev-down
```

## Real RouterOS Mode

Use same stack, but point mock-HA config to your real router.

1. Edit `.env.dev`:

```bash
ROUTER_HOST=192.168.88.1
ROUTER_USERNAME=admin
ROUTER_PASSWORD=your-password
ROUTER_SSL=true
ROUTER_VERIFY_TLS=false
ROUTER_POLL_INTERVAL_SEC=5
```

2. Start stack:

```bash
make dev-up-d
```

3. Trigger immediate refresh and inspect devices:

```bash
curl -fsS -X POST http://localhost:8099/api/refresh -H 'Content-Type: application/json' -d '{}'
curl -fsS http://localhost:8099/api/devices
```

### Важно про `.env.dev`

Параметры `ROUTER_*` читаются контейнером `mock-ha` при старте.
Если вы поменяли `.env.dev`, пересоздайте `mock-ha` и `backend`:

```bash
docker compose -f docker-compose.dev.yml up -d --force-recreate mock-ha backend
```

Проверить, какой роутер реально отдается backend:

```bash
curl -fsS http://127.0.0.1:8123/api/mikrotik_presence/config
```

## Hot Reload

- Backend: live reload via `air` inside `backend` container.
  - Edit any `.go` file in `addon/` and backend restarts automatically.
- Frontend: Vite HMR in `frontend` container.
  - Edit files in `addon/frontend/src` and browser updates instantly.

## Mock Control Endpoints

Mock RouterOS state control:

- Set online:

```bash
make mock-online
```

- Set offline:

```bash
make mock-offline
```

Direct endpoint:

```bash
curl "http://127.0.0.1:18080/admin/scenario?state=offline"
```

Mock HA config patch (increments config version):

```bash
curl -X POST http://127.0.0.1:8123/admin/config \
  -H 'Content-Type: application/json' \
  -d '{"poll_interval_sec":10}'
```

## Useful Endpoints

- `GET /healthz`
- `GET /api/devices`
- `GET /api/devices?status=registered`
- `GET /api/devices?online=true`
- `GET /api/devices/{mac}`
- `POST /api/devices/{mac}/register`
- `PATCH /api/devices/{mac}`
- `POST /api/refresh`

## Local Testing without Docker (optional)

Backend:

```bash
cd addon
go mod tidy
go test ./... -race
HTTP_ADDR=:8099 DB_PATH=/tmp/mikrotik_presence.db HA_BASE_URL=http://127.0.0.1:8123 SUPERVISOR_TOKEN= go run ./cmd/server
```

Frontend:

```bash
cd addon/frontend
npm install
VITE_API_PROXY_TARGET=http://127.0.0.1:8099 npm run dev
```
