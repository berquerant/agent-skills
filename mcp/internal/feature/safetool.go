package feature

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// AddSafeTool registers a typed MCP tool with automatic panic recovery and output size bounding.
// Domain errors returned by handler are surfaced as MCP tool errors (IsError: true) rather than
// protocol errors, matching go-sdk's AddTool behavior while protecting the server process from panics.
func AddSafeTool[In, Out any](
	s *mcp.Server,
	env Env,
	tool *mcp.Tool,
	handler func(ctx context.Context, req *mcp.CallToolRequest, in In) (Out, error),
) {
	maxBytes := env.MaxOutputSize
	if maxBytes <= 0 {
		maxBytes = DefaultMaxOutputSize
	}

	wrapper := func(
		ctx context.Context,
		req *mcp.CallToolRequest,
		in In,
	) (res *mcp.CallToolResult, out Out, err error) {
		defer func() {
			if rec := recover(); rec != nil {
				if env.Logger != nil {
					env.Logger.Error("tool handler panicked",
						slog.String("tool", tool.Name),
						slog.Any("panic", rec),
					)
				}
				var zero Out
				res = nil
				out = zero
				err = fmt.Errorf("internal panic in tool %s: %v", tool.Name, rec)
			}
		}()

		out, err = handler(ctx, req, in)
		if err != nil {
			var zero Out
			return nil, zero, err
		}

		if raw, mErr := json.Marshal(out); mErr == nil && len(raw) > maxBytes {
			truncated := map[string]any{
				"ok":         false,
				"truncated":  true,
				"size_bytes": len(raw),
				"max_bytes":  maxBytes,
				"error": map[string]any{
					"code":    "output_too_large",
					"message": fmt.Sprintf("serialized output (%d bytes) exceeds limit (%d bytes)", len(raw), maxBytes),
				},
			}
			b, _ := json.Marshal(truncated)
			var zero Out
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: string(b)},
				},
			}, zero, nil
		}

		return nil, out, nil
	}

	mcp.AddTool(s, tool, wrapper)
}
