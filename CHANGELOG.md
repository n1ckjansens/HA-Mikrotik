# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [v0.0.7] - 2026-02-20

### Added
- Added Home Assistant custom integration `mikrotik_presence` with:
  - config flow with add-on discovery (`async_step_hassio`) and manual fallback
  - async backend API client
  - coordinators for device/global capabilities
  - `switch` and `select` platforms with stable `unique_id` and device registry mapping
  - translations (`en`/`ru`) and HACS metadata (`hacs.json`)
- Added official Supervisor add-on management via `AddonManager`:
  - auto-connect when add-on is discovered
  - user-selectable add-on install/start flow
  - API health-check (`GET /api/devices`) before creating config entry

### Changed
- Bumped add-on version to `v0.0.7`:
  - `addon/config.json`
- Added add-on discovery metadata and standardized add-on API port to `8080`:
  - `addon/config.json`
  - `addon/config.yaml`
  - `addon/Dockerfile`
  - `addon/internal/config/config.go`
- Bumped custom integration version to `v0.0.7`:
  - `custom_components/mikrotik_presence/manifest.json`
- Bumped frontend package version to `v0.0.7`:
  - `addon/frontend/package.json`
  - `addon/frontend/package-lock.json`

## [0.0.6] - 2026-02-20

### Added
- Added resilient clipboard helper with fallback copy strategy:
  - `addon/frontend/src/lib/clipboard.ts`
- Added shadcn pagination UI component:
  - `addon/frontend/src/components/ui/pagination.tsx`

### Changed
- Updated devices table footer to use shadcn pagination controls with page links and ellipsis:
  - `addon/frontend/src/components/devices/DevicesTable.tsx`
- Updated copy actions in device table to use shared clipboard helper:
  - `addon/frontend/src/components/devices/CopyValue.tsx`
  - `addon/frontend/src/components/devices/device-table-columns.tsx`
- Updated devices list query key shape to remove client-side pagination fields from query identity:
  - `addon/frontend/src/hooks/useDevicesListQuery.ts`
  - `addon/frontend/src/lib/query-keys.ts`
- Bumped add-on version to `v0.0.6`:
  - `addon/config.json`

### Fixed
- Fixed MAC copy action not triggering in table rows.
- Fixed pagination behavior resetting to page 1 on polling/refetch.
- Kept explicit page reset on filter/search changes while preventing reset on regular background refetch.
- Added page index clamping when filtered result size shrinks:
  - `addon/frontend/src/pages/DevicesPage.tsx`

## [0.0.5] - 2026-02-20

### Added
- New RouterOS integration modules:
  - `addon/internal/routeros/config.go`
  - `addon/internal/routeros/errors.go`
  - `addon/internal/routeros/mapper.go`
  - `addon/internal/routeros/events.go`
  - `addon/internal/routeros/firewall.go`
  - `addon/internal/routeros/interfaces.go`
  - `addon/internal/routeros/monitor.go`
  - `addon/internal/routeros/addresslist.go`
- RouterOS mock client for tests:
  - `addon/internal/routeros/mock/client.go`
- New test suites:
  - `addon/internal/routeros/addresslist_test.go`
  - `addon/internal/routeros/firewall_test.go`
  - `addon/internal/routeros/monitor_test.go`
  - `addon/internal/services/device/service_test.go`
- Aggregator ARP completeness scenario test:
  - `addon/internal/aggregator/aggregator_test.go`

### Changed
- Refactored RouterOS client layer to new architecture based on `go-routeros/v3`.
- Updated RouterOS command/monitor behavior for ARP completeness handling (`complete`, `status`, `flags` fallback).
- Updated device snapshot persistence logic to keep unregistered observed clients (including ARP-only observations) instead of deleting non-ONLINE observations immediately.
- Updated server wiring to use the new RouterOS integration entrypoint:
  - `addon/cmd/server/main.go`
- Updated dependencies:
  - `addon/go.mod`
  - `addon/go.sum`
- Bumped add-on version to `v0.0.5`:
  - `addon/config.json`

### Removed
- Removed legacy REST adapter:
  - `addon/internal/adapters/mikrotik/rest_client.go`
- Removed legacy RouterOS files:
  - `addon/internal/routeros/address_list.go`
  - `addon/internal/routeros/address_list_lookup.go`
  - `addon/internal/routeros/address_list_remove.go`
  - `addon/internal/routeros/address_list_test.go`
  - `addon/internal/routeros/firewall_rule.go`
  - `addon/internal/routeros/firewall_rule_test.go`
  - `addon/internal/routeros/types.go`

### Fixed
- Fixed missing wired clients in "new" discovery flow when clients are observed via ARP but not DHCP.
- Fixed ARP completeness interpretation to support different RouterOS field variants.

### Commits included
- `8632236` feat(routeros): refactor firewall rule management and introduce interface handling
- `cca370f` feat(aggregator, routeros, services): enhance ARP handling and add tests for completeness
- `9bb24ba` fix: update version number to v0.0.5 in config.json for MikroTik Presence integration
