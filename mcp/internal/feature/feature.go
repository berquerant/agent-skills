package feature

import (
	"fmt"
	"log/slog"
	"slices"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Clock provides time operations so time-dependent features can be tested deterministically.
type Clock interface {
	Now() time.Time
}

// SystemClock implements Clock using standard time.Now.
type SystemClock struct{}

func (SystemClock) Now() time.Time { return time.Now() }

// Env holds runtime dependencies and configuration passed to features at registration.
type Env struct {
	Logger        *slog.Logger
	Clock         Clock
	StateDir      string
	CacheDir      string
	CheckpointDir string
	MaxOutputSize int

	CacheMaxTTL       int
	CacheMaxEntrySize int64
	CacheMaxTotalSize int64
}

// DefaultMaxOutputSize is the default cap on serialized tool response size (32 KiB).
const DefaultMaxOutputSize = 32 * 1024

// Feature represents a toggleable capability in agent-skills-mcp.
type Feature interface {
	// Name returns the unique feature flag name (e.g. "cache", "task-dag").
	Name() string
	// Tools returns the names of the MCP tools this feature registers.
	Tools() []string
	// Register registers the feature's tools with the MCP server.
	Register(s *mcp.Server, env Env) error
	// Close releases any background resources held by the feature.
	Close() error
}

// Registry manages available features and activates enabled ones.
type Registry struct {
	order    []string
	features map[string]Feature
}

// NewRegistry creates an empty Feature registry.
func NewRegistry() *Registry {
	return &Registry{
		features: make(map[string]Feature),
	}
}

// Add adds a feature to the registry, returning an error on empty or duplicate names.
func (r *Registry) Add(f Feature) error {
	if f == nil {
		return fmt.Errorf("nil feature")
	}
	name := f.Name()
	if name == "" {
		return fmt.Errorf("feature name must not be empty")
	}
	if _, exists := r.features[name]; exists {
		return fmt.Errorf("duplicate feature registration: %q", name)
	}
	r.order = append(r.order, name)
	r.features[name] = f
	return nil
}

// Names returns all registered feature names in registration order.
func (r *Registry) Names() []string {
	return slices.Clone(r.order)
}

// Active holds the features that were enabled and registered on a server.
type Active struct {
	Features  []Feature
	ToolNames []string
}

// Close closes all active features, aggregating any errors.
func (a *Active) Close() error {
	var firstErr error
	for _, f := range a.Features {
		if err := f.Close(); err != nil && firstErr == nil {
			firstErr = fmt.Errorf("close feature %s: %w", f.Name(), err)
		}
	}
	return firstErr
}

// RegisterEnabled registers all features marked true in enabled onto s.
// Unknown feature names in enabled return an error.
func (r *Registry) RegisterEnabled(s *mcp.Server, env Env, enabled map[string]bool) (*Active, error) {
	for k := range enabled {
		if _, ok := r.features[k]; !ok {
			return nil, fmt.Errorf("unknown feature %q", k)
		}
	}

	seenTools := make(map[string]string)
	active := &Active{}

	for _, name := range r.order {
		if !enabled[name] {
			continue
		}
		f := r.features[name]
		for _, tName := range f.Tools() {
			if prev, clash := seenTools[tName]; clash {
				return nil, fmt.Errorf("tool name collision %q between features %q and %q", tName, prev, name)
			}
			seenTools[tName] = name
			active.ToolNames = append(active.ToolNames, tName)
		}
		if err := f.Register(s, env); err != nil {
			_ = active.Close()
			return nil, fmt.Errorf("register feature %s: %w", name, err)
		}
		active.Features = append(active.Features, f)
	}

	return active, nil
}
