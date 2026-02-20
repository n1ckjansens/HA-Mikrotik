package routeros

import (
	"context"
	"fmt"
	"time"

	"github.com/micro-ha/mikrotik-presence/addon/internal/model"
)

// Event is one RouterOS listen sentence mapped into structured payload.
type Event struct {
	Path       string
	Type       string
	Values     map[string]string
	ReceivedAt time.Time
}

// Listen starts event stream for /interface/listen or /ip/address/listen.
func (c *Client) Listen(ctx context.Context, path string) (<-chan Event, error) {
	path = stringsTrim(path)
	if path == "" {
		return nil, &ValidationError{Field: "path", Reason: "is required"}
	}
	if path != "/interface/listen" && path != "/ip/address/listen" {
		return nil, &ValidationError{Field: "path", Reason: "unsupported listen endpoint"}
	}

	out := make(chan Event, 128)
	c.listenerWG.Add(1)
	go c.listenLoop(ctx, path, out)
	return out, nil
}

func (c *Client) listenLoop(ctx context.Context, path string, out chan<- Event) {
	defer c.listenerWG.Done()
	defer close(out)

	backoff := 200 * time.Millisecond
	for {
		if ctx.Err() != nil || c.isClosed() {
			return
		}
		if err := c.connect(ctx); err != nil {
			c.logger.Warn("listen connect failed", "path", path, "err", err)
			if sleepErr := c.sleepFn(ctx, backoff); sleepErr != nil {
				return
			}
			backoff = nextBackoff(backoff)
			continue
		}

		conn := c.currentConn()
		if conn == nil {
			if sleepErr := c.sleepFn(ctx, backoff); sleepErr != nil {
				return
			}
			backoff = nextBackoff(backoff)
			continue
		}

		listener, err := c.listenFn(conn, path)
		if err != nil {
			c.logger.Warn("listen subscribe failed", "path", path, "err", err)
			if isRetryableError(err) {
				c.disconnect()
			}
			if sleepErr := c.sleepFn(ctx, backoff); sleepErr != nil {
				return
			}
			backoff = nextBackoff(backoff)
			continue
		}

		backoff = 200 * time.Millisecond
		for {
			select {
			case <-ctx.Done():
				cancelCtx, cancel := context.WithTimeout(context.Background(), c.config.Timeout)
				_, _ = listener.CancelContext(cancelCtx)
				cancel()
				return
			case sentence, ok := <-listener.Chan():
				if !ok {
					if err := listener.Err(); err != nil {
						c.logger.Warn("listen stream closed", "path", path, "err", err)
						if isRetryableError(err) {
							c.disconnect()
							break
						}
						return
					}
					if listener.Done != nil {
						c.logger.Info("listen stream done", "path", path, "reason", listener.Done.Map["message"])
					}
					return
				}

				event := Event{
					Path:       path,
					Type:       sentence.Word,
					Values:     mapSentence(sentence),
					ReceivedAt: time.Now().UTC(),
				}

				select {
				case out <- event:
				case <-ctx.Done():
					return
				}
			}
		}
	}
}

// Listen keeps compatibility for manager-level access when multiple routers are used.
func (m *Manager) Listen(ctx context.Context, cfg model.RouterConfig, path string) (<-chan Event, error) {
	client, err := m.getClient(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("get client for listen: %w", err)
	}
	return client.Listen(ctx, path)
}
