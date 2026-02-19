package repository

import (
	"database/sql"
	"errors"
	"log/slog"
)

var ErrNotFound = errors.New("not found")

type Repository struct {
	db     *sql.DB
	logger *slog.Logger
}

func New(db *sql.DB, logger *slog.Logger) *Repository {
	return &Repository{db: db, logger: logger}
}
