# MikroTik Presence for Home Assistant

MikroTik Presence is a single Home Assistant product:
- Home Assistant Add-on with Ingress Web UI

Installation flow:
1. Add this repository in Home Assistant Add-on Store.
2. Install **MikroTik Presence** add-on.
3. Open add-on **Configuration** tab and fill router settings.
4. Start add-on and open the UI.

## Repository Structure

- `addon/` - Go backend, SQLite persistence, RouterOS poller, ingress server.
- `addon/frontend/` - React/Vite/TypeScript UI with shadcn/ui.

### Backend Architecture (`addon/internal`)

- `config` - runtime config loading from env.
- `logging` - centralized slog initialization.
- `domain/device`, `domain/automation` - domain models and interfaces.
- `services/device`, `services/automation` - use-cases and automation engine.
- `services/automation/registry` - pluggable `Action` and `StateSource` registry.
- `repository/sqlite` - repository implementations and migrations.
- `adapters/mikrotik` - RouterOS adapter + action/state-source primitives.
- `http` and `http/handlers` - transport layer (router, middleware, handlers).

## Features

- RouterOS v7 REST polling (`/ip/dhcp-server/lease`, `/interface/wifi/registration-table`, `/interface/bridge/host`, `/ip/arp`, `/ip/address`).
- Device aggregation by MAC.
- `new` vs `registered` status.
- Persistent SQLite data in `/data`.
- Register and edit devices from UI.
- Polling interval configurable in add-on options (minimum 5s).

## Development

Full local development (Docker Compose, mock/real RouterOS, smoke tests, hot reload):

- `docs/development.md`
- `docs/router-setup.md` (подготовка RouterOS v7 под add-on)

### Backend checks

```bash
cd addon
go test ./... -race
```

### Frontend checks

```bash
cd addon/frontend
npm install
npm run lint
npm run typecheck
npm run build
```

## API

- `GET /api/devices?status=new|registered&online=true|false&query=...`
- `GET /api/devices/{mac}`
- `POST /api/devices/{mac}/register`
- `PATCH /api/devices/{mac}`
- `POST /api/refresh`
- `GET /api/automation/action-types`
- `GET /api/automation/state-source-types`
- `GET /api/automation/capabilities`
- `GET /api/automation/capabilities/{id}`
- `POST /api/automation/capabilities`
- `PUT /api/automation/capabilities/{id}`
- `DELETE /api/automation/capabilities/{id}`
- `GET /api/automation/capabilities/{id}/devices`
- `PATCH /api/automation/capabilities/{id}/devices/{mac}`
- `GET /api/devices/{mac}/capabilities`
- `PATCH /api/devices/{mac}/capabilities/{capabilityId}`
- `GET /healthz`

All API routes are ingress-aware.
