package testutil_test

import (
	"context"
	"testing"
	"time"

	"github.com/berquerant/agent-skills/mcp/internal/testutil"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManualClock(t *testing.T) {
	base := time.Date(2026, 11, 15, 4, 0, 0, 0, time.UTC)
	clk := testutil.NewManualClock(base)
	assert.Equal(t, base, clk.Now())

	clk.Advance(90 * time.Second)
	assert.Equal(t, base.Add(90*time.Second), clk.Now())

	unixClk := testutil.NewManualClockUnix(1000)
	assert.Equal(t, int64(1000), unixClk.Now().Unix())
}

func TestConnectInMemoryAndFirstTextContent(t *testing.T) {
	srv := mcp.NewServer(&mcp.Implementation{Name: "echo-test", Version: "0.1.0"}, nil)
	mcp.AddTool(srv, &mcp.Tool{Name: "ping", Description: "echo ping"}, func(
		_ context.Context,
		_ *mcp.CallToolRequest,
		_ struct{},
	) (*mcp.CallToolResult, struct{ Pong bool }, error) {
		return nil, struct{ Pong bool }{Pong: true}, nil
	})

	session := testutil.ConnectInMemory(t, srv)
	res, err := session.CallTool(t.Context(), &mcp.CallToolParams{Name: "ping"})
	require.NoError(t, err)
	assert.False(t, res.IsError)
	assert.Contains(t, testutil.FirstTextContent(t, res), `"Pong":true`)
}
