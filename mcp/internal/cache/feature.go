package cache

import (
	"context"

	"github.com/berquerant/agent-skills/mcp/internal/feature"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const getOutputEnvelopeAllowance = 16384

// Feature implements feature.Feature for the cache module.
type Feature struct {
	svc *Service
}

// NewFeature creates a cache Feature.
func NewFeature() *Feature {
	return &Feature{}
}

// Name returns the canonical feature name ("cache").
func (f *Feature) Name() string {
	return "cache"
}

// Tools returns the 6 cache tool names.
func (f *Feature) Tools() []string {
	return []string{
		"cache_get",
		"cache_set",
		"cache_exist",
		"cache_delete",
		"cache_version",
		"cache_purge",
	}
}

// Register initializes the cache service and registers the 6 cache MCP tools on s.
func (f *Feature) Register(s *mcp.Server, env feature.Env) error {
	cfg := Config{
		Dir:                         env.CacheDir,
		MaxTTL:                      env.CacheMaxTTL,
		MaxCacheSize:                env.CacheMaxEntrySize,
		MaxTotalCompressedCacheSize: env.CacheMaxTotalSize,
		Clock:                       env.Clock,
	}
	svc, err := NewService(cfg)
	if err != nil {
		return err
	}
	f.svc = svc

	getMaxBytes := max(int(cfg.MaxCacheSize)+getOutputEnvelopeAllowance, env.MaxOutputSize)

	feature.AddSafeToolWithMax(s, env, getMaxBytes, &mcp.Tool{
		Name:        "cache_get",
		Description: "Retrieve cached execution record by key or invocation parameters (type, name, dir, args, stdin).",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in QueryInput) (map[string]any, error) {
		return svc.Get(in), nil
	})

	feature.AddSafeTool(s, env, &mcp.Tool{
		Name:        "cache_set",
		Description: "Store command or MCP tool execution result as a new immutable cache version.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in SetInput) (map[string]any, error) {
		return svc.Set(in), nil
	})

	feature.AddSafeTool(s, env, &mcp.Tool{
		Name:        "cache_exist",
		Description: "Check whether a valid (unexpired) cache entry exists without decompressing contents.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in QueryInput) (map[string]any, error) {
		return svc.Exist(in), nil
	})

	feature.AddSafeTool(s, env, &mcp.Tool{
		Name:        "cache_delete",
		Description: "Delete cache entries for a key (all versions, 'latest', or a specific version).",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in QueryInput) (map[string]any, error) {
		return svc.Delete(in), nil
	})

	feature.AddSafeTool(s, env, &mcp.Tool{
		Name:        "cache_version",
		Description: "List all existing, non-expired versions for a cache key without decompressing contents.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in VersionInput) (map[string]any, error) {
		return svc.Version(in), nil
	})

	feature.AddSafeTool(s, env, &mcp.Tool{
		Name:        "cache_purge",
		Description: "Purge cache files across all shard subdirectories (expired only by default, or all when all=true).",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in PurgeInput) (map[string]any, error) {
		return svc.Purge(in), nil
	})

	return nil
}

// Close releases underlying storage resources.
func (f *Feature) Close() error {
	if f.svc != nil {
		return f.svc.Close()
	}
	return nil
}
