package pace

import (
	"context"
	"path/filepath"

	"github.com/berquerant/agent-skills/mcp/internal/feature"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const featureName = "pace-guard"

var toolNames = []string{
	"pace_start",
	"pace_tick",
	"pace_status",
	"pace_reset",
	"pace_watch",
	"pace_liveness",
}

// Feature registers the pace-guard MCP tools.
type Feature struct {
	tracker *Tracker
	clock   feature.Clock
}

// NewFeature creates a pace-guard Feature.
func NewFeature() *Feature {
	return &Feature{}
}

// Name returns the canonical feature name ("pace-guard").
func (f *Feature) Name() string {
	return featureName
}

// Tools returns the registered tool names for pace-guard.
func (f *Feature) Tools() []string {
	return toolNames
}

// Register initializes the pace Tracker under env.StateDir and registers the 6 pace-guard tools.
func (f *Feature) Register(s *mcp.Server, env feature.Env) error {
	clock := env.Clock
	if clock == nil {
		clock = feature.SystemClock{}
	}
	tr, err := NewTracker(filepath.Join(env.StateDir, featureName))
	if err != nil {
		return err
	}
	f.tracker = tr
	f.clock = clock

	f.registerSessionTools(s, env)
	f.registerWatchTools(s, env)
	return nil
}

func (f *Feature) registerSessionTools(s *mcp.Server, env feature.Env) {
	feature.AddSafeTool(
		s,
		env,
		&mcp.Tool{
			Name:        "pace_start",
			Description: "Start an edit-pacing session with complexity presets (quick, standard, deep), command execution modes, and a user proposal statement.",
		},
		func(_ context.Context, _ *mcp.CallToolRequest, in StartInput) (StartOutput, error) {
			return f.tracker.Start(in, f.clock.Now())
		},
	)

	feature.AddSafeTool(
		s,
		env,
		&mcp.Tool{
			Name:        "pace_tick",
			Description: "Record an execution step (edit, non_edit, or parallel_batch), resetting pacing baselines on file edits or returning EXCEEDED with a checkpoint report when limits are reached.",
		},
		func(_ context.Context, _ *mcp.CallToolRequest, in TickInput) (PaceEvaluation, error) {
			return f.tracker.Tick(in, f.clock.Now())
		},
	)

	feature.AddSafeTool(
		s,
		env,
		&mcp.Tool{
			Name:        "pace_status",
			Description: "Query real wall-clock elapsed time and remaining non-edit tool calls for an edit-pacing session without modifying counters.",
		},
		func(_ context.Context, _ *mcp.CallToolRequest, in StatusInput) (PaceEvaluation, error) {
			return f.tracker.Status(in, f.clock.Now())
		},
	)

	feature.AddSafeTool(
		s,
		env,
		&mcp.Tool{
			Name:        "pace_reset",
			Description: "Reset pacing timers and non-edit counters after user direction, optionally updating thresholds or execution mode.",
		},
		func(_ context.Context, _ *mcp.CallToolRequest, in ResetInput) (PaceEvaluation, error) {
			return f.tracker.Reset(in, f.clock.Now())
		},
	)
}

func (f *Feature) registerWatchTools(s *mcp.Server, env feature.Env) {
	feature.AddSafeTool(
		s,
		env,
		&mcp.Tool{
			Name:        "pace_watch",
			Description: "Register a long-running or background command/process (by PID, log_path, or artifact_path) for liveness monitoring.",
		},
		func(_ context.Context, _ *mcp.CallToolRequest, in WatchInput) (WatchOutput, error) {
			return f.tracker.Watch(in, f.clock.Now())
		},
	)

	feature.AddSafeTool(
		s,
		env,
		&mcp.Tool{
			Name:        "pace_liveness",
			Description: "Sample CPU/memory and log/artifact activity for a watched command, classifying state as ACTIVE, STALLED, ZOMBIE, or EXITED.",
		},
		func(_ context.Context, _ *mcp.CallToolRequest, in LivenessInput) (LivenessOutput, error) {
			return f.tracker.CheckLiveness(in, f.clock.Now())
		},
	)
}

// Close is a no-op because file locks are per operation.
func (f *Feature) Close() error {
	return nil
}
