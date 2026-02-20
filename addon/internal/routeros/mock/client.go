package mock

import (
	"context"
	"sync"

	goros "github.com/go-routeros/routeros/v3"
	"github.com/go-routeros/routeros/v3/proto"
)

// API is test seam for RouterOS command execution.
type API interface {
	Run(ctx context.Context, cmd string, args ...string) (*goros.Reply, error)
}

// Call stores one Run invocation.
type Call struct {
	Cmd  string
	Args []string
}

// Client is a programmable mock implementation of API.
type Client struct {
	mu      sync.Mutex
	RunFunc func(ctx context.Context, cmd string, args ...string) (*goros.Reply, error)
	Calls   []Call
}

func (c *Client) Run(ctx context.Context, cmd string, args ...string) (*goros.Reply, error) {
	c.mu.Lock()
	c.Calls = append(c.Calls, Call{Cmd: cmd, Args: append([]string(nil), args...)})
	run := c.RunFunc
	c.mu.Unlock()

	if run == nil {
		return &goros.Reply{Done: &proto.Sentence{Word: "!done", Map: map[string]string{}}}, nil
	}
	return run(ctx, cmd, args...)
}

// CallsSnapshot returns copy of accumulated calls.
func (c *Client) CallsSnapshot() []Call {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]Call, len(c.Calls))
	copy(out, c.Calls)
	return out
}

// Reply creates RouterOS reply from map rows.
func Reply(rows ...map[string]string) *goros.Reply {
	re := make([]*proto.Sentence, 0, len(rows))
	for _, row := range rows {
		pairs := make([]proto.Pair, 0, len(row))
		copied := make(map[string]string, len(row))
		for key, value := range row {
			pairs = append(pairs, proto.Pair{Key: key, Value: value})
			copied[key] = value
		}
		re = append(re, &proto.Sentence{Word: "!re", Map: copied, List: pairs})
	}
	return &goros.Reply{Re: re, Done: &proto.Sentence{Word: "!done", Map: map[string]string{}}}
}
