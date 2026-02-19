package storage

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type Repository struct {
	db     *sql.DB
	logger *slog.Logger
}

func New(ctx context.Context, dbPath string, logger *slog.Logger) (*Repository, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	db.SetConnMaxLifetime(0)
	db.SetConnMaxIdleTime(0)

	repo := &Repository{db: db, logger: logger}
	if err := repo.migrate(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return repo, nil
}

func (r *Repository) Close() error {
	if r == nil || r.db == nil {
		return nil
	}
	return r.db.Close()
}

func (r *Repository) migrate(ctx context.Context) error {
	statements := []string{
		`PRAGMA journal_mode = WAL;`,
		`CREATE TABLE IF NOT EXISTS devices_registered (
			mac TEXT PRIMARY KEY,
			name TEXT,
			icon TEXT,
			comment TEXT,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS devices_state (
			mac TEXT PRIMARY KEY,
			online INTEGER NOT NULL,
			last_seen_at TEXT,
			connected_since_at TEXT,
			last_ip TEXT,
			last_subnet TEXT,
			host_name TEXT,
			interface TEXT,
			bridge TEXT,
			ssid TEXT,
			dhcp_server TEXT,
			dhcp_status TEXT,
			dhcp_last_seen_sec INTEGER,
			wifi_driver TEXT,
			wifi_interface TEXT,
			wifi_last_activity_sec INTEGER,
			wifi_uptime_sec INTEGER,
			wifi_auth_type TEXT,
			wifi_signal INTEGER,
			arp_ip TEXT,
			arp_interface TEXT,
			arp_is_complete INTEGER NOT NULL DEFAULT 0,
			bridge_host_port TEXT,
			bridge_host_vlan INTEGER,
			connection_status TEXT NOT NULL DEFAULT 'UNKNOWN',
			status_reason TEXT NOT NULL DEFAULT '',
			last_sources_json TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS devices_new_cache (
			mac TEXT PRIMARY KEY,
			first_seen_at TEXT NOT NULL,
			vendor TEXT NOT NULL,
			generated_name TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS capability_templates (
			id TEXT PRIMARY KEY,
			data TEXT NOT NULL,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS device_capabilities_state (
			device_id TEXT NOT NULL,
			capability_id TEXT NOT NULL,
			enabled INTEGER NOT NULL,
			state TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			PRIMARY KEY (device_id, capability_id)
		);`,
	}

	for _, stmt := range statements {
		if _, err := r.db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("migrate failed: %w", err)
		}
	}
	if _, err := r.db.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_devices_state_online ON devices_state(online);`); err != nil {
		return err
	}
	if _, err := r.db.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_capabilities_updated_at ON capability_templates(updated_at);`); err != nil {
		return err
	}
	if _, err := r.db.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_device_cap_state_device ON device_capabilities_state(device_id);`); err != nil {
		return err
	}
	if _, err := r.db.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_device_cap_state_capability ON device_capabilities_state(capability_id);`); err != nil {
		return err
	}
	if err := r.ensureStateColumns(ctx); err != nil {
		return err
	}
	return r.normalizeLegacyMACKeys(ctx)
}

