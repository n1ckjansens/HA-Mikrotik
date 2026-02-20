package routeros

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	goros "github.com/go-routeros/routeros/v3"
	"github.com/micro-ha/mikrotik-presence/addon/internal/model"
)

// API describes the minimal command runner used by typed wrappers and mocks.
type API interface {
	Run(ctx context.Context, cmd string, args ...string) (*goros.Reply, error)
}

// MetricsHooks allows optional observability callbacks for production telemetry.
type MetricsHooks struct {
	ObserveRun       func(cmd string, success bool, elapsed time.Duration)
	ObserveReconnect func(address string, attempt int, success bool)
}

// Client is a RouterOS API client wrapper with reconnect/backoff and typed commands.
type Client struct {
	conn   *goros.Client
	config Config

	logger *slog.Logger

	mu        sync.RWMutex
	connectMu sync.Mutex
	runMu     sync.Mutex

	closeOnce   sync.Once
	closed      chan struct{}
	listenerWG  sync.WaitGroup
	addressList sync.Mutex

	api API

	dialFn   func(ctx context.Context, cfg Config) (*goros.Client, error)
	runFn    func(ctx context.Context, conn *goros.Client, cmd string, args ...string) (*goros.Reply, error)
	listenFn func(conn *goros.Client, sentence ...string) (*goros.ListenReply, error)
	closeFn  func(conn *goros.Client) error
	sleepFn  func(ctx context.Context, wait time.Duration) error

	metrics MetricsHooks
}

// Manager keeps per-router clients and supports multiple RouterOS devices in one process.
type Manager struct {
	mu      sync.RWMutex
	clients map[string]*Client
	logger  *slog.Logger
	metrics MetricsHooks
}

// New creates and connects RouterOS client.
func New(ctx context.Context, cfg Config) (*Client, error) {
	normalized, err := normalizeConfig(cfg)
	if err != nil {
		return nil, err
	}

	client := &Client{
		config: normalized,
		logger: slog.Default(),
		closed: make(chan struct{}),
		dialFn: dialRouter,
		runFn:  runCommand,
		listenFn: func(conn *goros.Client, sentence ...string) (*goros.ListenReply, error) {
			return conn.Listen(sentence...)
		},
		closeFn: func(conn *goros.Client) error {
			return conn.Close()
		},
		sleepFn: sleepWithContext,
	}

	if err := client.connect(ctx); err != nil {
		return nil, err
	}
	return client, nil
}

// WithLogger attaches structured logger.
func (c *Client) WithLogger(logger *slog.Logger) *Client {
	if logger == nil {
		return c
	}
	c.logger = logger
	return c
}

// WithMetrics attaches optional metrics hooks.
func (c *Client) WithMetrics(hooks MetricsHooks) *Client {
	c.metrics = hooks
	return c
}

// Close gracefully shuts down connection and listener goroutines.
func (c *Client) Close() error {
	if c == nil {
		return nil
	}

	var closeErr error
	c.closeOnce.Do(func() {
		close(c.closed)
		conn := c.swapConn(nil)
		if conn != nil {
			if err := c.closeFn(conn); err != nil {
				closeErr = fmt.Errorf("close routeros connection: %w", err)
			}
		}

		done := make(chan struct{})
		go func() {
			c.listenerWG.Wait()
			close(done)
		}()

		timeout := c.config.Timeout
		if timeout <= 0 {
			timeout = 5 * time.Second
		}
		select {
		case <-done:
		case <-time.After(timeout):
			c.logger.Warn("routeros listener shutdown timeout", "timeout", timeout.String())
		}
	})
	return closeErr
}

