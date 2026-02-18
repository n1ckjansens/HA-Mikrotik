package configsync

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type Watcher struct {
	baseURL string
	token   string
	logger  *slog.Logger
}

func NewWatcher(baseURL, token string, logger *slog.Logger) *Watcher {
	return &Watcher{baseURL: strings.TrimSuffix(baseURL, "/"), token: token, logger: logger}
}

func (w *Watcher) Run(ctx context.Context, onConfigUpdated func()) {
	backoff := time.Second
	for {
		if ctx.Err() != nil {
			return
		}
		err := w.runSession(ctx, onConfigUpdated)
		if err != nil && ctx.Err() == nil {
			w.logger.Warn("config event watcher disconnected", "err", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(backoff):
		}
		if backoff < 20*time.Second {
			backoff *= 2
		}
	}
}

func (w *Watcher) runSession(ctx context.Context, onConfigUpdated func()) error {
	wsURL, err := toWebsocketURL(w.baseURL + "/api/websocket")
	if err != nil {
		return err
	}
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, wsURL, nil)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, msg, err := conn.ReadMessage()
	if err != nil {
		return err
	}
	if !strings.Contains(string(msg), "auth_required") {
		return nil
	}

	authPayload := map[string]any{"type": "auth", "access_token": w.token}
	if err := conn.WriteJSON(authPayload); err != nil {
		return err
	}

	_, msg, err = conn.ReadMessage()
	if err != nil {
		return err
	}
	if !strings.Contains(string(msg), "auth_ok") {
		return nil
	}

	subscribe := map[string]any{"id": 1, "type": "subscribe_events", "event_type": "mikrotik_presence_config_updated"}
	if err := conn.WriteJSON(subscribe); err != nil {
		return err
	}

	if err := conn.SetReadDeadline(time.Now().Add(120 * time.Second)); err != nil {
		return err
	}
	for {
		if err := conn.SetReadDeadline(time.Now().Add(120 * time.Second)); err != nil {
			return err
		}
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return err
		}
		if isConfigUpdatedEvent(msg) {
			onConfigUpdated()
		}
	}
}

func isConfigUpdatedEvent(body []byte) bool {
	var envelope struct {
		Type  string `json:"type"`
		Event struct {
			EventType string `json:"event_type"`
		} `json:"event"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return false
	}
	return envelope.Type == "event" && envelope.Event.EventType == "mikrotik_presence_config_updated"
}

func toWebsocketURL(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", err
	}
	switch u.Scheme {
	case "http":
		u.Scheme = "ws"
	case "https":
		u.Scheme = "wss"
	}
	return u.String(), nil
}
