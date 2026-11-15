package checkpoint

import (
	"context"

	"github.com/berquerant/agent-skills/mcp/internal/feature"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const featureName = "checkpoint"

var toolNames = []string{
	"checkpoint_init",
	"checkpoint_append",
	"checkpoint_resume",
	"checkpoint_set_status",
	"checkpoint_list",
}

// Feature registers the checkpoint MCP tools.
type Feature struct {
	rootDir string
	mgr     *Manager
	clock   feature.Clock
}

// NewFeature creates a checkpoint Feature.
func NewFeature() *Feature {
	return &Feature{}
}

// NewFeatureWithDir creates a checkpoint Feature with a custom root directory.
func NewFeatureWithDir(rootDir string) *Feature {
	return &Feature{rootDir: rootDir}
}

// Name returns the canonical feature name ("checkpoint").
func (f *Feature) Name() string {
	return featureName
}

// Tools returns the registered tool names for checkpoint.
func (f *Feature) Tools() []string {
	return toolNames
}

// Register initializes the checkpoint Manager under env.CheckpointRoot and registers the 5 checkpoint tools.
func (f *Feature) Register(s *mcp.Server, env feature.Env) error {
	clock := env.Clock
	if clock == nil {
		clock = feature.SystemClock{}
	}
	rootDir := f.rootDir
	if rootDir == "" {
		rootDir = env.CheckpointDir
	}
	mgr, err := NewManager(rootDir)
	if err != nil {
		return err
	}
	f.mgr = mgr
	f.clock = clock

	f.registerInitAndAppend(s, env)
	f.registerResumeSetStatusList(s, env)
	return nil
}

func (f *Feature) registerInitAndAppend(s *mcp.Server, env feature.Env) {
	feature.AddSafeTool(
		s,
		env,
		&mcp.Tool{
			Name:        "checkpoint_init",
			Description: "Resolve a context-scoped checkpoint log path and initialize its Markdown header block (set propose_only=true to preview path before user confirmation).",
		},
		func(_ context.Context, _ *mcp.CallToolRequest, in InitInput) (InitOutput, error) {
			return f.mgr.Init(in, f.clock.Now())
		},
	)

	feature.AddSafeTool(
		s,
		env,
		&mcp.Tool{
			Name:        "checkpoint_append",
			Description: "Append a numbered milestone step entry to an existing checkpoint log file using append-only file operations.",
		},
		func(_ context.Context, _ *mcp.CallToolRequest, in AppendInput) (AppendOutput, error) {
			return f.mgr.Append(in, f.clock.Now())
		},
	)
}

func (f *Feature) registerResumeSetStatusList(s *mcp.Server, env feature.Env) {
	feature.AddSafeTool(
		s,
		env,
		&mcp.Tool{
			Name:        "checkpoint_resume",
			Description: "Locate and parse a checkpoint log (by path or latest generation for context) to recover completed milestones, unresolved steps, and next actions.",
		},
		func(_ context.Context, _ *mcp.CallToolRequest, in ResumeInput) (ResumeOutput, error) {
			return f.mgr.Resume(in)
		},
	)

	feature.AddSafeTool(
		s,
		env,
		&mcp.Tool{
			Name:        "checkpoint_set_status",
			Description: "Update the overall task Status field in a checkpoint log header (IN_PROGRESS, COMPLETED, or ABORTED).",
		},
		func(_ context.Context, _ *mcp.CallToolRequest, in SetStatusInput) (SetStatusOutput, error) {
			return f.mgr.SetStatus(in)
		},
	)

	feature.AddSafeTool(
		s,
		env,
		&mcp.Tool{
			Name:        "checkpoint_list",
			Description: "List checkpoint, runbook, and report files under the checkpoint root with optional project, branch, context, and doc_type filters.",
		},
		func(_ context.Context, _ *mcp.CallToolRequest, in ListInput) (ListOutput, error) {
			return f.mgr.List(in)
		},
	)
}

// Close is a no-op because file locks are acquired per operation.
func (f *Feature) Close() error {
	return nil
}
