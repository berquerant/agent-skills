package skilllint

import (
	"context"
	"fmt"

	"github.com/berquerant/agent-skills/mcp/internal/feature"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const featureName = "skill-lint"

var toolNames = []string{
	"lint_skill",
	"lint_all",
}

// Feature registers the skill-lint MCP tools (lint_skill, lint_all).
type Feature struct{}

// NewFeature creates a skill-lint Feature instance.
func NewFeature() *Feature {
	return &Feature{}
}

// Name returns the canonical feature name ("skill-lint").
func (f *Feature) Name() string {
	return featureName
}

// Tools returns the registered tool names for skill-lint.
func (f *Feature) Tools() []string {
	return toolNames
}

// Register adds lint_skill and lint_all to the MCP server.
func (f *Feature) Register(s *mcp.Server, env feature.Env) error {
	skillMaxBytes := max(env.MaxOutputSize, 64*1024)
	batchMaxBytes := max(env.MaxOutputSize, 256*1024)

	feature.AddSafeToolWithMax(
		s,
		env,
		skillMaxBytes,
		&mcp.Tool{
			Name:        "lint_skill",
			Description: "Run automated quality checks A through K on a target agent skill directory.",
		},
		func(_ context.Context, _ *mcp.CallToolRequest, in LintSkillInput) (SkillReport, error) {
			if in.Path == "" {
				return SkillReport{}, fmt.Errorf("path is required")
			}
			return LintSkill(in.Path, in.Strict)
		},
	)

	feature.AddSafeToolWithMax(
		s,
		env,
		batchMaxBytes,
		&mcp.Tool{
			Name:        "lint_all",
			Description: "Discover and run automated quality checks A through K on all skills under a repository or skills directory.",
		},
		func(_ context.Context, _ *mcp.CallToolRequest, in LintAllInput) (BatchReport, error) {
			if in.Root == "" {
				return BatchReport{}, fmt.Errorf("root is required")
			}
			return LintAll(in.Root, in.Strict)
		},
	)

	return nil
}

// Close is a no-op because skill-lint is stateless.
func (f *Feature) Close() error {
	return nil
}
