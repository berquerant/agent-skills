package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/berquerant/agent-skills/mcp/internal/feature"
	"github.com/berquerant/agent-skills/mcp/internal/server"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubClock struct {
	t time.Time
}

func (c stubClock) Now() time.Time { return c.t }

type faultFeature struct{}

func (faultFeature) Name() string    { return "fault" }
func (faultFeature) Tools() []string { return []string{"fault_panic", "fault_huge"} }
func (faultFeature) Register(s *mcp.Server, env feature.Env) error {
	feature.AddSafeTool(s, env, &mcp.Tool{
		Name:        "fault_panic",
		Description: "triggers a panic to verify recovery",
	}, func(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (struct{ OK bool }, error) {
		panic("simulated tool panic")
	})

	feature.AddSafeTool(s, env, &mcp.Tool{
		Name:        "fault_huge",
		Description: "returns an oversized payload to verify size bounding",
	}, func(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (struct{ Payload string }, error) {
		return struct{ Payload string }{Payload: strings.Repeat("A", 1024)}, nil
	})
	return nil
}
func (faultFeature) Close() error { return nil }

func TestServerFeatureToggles(t *testing.T) {
	fixedNow := time.Date(2026, 11, 15, 2, 0, 0, 0, time.UTC)

	tests := []struct {
		name            string
		extraArgs       []string
		wantSummarySub  string
		wantTools       []string
		wantAbsentTools []string
	}{
		{
			name:            "all features enabled by default",
			wantSummarySub:  "features=[cache,task-dag,skill-lint,snippet-analyzer,checkpoint,pace-guard,pr-triage]",
			wantTools:       []string{"cache_get", "dag_ping", "skill_lint_ping", "snippet_ping", "checkpoint_ping", "pace_ping", "pr_triage_ping"},
			wantAbsentTools: nil,
		},
		{
			name:            "disabling pace-guard and pr-triage omits their tools",
			extraArgs:       []string{"--pace-guard=false", "--pr-triage=false"},
			wantSummarySub:  "features=[cache,task-dag,skill-lint,snippet-analyzer,checkpoint]",
			wantTools:       []string{"cache_get", "cache_set", "dag_ping", "skill_lint_ping", "snippet_ping", "checkpoint_ping"},
			wantAbsentTools: []string{"pace_ping", "pr_triage_ping"},
		},
		{
			name:            "unprefixed cache tools when --cache-tool-prefix is empty",
			extraArgs:       []string{"--cache-tool-prefix="},
			wantSummarySub:  "tools=[get,set,exist,delete,version,purge",
			wantTools:       []string{"get", "set", "exist", "delete", "version", "purge"},
			wantAbsentTools: []string{"cache_get"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			tmp := t.TempDir()

			baseArgs := []string{
				"--state-dir=" + tmp + "/state",
				"--cache-dir=" + tmp + "/cache",
				"--checkpoint-root=" + tmp + "/checkpoints",
			}
			args := append(baseArgs, tt.extraArgs...)
			cfg, err := feature.ParseConfig(args, tmp)
			require.NoError(t, err)

			var stderrBuf bytes.Buffer
			inst, err := server.New(cfg, nil, stubClock{t: fixedNow}, &stderrBuf)
			require.NoError(t, err)
			defer func() { _ = inst.Close() }()

			summary := stderrBuf.String()
			assert.Contains(t, summary, "agent-skills-mcp")
			assert.Contains(t, summary, tt.wantSummarySub)

			serverTransport, clientTransport := mcp.NewInMemoryTransports()
			go func() {
				_ = inst.MCPServer.Run(ctx, serverTransport)
			}()

			client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "1.0.0"}, nil)
			session, err := client.Connect(ctx, clientTransport, nil)
			require.NoError(t, err)
			defer func() { _ = session.Close() }()

			listRes, err := session.ListTools(ctx, nil)
			require.NoError(t, err)

			toolNames := make([]string, 0, len(listRes.Tools))
			for _, tool := range listRes.Tools {
				toolNames = append(toolNames, tool.Name)
			}

			for _, w := range tt.wantTools {
				assert.Contains(t, toolNames, w)
			}
			for _, a := range tt.wantAbsentTools {
				assert.NotContains(t, toolNames, a)
			}

			// Verify calling an active tool uses the injected clock
			callRes, err := session.CallTool(ctx, &mcp.CallToolParams{
				Name:      "dag_ping",
				Arguments: map[string]any{"message": "hello-dag"},
			})
			require.NoError(t, err)
			assert.False(t, callRes.IsError)
			require.NotEmpty(t, callRes.Content)
			tc, ok := callRes.Content[0].(*mcp.TextContent)
			require.True(t, ok)

			var pingOut feature.PingResult
			require.NoError(t, json.Unmarshal([]byte(tc.Text), &pingOut))
			assert.True(t, pingOut.OK)
			assert.Equal(t, "task-dag", pingOut.Feature)
			assert.Equal(t, "hello-dag", pingOut.Echo)
			assert.Equal(t, fixedNow.Unix(), pingOut.Timestamp)
		})
	}
}

func TestSafeToolFaults(t *testing.T) {
	ctx := context.Background()
	tmp := t.TempDir()

	reg := feature.NewRegistry()
	require.NoError(t, reg.Add(faultFeature{}))

	cfg := &feature.Config{
		Enabled:       map[string]bool{"fault": true},
		StateDir:      tmp,
		CacheDir:      tmp,
		MaxOutputSize: 256,
		LogLevel:      "error",
	}

	var stderrBuf bytes.Buffer
	inst, err := server.New(cfg, reg, nil, &stderrBuf)
	require.NoError(t, err)
	defer func() { _ = inst.Close() }()

	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	go func() {
		_ = inst.MCPServer.Run(ctx, serverTransport)
	}()

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "1.0.0"}, nil)
	session, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	defer func() { _ = session.Close() }()

	tests := []struct {
		name        string
		toolName    string
		wantContain []string
	}{
		{
			name:        "panic in tool handler returns tool error without crashing server",
			toolName:    "fault_panic",
			wantContain: []string{"internal panic in tool fault_panic"},
		},
		{
			name:        "oversized output is truncated and flagged",
			toolName:    "fault_huge",
			wantContain: []string{`"truncated":true`, `"output_too_large"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := session.CallTool(ctx, &mcp.CallToolParams{
				Name: tt.toolName,
			})
			require.NoError(t, err)
			assert.True(t, res.IsError)
			require.NotEmpty(t, res.Content)
			tc, ok := res.Content[0].(*mcp.TextContent)
			require.True(t, ok)
			for _, sub := range tt.wantContain {
				assert.Contains(t, tc.Text, sub)
			}
		})
	}
}
