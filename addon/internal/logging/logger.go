package logging

import (
	"log/slog"
	"os"
)

// New creates a process logger with JSON output for backend services.
func New(level slog.Level) *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
}
