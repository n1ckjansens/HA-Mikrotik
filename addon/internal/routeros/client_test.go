package routeros

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	goros "github.com/go-routeros/routeros/v3"
	"github.com/go-routeros/routeros/v3/proto"
)

func TestRunReconnectsOnNetworkFailure(t *testing.T) {
	t.Helper()

	var (
		mu        sync.Mutex
		dialCalls int
		runCalls  int
	)

	client := &Client{
		config: Config{Address: "127.0.0.1:8728", Username: "u", Password: "p", Timeout: time.Second},
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		closed: make(chan struct{}),
		dialFn: func(ctx context.Context, cfg Config) (*goros.Client, error) {
			_ = ctx
			_ = cfg
			mu.Lock()
			dialCalls++
			mu.Unlock()
			return &goros.Client{}, nil
		},
		runFn: func(ctx context.Context, conn *goros.Client, cmd string, args ...string) (*goros.Reply, error) {
			_ = ctx
			_ = conn
			_ = cmd
			_ = args
			mu.Lock()
			runCalls++
			current := runCalls
			mu.Unlock()
			if current == 1 {
				return nil, io.EOF
			}
			return &goros.Reply{Done: &proto.Sentence{Word: "!done", Map: map[string]string{}}}, nil
		},
		closeFn: func(conn *goros.Client) error {
			_ = conn
			return nil
		},
		sleepFn: func(ctx context.Context, wait time.Duration) error {
			_ = ctx
			_ = wait
			return nil
		},
	}

	reply, err := client.Run(context.Background(), "/system/identity/print")
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if reply == nil || reply.Done == nil {
		t.Fatalf("expected non-nil done sentence")
	}

	mu.Lock()
	defer mu.Unlock()
	if dialCalls != 2 {
		t.Fatalf("expected 2 dial attempts, got %d", dialCalls)
	}
	if runCalls != 2 {
		t.Fatalf("expected 2 run attempts, got %d", runCalls)
	}
}

func TestRunPropagatesNonRetryableError(t *testing.T) {
	t.Helper()

	expected := errors.New("permission denied")
	client := &Client{
		config: Config{Address: "127.0.0.1:8728", Username: "u", Password: "p", Timeout: time.Second},
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		closed: make(chan struct{}),
		dialFn: func(ctx context.Context, cfg Config) (*goros.Client, error) {
			_ = ctx
			_ = cfg
			return &goros.Client{}, nil
		},
		runFn: func(ctx context.Context, conn *goros.Client, cmd string, args ...string) (*goros.Reply, error) {
			_ = ctx
			_ = conn
			_ = cmd
			_ = args
			return nil, expected
		},
		closeFn: func(conn *goros.Client) error {
			_ = conn
			return nil
		},
		sleepFn: func(ctx context.Context, wait time.Duration) error {
			_ = ctx
			_ = wait
			return nil
		},
	}

	_, err := client.Run(context.Background(), "/x")
	if err == nil {
		t.Fatalf("expected error")
	}
	if !errors.Is(err, expected) {
		t.Fatalf("expected wrapped original error, got %v", err)
	}
}
