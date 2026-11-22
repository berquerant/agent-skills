package snippetmgr

import (
	"context"

	"github.com/berquerant/agent-skills/mcp/internal/feature"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const featureName = "snippet-manager"

var toolNames = []string{
	"snippet_check_inline",
	"snippet_create",
	"snippet_get",
	"snippet_list",
	"snippet_verify",
	"snippet_record_invocation",
	"snippet_deprecate",
	"snippet_cleanup",
}

// Feature registers the snippet-manager MCP tools.
type Feature struct{}

// NewFeature creates a snippet-manager Feature instance.
func NewFeature() *Feature {
	return &Feature{}
}

// Name returns the canonical feature name ("snippet-manager").
func (f *Feature) Name() string {
	return featureName
}

// Tools returns the registered tool names for snippet-manager.
func (f *Feature) Tools() []string {
	return toolNames
}

// Register adds all snippet-manager tools to the MCP server.
func (f *Feature) Register(s *mcp.Server, env feature.Env) error {
	mgr, err := NewManager(env.SnippetDir)
	if err != nil {
		return err
	}
	clock := env.Clock
	if clock == nil {
		clock = feature.SystemClock{}
	}

	feature.AddSafeTool(s, env, &mcp.Tool{
		Name:        "snippet_check_inline",
		Description: "Evaluate whether an inline shell command exceeds command-snippet avoidance thresholds and should be extracted into a snippet file.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in CheckInlineInput) (CheckInlineOutput, error) {
		return CheckInline(in), nil
	})

	feature.AddSafeTool(s, env, &mcp.Tool{
		Name:        "snippet_create",
		Description: "Validate, audit, and save an immutable versioned snippet (mode 0500) under the XDG snippet root, auto-deprecating prior versions and returning a Pre-Execution Notice skeleton.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in CreateInput) (CreateOutput, error) {
		return mgr.Create(in, clock.Now())
	})

	feature.AddSafeTool(s, env, &mcp.Tool{
		Name:        "snippet_get",
		Description: "Retrieve metadata, version history, audit details, and invocation records for a managed snippet.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in GetInput) (GetOutput, error) {
		return mgr.Get(in)
	})

	feature.AddSafeTool(s, env, &mcp.Tool{
		Name:        "snippet_list",
		Description: "List managed snippets filtered by project, branch, scope, or status.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in ListInput) (ListOutput, error) {
		return mgr.List(in)
	})

	feature.AddSafeTool(s, env, &mcp.Tool{
		Name:        "snippet_verify",
		Description: "Verify a snippet version is active, executable, write-protected, and matches its recorded SHA-256 digest before host execution.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in VerifyInput) (VerifyOutput, error) {
		return mgr.Verify(in)
	})

	feature.AddSafeTool(s, env, &mcp.Tool{
		Name:        "snippet_record_invocation",
		Description: "Record a host-side snippet execution (version, arguments, exit code, duration, note) in the snippet metadata ledger.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in RecordInvocationInput) (RecordInvocationOutput, error) {
		return mgr.RecordInvocation(in, clock.Now())
	})

	feature.AddSafeTool(s, env, &mcp.Tool{
		Name:        "snippet_deprecate",
		Description: "Mark one or all active versions of a snippet as deprecated and strip execute permissions (mode 0400).",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in DeprecateInput) (DeprecateOutput, error) {
		return mgr.Deprecate(in, clock.Now())
	})

	feature.AddSafeTool(s, env, &mcp.Tool{
		Name:        "snippet_cleanup",
		Description: "Remove deprecated (and optionally active one-off) snippet files while preserving metadata history.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in CleanupInput) (CleanupOutput, error) {
		return mgr.Cleanup(in, clock.Now())
	})

	return nil
}

// Close is a no-op because snippet-manager holds no background resources.
func (f *Feature) Close() error {
	return nil
}
