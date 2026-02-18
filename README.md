# MikroTik Presence for Home Assistant

MikroTik Presence is a single Home Assistant product:
- Home Assistant Add-on with Ingress Web UI
- Home Assistant Custom Integration (Config Flow)

Users install it as one solution:
1. Add this repository in Home Assistant Add-on Store.
2. Install **MikroTik Presence** add-on.
3. Add **MikroTik Presence** integration in **Devices & Services**.
4. Open add-on UI. Devices start syncing automatically.

No duplicate manual configuration is required in the add-on.

## Repository Structure

- `custom_components/mikrotik_presence/` - integration with config flow and internal API endpoint.
- `addon/` - Go backend, SQLite persistence, RouterOS poller, ingress server.
- `frontend/` - React/Vite/TypeScript UI with shadcn/ui.

## Features

- RouterOS v7 REST polling (`/ip/dhcp-server/lease`, `/interface/wifi/registration-table`, `/interface/bridge/host`, `/ip/arp`, `/ip/address`).
- Device aggregation by MAC.
- `new` vs `registered` status.
- Persistent SQLite data in `/data`.
- Register and edit devices from UI.
- Polling interval configurable in integration (minimum 5s).
- Config sync from Integration to Add-on using Home Assistant API.

## Development

### Backend checks

```bash
cd addon
go test ./... -race
```

### Frontend checks

```bash
cd frontend
npm install
npm run lint
npm run typecheck
npm run build
```

### Sync frontend bundle into add-on image context

```bash
cd addon
./sync_frontend.sh
```

This syncs `frontend/` source into `addon/frontend/` so the add-on Docker image can build the ingress UI.

## API

- `GET /api/devices?status=new|registered&online=true|false&query=...`
- `GET /api/devices/{mac}`
- `POST /api/devices/{mac}/register`
- `PATCH /api/devices/{mac}`
- `POST /api/refresh`
- `GET /healthz`

All API routes are ingress-aware.
