package snippet

import (
	"context"
	"fmt"

	"github.com/berquerant/agent-skills/mcp/internal/feature"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const featureName = "snippet-analyzer"

var toolNames = []string{
	"snippet_audit",
}

// Feature registers the snippet-analyzer MCP tool (snippet_audit).
type Feature struct{}

// NewFeature creates a snippet-analyzer Feature instance.
func NewFeature() *Feature {
	return &Feature{}
}

// Name returns the canonical feature name ("snippet-analyzer").
func (f *Feature) Name() string {
	return featureName
}

// Tools returns the registered tool names for snippet-analyzer.
func (f *Feature) Tools() []string {
	return toolNames
}

// Register adds snippet_audit to the MCP server.
func (f *Feature) Register(s *mcp.Server, env feature.Env) error {
	feature.AddSafeTool(
		s,
		env,
		&mcp.Tool{
			Name:        "snippet_audit",
			Description: "Audit an ad-hoc shell or Python snippet (inline code or file path) across the 5 intelligibility dimensions and assign Level 1-4.",
		},
		func(_ context.Context, _ *mcp.CallToolRequest, in AuditInput) (AuditReport, error) {
			if in.Code == "" && in.Path == "" {
				return AuditReport{}, fmt.Errorf("either code or path must be provided")
			}
			if in.Code != "" && in.Path != "" {
				return AuditReport{}, fmt.Errorf("provide either code or path, not both")
			}
			if in.Path != "" {
				return AnalyzeFile(in.Path, in.Language)
			}
			rep := Analyze(in.Code, in.Language)
			if !rep.OK {
				return rep, fmt.Errorf("snippet analysis failed: %s", rep.ParseError)
			}
			return rep, nil
		},
	)
	return nil
}

// Close is a no-op because snippet-analyzer is stateless.
func (f *Feature) Close() error {
	return nil
}