func (r *Repository) ensureStateColumns(ctx context.Context) error {
	columns := []string{
		`ALTER TABLE devices_state ADD COLUMN host_name TEXT`,
		`ALTER TABLE devices_state ADD COLUMN interface TEXT`,
		`ALTER TABLE devices_state ADD COLUMN bridge TEXT`,
		`ALTER TABLE devices_state ADD COLUMN ssid TEXT`,
		`ALTER TABLE devices_state ADD COLUMN dhcp_server TEXT`,
		`ALTER TABLE devices_state ADD COLUMN dhcp_status TEXT`,
		`ALTER TABLE devices_state ADD COLUMN dhcp_last_seen_sec INTEGER`,
		`ALTER TABLE devices_state ADD COLUMN wifi_driver TEXT`,
		`ALTER TABLE devices_state ADD COLUMN wifi_interface TEXT`,
		`ALTER TABLE devices_state ADD COLUMN wifi_last_activity_sec INTEGER`,
		`ALTER TABLE devices_state ADD COLUMN wifi_uptime_sec INTEGER`,
		`ALTER TABLE devices_state ADD COLUMN wifi_auth_type TEXT`,
		`ALTER TABLE devices_state ADD COLUMN wifi_signal INTEGER`,
		`ALTER TABLE devices_state ADD COLUMN arp_ip TEXT`,
		`ALTER TABLE devices_state ADD COLUMN arp_interface TEXT`,
		`ALTER TABLE devices_state ADD COLUMN arp_is_complete INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE devices_state ADD COLUMN bridge_host_port TEXT`,
		`ALTER TABLE devices_state ADD COLUMN bridge_host_vlan INTEGER`,
		`ALTER TABLE devices_state ADD COLUMN connection_status TEXT NOT NULL DEFAULT 'UNKNOWN'`,
		`ALTER TABLE devices_state ADD COLUMN status_reason TEXT NOT NULL DEFAULT ''`,
	}

	for _, stmt := range columns {
		if _, err := r.db.ExecContext(ctx, stmt); err != nil && !isDuplicateColumnError(err) {
			return fmt.Errorf("state schema update failed: %w", err)
		}
	}
	return nil
}

func isDuplicateColumnError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(strings.ToLower(err.Error()), "duplicate column name")
}

func (r *Repository) normalizeLegacyMACKeys(ctx context.Context) error {
	tables := []string{"devices_registered", "devices_state", "devices_new_cache"}
	for _, table := range tables {
		updateStmt := "UPDATE OR IGNORE " + table + " SET mac = REPLACE(REPLACE(UPPER(TRIM(mac)), '%3A', ':'), '-', ':') " +
			"WHERE mac LIKE '%3A%' OR mac LIKE '%3a%' OR mac LIKE '%-%' OR mac != UPPER(mac) OR mac != TRIM(mac);"
		res, err := r.db.ExecContext(ctx, updateStmt)
		if err != nil {
			return fmt.Errorf("legacy mac normalization failed for %s: %w", table, err)
		}
		if rows, _ := res.RowsAffected(); rows > 0 && r.logger != nil {
			r.logger.Info("normalized legacy mac rows", "table", table, "rows", rows)
		}

		deleteStmt := "DELETE FROM " + table + " WHERE mac LIKE '%3A%' OR mac LIKE '%3a%';"
		res, err = r.db.ExecContext(ctx, deleteStmt)
		if err != nil {
			return fmt.Errorf("legacy mac cleanup failed for %s: %w", table, err)
		}
		if rows, _ := res.RowsAffected(); rows > 0 && r.logger != nil {
			r.logger.Warn("removed conflicting legacy mac rows", "table", table, "rows", rows)
		}
	}
	return nil
}

func toTimePtr(v sql.NullString) *time.Time {
	if !v.Valid || v.String == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339Nano, v.String)
	if err != nil {
		return nil
	}
	u := t.UTC()
	return &u
}

func fromTimePtr(v *time.Time) any {
	if v == nil {
		return nil
	}
	return v.UTC().Format(time.RFC3339Nano)
}

func fromStringPtr(v *string) any {
	if v == nil {
		return nil
	}
	return *v
}

func strPtr(v sql.NullString) *string {
	if !v.Valid {
		return nil
	}
	s := v.String
	return &s
}

func int64Ptr(v sql.NullInt64) *int64 {
	if !v.Valid {
		return nil
	}
	value := v.Int64
	return &value
}

func intPtr(v sql.NullInt64) *int {
	if !v.Valid {
		return nil
	}
	value := int(v.Int64)
	return &value
}

func fromInt64Ptr(v *int64) any {
	if v == nil {
		return nil
	}
	return *v
}

func fromIntPtr(v *int) any {
	if v == nil {
		return nil
	}
	return *v
}
