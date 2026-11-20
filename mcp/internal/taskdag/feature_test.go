package taskdag_test

import (
	"context"
	"testing"

	"github.com/berquerant/agent-skills/mcp/internal/feature"
	"github.com/berquerant/agent-skills/mcp/internal/taskdag"
	"github.com/berquerant/agent-skills/mcp/internal/testutil"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestSession(t *testing.T) (context.Context, *mcp.ClientSession) {
	t.Helper()
	tmp := t.TempDir()

	f := taskdag.NewFeatureWithDir(tmp)
	assert.Equal(t, "task-dag", f.Name())
	assert.Len(t, f.Tools(), 7)

	mcpSrv := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0.1.0"}, nil)
	env := feature.Env{
		StateDir: tmp,
		Clock:    feature.SystemClock{},
	}
	require.NoError(t, f.Register(mcpSrv, env))
	t.Cleanup(func() { _ = f.Close() })

	return t.Context(), testutil.ConnectInMemory(t, mcpSrv)
}

func TestFeatureToolExecution(t *testing.T) {
	ctx, session := setupTestSession(t)

	steps := []struct {
		name string
		args map[string]any
	}{
		{
			name: "dag_create",
			args: map[string]any{
				"dag_id": "test-dag",
				"name":   "Feature Pipeline",
				"tasks": []map[string]any{
					{"id": "build", "title": "Build app", "estimate_sec": 30},
					{"id": "test", "title": "Run tests", "estimate_sec": 60, "depends_on": []string{"build"}},
				},
			},
		},
		{
			name: "dag_next",
			args: map[string]any{"dag_id": "test-dag"},
		},
		{
			name: "dag_transition",
			args: map[string]any{
				"dag_id":  "test-dag",
				"task_id": "build",
				"to":      "RUNNING",
			},
		},
		{
			name: "dag_status",
			args: map[string]any{"dag_id": "test-dag"},
		},
		{
			name: "dag_render",
			args: map[string]any{"dag_id": "test-dag"},
		},
		{
			name: "dag_list",
			args: map[string]any{},
		},
		{
			name: "dag_delete",
			args: map[string]any{"dag_id": "test-dag"},
		},
	}

	for _, step := range steps {
		t.Run(step.name, func(t *testing.T) {
			res, err := session.CallTool(ctx, &mcp.CallToolParams{
				Name:      step.name,
				Arguments: step.args,
			})
			require.NoError(t, err)
			assert.False(t, res.IsError)
		})
	}
}
