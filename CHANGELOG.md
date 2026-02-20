# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

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
