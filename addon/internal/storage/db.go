package storage

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
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
			last_sources_json TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS devices_new_cache (
			mac TEXT PRIMARY KEY,
			first_seen_at TEXT NOT NULL,
			vendor TEXT NOT NULL,
			generated_name TEXT NOT NULL
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
	return r.normalizeLegacyMACKeys(ctx)
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
