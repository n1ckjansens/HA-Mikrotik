package sqlite

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/micro-ha/mikrotik-presence/addon/internal/storage"
)

// DB is root sqlite storage handle for repositories.
type DB struct {
	storage *storage.Repository
	logger  *slog.Logger
}

// Open initializes sqlite database and runs migrations.
func Open(ctx context.Context, dbPath string, logger *slog.Logger) (*DB, error) {
	base, err := storage.New(ctx, dbPath, logger)
	if err != nil {
		return nil, err
	}
	return &DB{
		storage: base,
		logger:  logger,
	}, nil
}

// Close closes active sqlite connection pool.
func (d *DB) Close() error {
	if d == nil || d.storage == nil {
		return nil
	}
	return d.storage.Close()
}

// SQLDB returns low-level sql.DB for callers requiring direct access.
func (d *DB) SQLDB() *sql.DB {
	if d == nil || d.storage == nil {
		return nil
	}
	return d.storage.SQLDB()
}
