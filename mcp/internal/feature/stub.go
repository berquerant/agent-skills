package feature

import (
	"context"
	"fmt"
	"slices"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// StubFeature provides a minimal feature implementation for Phase 0 skeleton and feature-flag testing.
// As each phase (1..7) is implemented, its corresponding stub is replaced by the real feature package.
type StubFeature struct {
	FeatureName string
	ToolVerbs   []string
}

// PingParams defines the input for a diagnostic ping tool.
type PingParams struct {
	Message string `json:"message,omitempty" jsonschema:"optional echo message"`
}

// PingResult defines the output of a diagnostic ping tool.
type PingResult struct {
	OK        bool   `json:"ok" jsonschema:"true when feature is active"`
	Feature   string `json:"feature" jsonschema:"feature module name"`
	Echo      string `json:"echo,omitempty" jsonschema:"echoed message if provided"`
	Timestamp int64  `json:"timestamp" jsonschema:"current server unix timestamp"`
}

func (f *StubFeature) Name() string {
	return f.FeatureName
}

func (f *StubFeature) Tools() []string {
	return slices.Clone(f.ToolVerbs)
}

func (f *StubFeature) Register(s *mcp.Server, env Env) error {
	clock := env.Clock
	if clock == nil {
		clock = SystemClock{}
	}
	for _, toolName := range f.Tools() {
		tName := toolName
		tool := &mcp.Tool{
			Name:        tName,
			Description: fmt.Sprintf("Diagnostic stub tool for %s feature (%s).", f.FeatureName, tName),
		}
		AddSafeTool(s, env, tool, func(_ context.Context, _ *mcp.CallToolRequest, params PingParams) (PingResult, error) {
			return PingResult{
				OK:        true,
				Feature:   f.FeatureName,
				Echo:      params.Message,
				Timestamp: clock.Now().Unix(),
			}, nil
		})
	}
	return nil
}

func (f *StubFeature) Close() error {
	return nil
}

// RemainingStubFeatures returns the stub implementations for features not yet replaced by real packages.
func RemainingStubFeatures() []Feature {
	return []Feature{
		&StubFeature{
			FeatureName: "task-dag",
			ToolVerbs:   []string{"dag_ping"},
		},
		&StubFeature{
			FeatureName: "snippet-analyzer",
			ToolVerbs:   []string{"snippet_ping"},
		},
		&StubFeature{
			FeatureName: "checkpoint",
			ToolVerbs:   []string{"checkpoint_ping"},
		},
		&StubFeature{
			FeatureName: "pace-guard",
			ToolVerbs:   []string{"pace_ping"},
		},
		&StubFeature{
			FeatureName: "pr-triage",
			ToolVerbs:   []string{"pr_triage_ping"},
		},
	}
}
