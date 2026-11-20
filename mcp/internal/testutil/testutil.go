// Package testutil provides shared test doubles and in-memory MCP transport helpers
// for unit and integration tests across internal packages.
package testutil

import (
	"sync"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

// ManualClock is a thread-safe deterministic feature.Clock for tests.
type ManualClock struct {
	mu sync.Mutex
	t  time.Time
}

// NewManualClock returns a ManualClock initialized to t.
func NewManualClock(t time.Time) *ManualClock {
	return &ManualClock{t: t}
}

// NewManualClockUnix returns a ManualClock initialized to unixSec UTC.
func NewManualClockUnix(unixSec int64) *ManualClock {
	return &ManualClock{t: time.Unix(unixSec, 0).UTC()}
}

// Now returns the clock's current time.
func (c *ManualClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.t
}

// Advance shifts the clock forward by d.
func (c *ManualClock) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.t = c.t.Add(d)
}

// ConnectInMemory starts srv on an in-memory MCP transport bound to t.Context(),
// connects a test client session, and registers cleanup to close the session and wait
// for the server loop to exit.
func ConnectInMemory(t testing.TB, srv *mcp.Server) *mcp.ClientSession {
	t.Helper()
	ctx := t.Context()
	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = srv.Run(ctx, serverTransport)
	}()

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "1.0.0"}, nil)
	session, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = session.Close()
		<-done
	})
	return session
}

// FirstTextContent extracts the text payload from the first Content item of res.
func FirstTextContent(t testing.TB, res *mcp.CallToolResult) string {
	t.Helper()
	require.NotEmpty(t, res.Content)
	tc, ok := res.Content[0].(*mcp.TextContent)
	require.True(t, ok, "expected *mcp.TextContent in tool response")
	return tc.Text
}
