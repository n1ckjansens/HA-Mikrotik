package httpapi

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// LogProvider provides request logger for middleware.
type LogProvider interface {
	Logger() *slog.Logger
}

// RequestLogger logs basic structured request/response metadata.
func RequestLogger(provider LogProvider) func(http.Handler) http.Handler {
	logger := slog.Default()
	if provider != nil && provider.Logger() != nil {
		logger = provider.Logger()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startedAt := time.Now()
			wrapped := &responseCapture{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(wrapped, r)
			logger.Info(
				"http request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", wrapped.statusCode,
				"bytes", wrapped.size,
				"duration_ms", time.Since(startedAt).Milliseconds(),
			)
		})
	}
}

// StripIngressPrefix removes ingress path prefix sent in reverse proxy header.
func StripIngressPrefix(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		prefix := strings.TrimSpace(r.Header.Get("X-Ingress-Path"))
		if prefix != "" && strings.HasPrefix(r.URL.Path, prefix) {
			r.URL.Path = strings.TrimPrefix(r.URL.Path, prefix)
			if r.URL.Path == "" {
				r.URL.Path = "/"
			}
		}
		next.ServeHTTP(w, r)
	})
}

// RecoverJSON converts panic into structured JSON error response.
func RecoverJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				slog.Default().Error("panic recovered", "panic", fmt.Sprint(recovered), "path", r.URL.Path)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(map[string]any{
					"error": map[string]any{
						"code":    "internal_error",
						"message": "Internal server error",
					},
				})
			}
		}()
		next.ServeHTTP(w, r)
	})
}

type responseCapture struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func (w *responseCapture) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *responseCapture) Write(body []byte) (int, error) {
	size, err := w.ResponseWriter.Write(body)
	w.size += size
	return size, err
}