// Run executes one RouterOS command with retry/reconnect logic.
func (c *Client) Run(ctx context.Context, cmd string, args ...string) (*goros.Reply, error) {
	if c == nil {
		return nil, &ValidationError{Field: "client", Reason: "is nil"}
	}
	cmd = stringsTrim(cmd)
	if cmd == "" {
		return nil, &ValidationError{Field: "cmd", Reason: "is required"}
	}
	if c.isClosed() {
		return nil, errors.New("routeros client is closed")
	}

	const maxAttempts = 4
	backoff := 120 * time.Millisecond
	var lastErr error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if err := c.connect(ctx); err != nil {
			lastErr = err
			if attempt == maxAttempts {
				break
			}
			if sleepErr := c.sleepFn(ctx, backoff); sleepErr != nil {
				return nil, fmt.Errorf("routeros run %s canceled: %w", cmd, sleepErr)
			}
			backoff = nextBackoff(backoff)
			continue
		}

		conn := c.currentConn()
		if conn == nil {
			lastErr = errors.New("routeros connection is nil")
			continue
		}

		runCtx, cancel := withTimeout(ctx, c.config.Timeout)
		started := time.Now()

		c.runMu.Lock()
		reply, err := c.runFn(runCtx, conn, cmd, args...)
		c.runMu.Unlock()
		cancel()

		if c.metrics.ObserveRun != nil {
			c.metrics.ObserveRun(cmd, err == nil, time.Since(started))
		}

		if err == nil {
			return reply, nil
		}
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			if ctx.Err() != nil {
				return nil, fmt.Errorf("routeros run %s canceled: %w", cmd, ctx.Err())
			}
		}
		if !isRetryableError(err) {
			return nil, fmt.Errorf("routeros run %s failed: %w", cmd, err)
		}

		lastErr = err
		c.logger.Warn("routeros command failed; reconnecting", "cmd", cmd, "attempt", attempt, "err", err)
		c.disconnect()

		if attempt == maxAttempts {
			break
		}
		if sleepErr := c.sleepFn(ctx, backoff); sleepErr != nil {
			return nil, fmt.Errorf("routeros run %s canceled: %w", cmd, sleepErr)
		}
		backoff = nextBackoff(backoff)
	}

	return nil, fmt.Errorf("routeros run %s failed after retries: %w", cmd, lastErr)
}

// RunCommand converts param map to RouterOS words and maps !re sentences.
func (c *Client) RunCommand(ctx context.Context, path string, params map[string]string) ([]map[string]string, error) {
	path = stringsTrim(path)
	if path == "" {
		return nil, &ValidationError{Field: "path", Reason: "is required"}
	}

	words := mapParams(params)
	runner := API(c)
	if c.api != nil {
		runner = c.api
	}

	reply, err := runner.Run(ctx, path, words...)
	if err != nil {
		return nil, fmt.Errorf("run command %s: %w", path, err)
	}
	return mapReplyRows(reply), nil
}

