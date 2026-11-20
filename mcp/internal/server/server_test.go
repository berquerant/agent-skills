package server_test

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/berquerant/agent-skills/mcp/internal/feature"
	"github.com/berquerant/agent-skills/mcp/internal/server"
	"github.com/berquerant/agent-skills/mcp/internal/testutil"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
			wantTools:       []string{"cache_get", "dag_create", "dag_next", "lint_skill", "lint_all", "snippet_audit", "checkpoint_init", "checkpoint_append", "pace_start", "pace_tick", "pr_classify_files", "pr_filter_candidates"},
			wantAbsentTools: nil,
		},
		{
			name:            "disabling pace-guard and pr-triage omits their tools",
			extraArgs:       []string{"--pace-guard=false", "--pr-triage=false"},
			wantSummarySub:  "features=[cache,task-dag,skill-lint,snippet-analyzer,checkpoint]",
			wantTools:       []string{"cache_get", "cache_set", "dag_create", "lint_skill", "snippet_audit", "checkpoint_init"},
			wantAbsentTools: []string{"pace_start", "pace_tick", "pr_classify_files", "pr_filter_candidates"},
		},
		{
			name:            "disabling cache omits cache_* tools",
			extraArgs:       []string{"--cache=false"},
			wantSummarySub:  "features=[task-dag,skill-lint,snippet-analyzer,checkpoint,pace-guard,pr-triage]",
			wantTools:       []string{"dag_create", "lint_skill", "snippet_audit", "checkpoint_init", "pace_start", "pr_classify_files"},
			wantAbsentTools: []string{"cache_get", "cache_set", "cache_exist", "cache_delete", "cache_version", "cache_purge"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
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
			inst, err := server.New(cfg, nil, testutil.NewManualClock(fixedNow), &stderrBuf)
			require.NoError(t, err)
			defer func() { _ = inst.Close() }()

			summary := stderrBuf.String()
			assert.Contains(t, summary, "agent-skills-mcp")
			assert.Contains(t, summary, tt.wantSummarySub)

			session := testutil.ConnectInMemory(t, inst.MCPServer)

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
			verifyCheckpointClock(ctx, t, session)
		})
	}
}

func verifyCheckpointClock(ctx context.Context, t *testing.T, session *mcp.ClientSession) {
	t.Helper()
	callRes, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "checkpoint_init",
		Arguments: map[string]any{
			"project":      "test-proj",
			"branch":       "main",
			"context":      "clock-check",
			"title":        "Clock Check",
			"objective":    "Verify injected clock timestamp",
			"propose_only": true,
		},
	})
	require.NoError(t, err)
	assert.False(t, callRes.IsError)
	assert.Contains(t, testutil.FirstTextContent(t, callRes), "checkpoint-20261115-020000.md")
}

func TestSafeToolFaults(t *testing.T) {
	ctx := t.Context()
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

	session := testutil.ConnectInMemory(t, inst.MCPServer)

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
			text := testutil.FirstTextContent(t, res)
			for _, sub := range tt.wantContain {
				assert.Contains(t, text, sub)
			}
		})
	}
}

func TestServerCacheEndToEnd(t *testing.T) {
	ctx := t.Context()
	tmp := t.TempDir()
	fixedNow := time.Date(2026, 11, 15, 2, 30, 0, 0, time.UTC)

	cfg, err := feature.ParseConfig([]string{
		"--state-dir=" + tmp + "/state",
		"--cache-dir=" + tmp + "/cache",
		"--checkpoint-root=" + tmp + "/checkpoints",
		"--max-output-size=4096",
	}, tmp)
	require.NoError(t, err)

	var stderrBuf bytes.Buffer
	inst, err := server.New(cfg, nil, testutil.NewManualClock(fixedNow), &stderrBuf)
	require.NoError(t, err)
	defer func() { _ = inst.Close() }()

	session := testutil.ConnectInMemory(t, inst.MCPServer)

	// Store a 10 KiB payload (> max-output-size=4096) and verify cache_get succeeds via CacheMaxEntrySize
	largeStdout := strings.Repeat("Z", 10*1024)
	setRes, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "cache_set",
		Arguments: map[string]any{
			"type":    "command",
			"name":    "gh",
			"dir":     tmp,
			"args":    []string{"pr", "list"},
			"ttl":     300,
			"stdout":  largeStdout,
			"stderr":  "",
			"success": true,
		},
	})
	require.NoError(t, err)
	assert.False(t, setRes.IsError)

	getRes, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "cache_get",
		Arguments: map[string]any{
			"type": "command",
			"name": "gh",
			"dir":  tmp,
			"args": []string{"pr", "list"},
		},
	})
	require.NoError(t, err)
	assert.False(t, getRes.IsError)
	getText := testutil.FirstTextContent(t, getRes)
	assert.Contains(t, getText, `"hit":true`)
	assert.Contains(t, getText, largeStdout)

	verifyLintAndSnippetTools(ctx, t, session)
}

func verifyLintAndSnippetTools(ctx context.Context, t *testing.T, session *mcp.ClientSession) {
	t.Helper()
	skillDir, err := filepath.Abs(filepath.Join("..", "..", "..", "skills", "skill-review"))
	require.NoError(t, err)
	lintRes, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "lint_skill",
		Arguments: map[string]any{
			"path":   skillDir,
			"strict": true,
		},
	})
	require.NoError(t, err)
	assert.False(t, lintRes.IsError)
	lintText := testutil.FirstTextContent(t, lintRes)
	assert.Contains(t, lintText, `"passed":true`)
	assert.Contains(t, lintText, `"skill_name":"skill-review"`)

	lintAllRes, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "lint_all",
		Arguments: map[string]any{
			"root":   skillDir,
			"strict": true,
		},
	})
	require.NoError(t, err)
	assert.False(t, lintAllRes.IsError)
	assert.Contains(t, testutil.FirstTextContent(t, lintAllRes), `"total_skills":1`)

	snipRes, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "snippet_audit",
		Arguments: map[string]any{
			"code": "eval \"$CMD\"",
		},
	})
	require.NoError(t, err)
	assert.False(t, snipRes.IsError)
	assert.Contains(t, testutil.FirstTextContent(t, snipRes), `"level":4`)

	dagRes, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "dag_create",
		Arguments: map[string]any{
			"dag_id": "dag-e2e",
			"tasks": []map[string]any{
				{"id": "step-1", "title": "Setup"},
				{"id": "step-2", "title": "Verify", "depends_on": []string{"step-1"}},
			},
		},
	})
	require.NoError(t, err)
	assert.False(t, dagRes.IsError)
	assert.Contains(t, testutil.FirstTextContent(t, dagRes), `"dag_id":"dag-e2e"`)
}
