package sqlite

import (
	"context"
	"database/sql"
	"log/slog"

	legacyautomationrepo "github.com/micro-ha/mikrotik-presence/addon/internal/automation/repository"
	legacystorage "github.com/micro-ha/mikrotik-presence/addon/internal/storage"
)

// DB is root sqlite storage handle for repositories.
type DB struct {
	storage        *legacystorage.Repository
	automationRepo *legacyautomationrepo.Repository
	logger         *slog.Logger
}

// Open initializes sqlite database and runs migrations.
func Open(ctx context.Context, dbPath string, logger *slog.Logger) (*DB, error) {
	base, err := legacystorage.New(ctx, dbPath, logger)
	if err != nil {
		return nil, err
	}
	return &DB{
		storage:        base,
		automationRepo: legacyautomationrepo.New(base.SQLDB(), logger),
		logger:         logger,
	}, nil
}

// Close closes active sqlite connection pool.
func (d *DB) Close() error {
	if d == nil || d.storage == nil {
		return nil
	}
	return d.storage.Close()
}

// SQLDB returns low-level sql.DB for integrations requiring direct access.
func (d *DB) SQLDB() *sql.DB {
	if d == nil || d.storage == nil {
		return nil
	}
	return d.storage.SQLDB()
}
