package feature

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/berquerant/agent-skills/mcp/internal/guard"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// AddSafeTool registers a typed MCP tool with automatic panic recovery, server-wide tool call logging,
// and output size bounding using env.MaxOutputSize.
func AddSafeTool[In, Out any](
	s *mcp.Server,
	env Env,
	tool *mcp.Tool,
	handler func(ctx context.Context, req *mcp.CallToolRequest, in In) (Out, error),
) {
	AddSafeToolWithMax(s, env, env.MaxOutputSize, tool, handler)
}

// AddSafeToolWithMax registers a typed MCP tool with automatic panic recovery, server-wide tool call
// logging via env.Logger, and a custom maximum output size limit in bytes.
func AddSafeToolWithMax[In, Out any](
	s *mcp.Server,
	env Env,
	maxBytes int,
	tool *mcp.Tool,
	handler func(ctx context.Context, req *mcp.CallToolRequest, in In) (Out, error),
) {
	if maxBytes <= 0 {
		maxBytes = env.MaxOutputSize
	}
	if maxBytes <= 0 {
		maxBytes = DefaultMaxOutputSize
	}

	wrapper := func(
		ctx context.Context,
		req *mcp.CallToolRequest,
		in In,
	) (res *mcp.CallToolResult, out Out, err error) {
		start := time.Now()
		defer func() {
			durationMS := time.Since(start).Milliseconds()
			if rec := recover(); rec != nil {
				if env.Logger != nil {
					env.Logger.Error("tool handler panicked",
						slog.String("tool", tool.Name),
						slog.Int64("duration_ms", durationMS),
						slog.Any("panic", rec),
					)
				}
				var zero Out
				res = nil
				out = zero
				err = fmt.Errorf("internal panic in tool %s: %v", tool.Name, rec)
				return
			}
			logToolCall(env.Logger, tool.Name, durationMS, in, out, res, err)
		}()

		out, err = handler(ctx, req, in)
		if err != nil {
			var zero Out
			return nil, zero, err
		}

		if raw, mErr := json.Marshal(out); mErr == nil && len(raw) > maxBytes {
			return buildTruncatedResult[Out](len(raw), maxBytes)
		}

		return nil, out, nil
	}

	mcp.AddTool(s, tool, wrapper)
}

func logToolCall[In, Out any](
	logger *slog.Logger,
	toolName string,
	durationMS int64,
	in In,
	out Out,
	res *mcp.CallToolResult,
	err error,
) {
	if logger == nil {
		return
	}
	isErr := err != nil || (res != nil && res.IsError)
	logger.Info("tool call",
		slog.String("tool", toolName),
		slog.Int64("duration_ms", durationMS),
		slog.Bool("is_error", isErr),
	)
	var errText any
	if err != nil {
		errText = guard.RedactSecretsInString(err.Error())
	}
	logger.Debug("tool call payload",
		slog.String("tool", toolName),
		slog.Any("input", guard.SanitizeForLog(in)),
		slog.Any("output", guard.SanitizeForLog(out)),
		slog.Any("error", errText),
	)
}

func buildTruncatedResult[Out any](sizeBytes, maxBytes int) (*mcp.CallToolResult, Out, error) {
	truncated := map[string]any{
		"ok":         false,
		"truncated":  true,
		"size_bytes": sizeBytes,
		"max_bytes":  maxBytes,
		"error": map[string]any{
			"code":    "output_too_large",
			"message": fmt.Sprintf("serialized output (%d bytes) exceeds limit (%d bytes)", sizeBytes, maxBytes),
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
