package prtriage

import (
	"context"

	"github.com/berquerant/agent-skills/mcp/internal/feature"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const featureName = "pr-triage"

var toolNames = []string{
	"pr_classify_files",
	"pr_filter_candidates",
	"pr_detect_tasks",
	"pr_match_failure",
}

// Feature registers the pr-triage MCP tools.
type Feature struct{}

// NewFeature creates a pr-triage Feature.
func NewFeature() *Feature {
	return &Feature{}
}

// Name returns the canonical feature name ("pr-triage").
func (f *Feature) Name() string {
	return featureName
}

// Tools returns the registered tool names for pr-triage.
func (f *Feature) Tools() []string {
	return toolNames
}

// Register adds the 4 pr-triage tools to the MCP server.
func (f *Feature) Register(s *mcp.Server, env feature.Env) error {
	f.registerClassificationTools(s, env)
	f.registerDetectionAndMatchingTools(s, env)
	return nil
}

func (f *Feature) registerClassificationTools(s *mcp.Server, env feature.Env) {
	feature.AddSafeTool(
		s,
		env,
		&mcp.Tool{
			Name:        "pr_classify_files",
			Description: "Inspect a list of modified file paths in a PR diff and verify strict dependency manifest exclusivity against supported ecosystems.",
		},
		func(_ context.Context, _ *mcp.CallToolRequest, in ClassifyFilesInput) (ClassifyFilesOutput, error) {
			return ClassifyFiles(in.Files)
		},
	)

	feature.AddSafeTool(
		s,
		env,
		&mcp.Tool{
			Name:        "pr_filter_candidates",
			Description: "Filter a batch list of candidate PRs (from gh pr list or glab api) by bot author, dependency labels, title conventions, and manifest exclusivity.",
		},
		func(_ context.Context, _ *mcp.CallToolRequest, in FilterCandidatesInput) (FilterCandidatesOutput, error) {
			return FilterCandidates(in.PRs)
		},
	)
}

func (f *Feature) registerDetectionAndMatchingTools(s *mcp.Server, env feature.Env) {
	feature.AddSafeTool(
		s,
		env,
		&mcp.Tool{
			Name:        "pr_detect_tasks",
			Description: "Inspect local repository configurations (Makefile, Taskfile, mise.toml, justfile, package.json, Cargo.toml, go.mod, workflows) to discover verification commands without execution.",
		},
		func(_ context.Context, _ *mcp.CallToolRequest, in DetectTasksInput) (DetectTasksOutput, error) {
			return DetectTasks(in.RepoDir)
		},
	)

	feature.AddSafeTool(
		s,
		env,
		&mcp.Tool{
			Name:        "pr_match_failure",
			Description: "Match CI build/test failure logs against 5 automated dependency failure archetypes and suggest concrete remediation playbooks.",
		},
		func(_ context.Context, _ *mcp.CallToolRequest, in MatchFailureInput) (MatchFailureOutput, error) {
			return MatchFailure(in.LogExcerpt)
		},
	)
}

// Close is a no-op for stateless pr-triage.
func (f *Feature) Close() error {
	return nil
}
