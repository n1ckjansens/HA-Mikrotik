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

2. Start full dev stack (backend + frontend + mock RouterOS):

```bash
make dev-up-d
```

3. Open UI:

- Frontend dev server: `http://localhost:5173`
- Backend API: `http://localhost:8080`

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

1. Edit `.env.dev`:

```bash
ROUTER_HOST=192.168.88.1
ROUTER_USERNAME=admin
ROUTER_PASSWORD=your-password
ROUTER_SSL=true
ROUTER_VERIFY_TLS=false
ROUTER_POLL_INTERVAL_SEC=5
```

2. Recreate backend after env update:

```bash
docker compose -f docker-compose.dev.yml up -d --force-recreate backend
```

3. Trigger immediate refresh and inspect devices:

```bash
curl -fsS -X POST http://localhost:8080/api/refresh -H 'Content-Type: application/json' -d '{}'
curl -fsS http://localhost:8080/api/devices
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
HTTP_ADDR=:8080 DB_PATH=/tmp/mikrotik_presence.db ROUTER_HOST=127.0.0.1:18080 ROUTER_USERNAME=admin ROUTER_PASSWORD=admin ROUTER_SSL=false ROUTER_VERIFY_TLS=false ROUTER_POLL_INTERVAL_SEC=5 go run ./cmd/server
```

Frontend:

```bash
cd addon/frontend
npm install
VITE_API_PROXY_TARGET=http://127.0.0.1:8080 npm run dev
```