// RunCommand executes command on pooled client selected by cfg.
func (m *Manager) RunCommand(
	ctx context.Context,
	cfg model.RouterConfig,
	path string,
	params map[string]string,
) ([]map[string]string, error) {
	client, err := m.getClient(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return client.RunCommand(ctx, path, params)
}

// Healthcheck verifies RouterOS API command path is available.
func (c *Client) Healthcheck(ctx context.Context) error {
	_, err := c.RunCommand(ctx, "/system/identity/print", map[string]string{
		".proplist": "name",
	})
	if err != nil {
		return fmt.Errorf("routeros healthcheck failed: %w", err)
	}
	return nil
}

// NewManager creates multi-router manager.
func NewManager(logger *slog.Logger) *Manager {
	if logger == nil {
		logger = slog.Default()
	}
	return &Manager{
		clients: make(map[string]*Client),
		logger:  logger,
	}
}

// WithMetrics attaches optional metrics hooks for all newly created clients.
func (m *Manager) WithMetrics(hooks MetricsHooks) *Manager {
	m.metrics = hooks
	return m
}

// Close closes all pooled clients.
func (m *Manager) Close() error {
	if m == nil {
		return nil
	}

	m.mu.Lock()
	clients := make([]*Client, 0, len(m.clients))
	for _, client := range m.clients {
		clients = append(clients, client)
	}
	m.clients = make(map[string]*Client)
	m.mu.Unlock()

	var firstErr error
	for _, client := range clients {
		if err := client.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (m *Manager) getClient(ctx context.Context, cfg model.RouterConfig) (*Client, error) {
	normalized, err := normalizeConfig(configFromModel(cfg))
	if err != nil {
		return nil, err
	}
	key := configKey(normalized)

	m.mu.RLock()
	existing := m.clients[key]
	m.mu.RUnlock()
	if existing != nil {
		return existing, nil
	}

	created, err := New(ctx, normalized)
	if err != nil {
		return nil, err
	}
	created.WithLogger(m.logger.With("router_address", normalized.Address)).WithMetrics(m.metrics)

	m.mu.Lock()
	defer m.mu.Unlock()
	if existing = m.clients[key]; existing != nil {
		_ = created.Close()
		return existing, nil
	}
	m.clients[key] = created
	return created, nil
}

// Healthcheck verifies RouterOS API availability for selected router config.
func (m *Manager) Healthcheck(ctx context.Context, cfg model.RouterConfig) error {
	client, err := m.getClient(ctx, cfg)
	if err != nil {
		return err
	}
	return client.Healthcheck(ctx)
}

func (c *Client) connect(ctx context.Context) error {
	if c.isClosed() {
		return errors.New("routeros client is closed")
	}
	if c.currentConn() != nil {
		return nil
	}

	c.connectMu.Lock()
	defer c.connectMu.Unlock()

	if c.currentConn() != nil {
		return nil
	}

	const maxAttempts = 5
	backoff := 150 * time.Millisecond
	var lastErr error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if c.metrics.ObserveReconnect != nil {
			c.metrics.ObserveReconnect(c.config.Address, attempt, false)
		}

		dialCtx, cancel := withTimeout(ctx, c.config.Timeout)
		conn, err := c.dialFn(dialCtx, c.config)
		cancel()
		if err == nil {
			c.swapConn(conn)
			if c.metrics.ObserveReconnect != nil {
				c.metrics.ObserveReconnect(c.config.Address, attempt, true)
			}
			c.logger.Info("routeros connected", "address", c.config.Address, "attempt", attempt)
			return nil
		}

		lastErr = err
		c.logger.Warn("routeros dial failed", "address", c.config.Address, "attempt", attempt, "err", err)

		if attempt == maxAttempts {
			break
		}
		if sleepErr := c.sleepFn(ctx, backoff); sleepErr != nil {
			return fmt.Errorf("routeros reconnect canceled: %w", sleepErr)
		}
		backoff = nextBackoff(backoff)
	}

	return &ReconnectError{Address: c.config.Address, Attempts: maxAttempts, Err: lastErr}
}

func (c *Client) disconnect() {
	old := c.swapConn(nil)
	if old != nil {
		_ = c.closeFn(old)
	}
}

func (c *Client) currentConn() *goros.Client {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.conn
}

func (c *Client) swapConn(next *goros.Client) *goros.Client {
	c.mu.Lock()
	defer c.mu.Unlock()
	prev := c.conn
	c.conn = next
	return prev
}

func (c *Client) isClosed() bool {
	select {
	case <-c.closed:
		return true
	default:
		return false
	}
}

func dialRouter(ctx context.Context, cfg Config) (*goros.Client, error) {
	if cfg.UseTLS {
		return goros.DialTLSContext(ctx, cfg.Address, cfg.Username, cfg.Password, &tls.Config{
			InsecureSkipVerify: !cfg.VerifyTLS, //nolint:gosec
		})
	}
	return goros.DialContext(ctx, cfg.Address, cfg.Username, cfg.Password)
}

func runCommand(ctx context.Context, conn *goros.Client, cmd string, args ...string) (*goros.Reply, error) {
	return conn.RunContext(ctx, append([]string{cmd}, args...)...)
}

func sleepWithContext(ctx context.Context, wait time.Duration) error {
	timer := time.NewTimer(wait)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func withTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout <= 0 {
		return ctx, func() {}
	}
	deadline, ok := ctx.Deadline()
	if !ok {
		return context.WithTimeout(ctx, timeout)
	}
	if time.Until(deadline) <= timeout {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, timeout)
}

func nextBackoff(current time.Duration) time.Duration {
	next := current * 2
	if next > 2*time.Second {
		return 2 * time.Second
	}
	return next
}

func stringsTrim(value string) string {
	return strings.TrimSpace(value)
}
