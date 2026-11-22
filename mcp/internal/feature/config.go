package feature

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/berquerant/agent-skills/mcp/internal/guard"
	"github.com/berquerant/structconfig"
	"github.com/spf13/pflag"
)

// EnvPrefix is the prefix for all structconfig-managed environment variables.
const EnvPrefix = "AGENT_SKILLS_MCP_"

// FeatureNames lists all 8 features in canonical order.
var FeatureNames = []string{
	"cache",
	"task-dag",
	"skill-lint",
	"snippet-analyzer",
	"snippet-manager",
	"checkpoint",
	"pace-guard",
	"pr-triage",
}

// Config holds the runtime configuration for agent-skills-mcp.
// Values are resolved in precedence order: default < environment variable (AGENT_SKILLS_MCP_*) < CLI flag,
// using github.com/berquerant/structconfig.
type Config struct {
	Cache           bool `name:"cache" default:"true" usage:"enable cache feature"`
	TaskDAG         bool `name:"task-dag" default:"true" usage:"enable task-dag feature"`
	SkillLint       bool `name:"skill-lint" default:"true" usage:"enable skill-lint feature"`
	SnippetAnalyzer bool `name:"snippet-analyzer" default:"true" usage:"enable snippet-analyzer feature"`
	SnippetManager  bool `name:"snippet-manager" default:"true" usage:"enable snippet-manager feature"`
	Checkpoint      bool `name:"checkpoint" default:"true" usage:"enable checkpoint feature"`
	PaceGuard       bool `name:"pace-guard" default:"true" usage:"enable pace-guard feature"`
	PRTriage        bool `name:"pr-triage" default:"true" usage:"enable pr-triage feature"`

	StateDir       string `name:"state-dir" default:"" usage:"state directory (default: $XDG_STATE_HOME/agent-skills-mcp)"`
	CacheDir       string `name:"cache-dir" default:"" usage:"cache directory (default: $XDG_CACHE_HOME/agent-skills-mcp/cache)"`
	CheckpointRoot string `name:"checkpoint-root" default:"" usage:"checkpoint root directory (default: $XDG_CACHE_HOME/agent-skills-mcp/checkpoints)"`
	SnippetRoot    string `name:"snippet-root" default:"" usage:"snippet root directory (default: $XDG_STATE_HOME/agent-skills-mcp/snippets)"`

	CacheMaxTTL       int   `name:"cache-max-ttl" default:"1800" usage:"maximum allowed cache TTL in seconds"`
	CacheMaxEntrySize int64 `name:"cache-max-entry-size" default:"10485760" usage:"maximum uncompressed cache entry size in bytes"`
	CacheMaxTotalSize int64 `name:"cache-max-total-size" default:"524288000" usage:"maximum total compressed cache size in bytes"`

	MaxOutputSize int    `name:"max-output-size" default:"32768" usage:"maximum serialized tool output size in bytes"`
	LogLevel      string `name:"log-level" default:"warn" usage:"log level (debug, info, warn, error)"`
	LogFile       string `name:"log-file" default:"" usage:"optional log file path (default: stderr)"`

	// Enabled maps each canonical feature name to its resolved boolean state.
	Enabled map[string]bool `name:"-"`
	// CheckpointDefault reports whether CheckpointRoot was resolved from the XDG default rather than an explicit flag/env.
	CheckpointDefault bool `name:"-"`
}

// SetFlags registers all Config fields as pflags on fs using structconfig.
func SetFlags(fs *pflag.FlagSet) error {
	sc := structconfig.New[Config](structconfig.WithEnvPrefix(EnvPrefix))
	return sc.SetFlags(fs)
}

// FlagUsages returns the formatted pflag usage string for Config flags.
func FlagUsages() string {
	fs := pflag.NewFlagSet("agent-skills-mcp", pflag.ContinueOnError)
	_ = SetFlags(fs)
	return strings.TrimRight(fs.FlagUsages(), "\n")
}

// ParseConfig parses flags from args and resolves environment variables and XDG defaults
// using github.com/berquerant/structconfig.
// Precedence: CLI flag > environment variable > default value.
func ParseConfig(args []string, homeDir string) (*Config, error) {
	if homeDir == "" {
		if h, err := os.UserHomeDir(); err == nil {
			homeDir = h
		}
	}

	fs := pflag.NewFlagSet("agent-skills-mcp", pflag.ContinueOnError)
	fs.Usage = func() {}
	sc := structconfig.New[Config](structconfig.WithEnvPrefix(EnvPrefix))
	merger := structconfig.NewMerger[Config]()

	cfg, err := structconfig.NewConfigWithMerge(
		sc,
		merger,
		fs,
		structconfig.WithEnvPrefix(EnvPrefix),
		structconfig.WithArguments(args),
	)
	if err != nil {
		return nil, err
	}

	// When a flag is explicitly passed on the CLI with a value equal to the default tag
	// (e.g. --task-dag=true when AGENT_SKILLS_MCP_TASK_DAG=false), structconfig.Merger
	// compares against the default tag and keeps left (env) unless we apply explicitly
	// changed flags from fs.Changed.
	var fromFlags Config
	if err := sc.FromFlags(&fromFlags, fs); err != nil {
		return nil, err
	}
	applyChangedFlags(cfg, &fromFlags, fs)

	if err := cfg.finalize(homeDir); err != nil {
		return nil, err
	}
	return cfg, nil
}

func applyChangedFlags(dst, fromFlags *Config, fs *pflag.FlagSet) {
	dstVal := reflect.ValueOf(dst).Elem()
	srcVal := reflect.ValueOf(fromFlags).Elem()
	typ := dstVal.Type()
	for i := range typ.NumField() {
		tag := typ.Field(i).Tag.Get("name")
		if tag == "" || tag == "-" {
			continue
		}
		if fs.Changed(tag) && dstVal.Field(i).CanSet() {
			dstVal.Field(i).Set(srcVal.Field(i))
		}
	}
}

func (c *Config) finalize(homeDir string) error {
	c.SyncEnabledMap()

	xdgCache, err := resolveXDGDir("XDG_CACHE_HOME", filepath.Join(homeDir, ".cache"))
	if err != nil {
		return err
	}
	xdgState, err := resolveXDGDir("XDG_STATE_HOME", filepath.Join(homeDir, ".local", "state"))
	if err != nil {
		return err
	}

	if c.StateDir == "" {
		c.StateDir = filepath.Join(xdgState, "agent-skills-mcp")
	}

	if c.CacheDir == "" {
		c.CacheDir = filepath.Join(xdgCache, "agent-skills-mcp", "cache")
	}

	if c.CheckpointRoot == "" {
		c.CheckpointRoot = filepath.Join(xdgCache, "agent-skills-mcp", "checkpoints")
		c.CheckpointDefault = true
	}

	if c.SnippetRoot == "" {
		c.SnippetRoot = filepath.Join(xdgState, "agent-skills-mcp", "snippets")
	}

	return c.validateLimits()
}

// SyncEnabledMap initializes Enabled from the canonical feature boolean fields when
// Enabled is nil, or refreshes the canonical feature entries if they are already present.
func (c *Config) SyncEnabledMap() {
	if c.Enabled == nil {
		c.Enabled = make(map[string]bool, len(FeatureNames))
	} else if _, hasStandard := c.Enabled["cache"]; !hasStandard {
		return
	}
	c.Enabled["cache"] = c.Cache
	c.Enabled["task-dag"] = c.TaskDAG
	c.Enabled["skill-lint"] = c.SkillLint
	c.Enabled["snippet-analyzer"] = c.SnippetAnalyzer
	c.Enabled["snippet-manager"] = c.SnippetManager
	c.Enabled["checkpoint"] = c.Checkpoint
	c.Enabled["pace-guard"] = c.PaceGuard
	c.Enabled["pr-triage"] = c.PRTriage
}

func (c *Config) validateLimits() error {
	if c.CacheMaxTTL <= 0 {
		return fmt.Errorf("--cache-max-ttl must be > 0 (got %d)", c.CacheMaxTTL)
	}
	if c.CacheMaxEntrySize <= 0 {
		return fmt.Errorf("--cache-max-entry-size must be > 0 (got %d)", c.CacheMaxEntrySize)
	}
	if c.CacheMaxTotalSize <= 0 {
		return fmt.Errorf("--cache-max-total-size must be > 0 (got %d)", c.CacheMaxTotalSize)
	}
	if c.MaxOutputSize <= 0 {
		return fmt.Errorf("--max-output-size must be > 0 (got %d)", c.MaxOutputSize)
	}
	for _, p := range []string{c.StateDir, c.CacheDir, c.CheckpointRoot, c.SnippetRoot, c.LogFile} {
		if err := guard.CheckPathAllowed(p); err != nil {
			return err
		}
	}
	return nil
}

// ValidateDirs ensures required directories for enabled features exist and are writable (fail-fast).
func (c *Config) ValidateDirs() error {
	c.SyncEnabledMap()

	if c.Enabled["task-dag"] || c.Enabled["pace-guard"] {
		if err := ensureWritableDir(c.StateDir); err != nil {
			return fmt.Errorf("state-dir (%s): %w", c.StateDir, err)
		}
	}
	if c.Enabled["cache"] {
		if err := ensureWritableDir(c.CacheDir); err != nil {
			return fmt.Errorf("cache-dir (%s): %w", c.CacheDir, err)
		}
	}
	if err := ensureFeatureDir(c.Enabled["checkpoint"], "checkpoint-root", c.CheckpointRoot); err != nil {
		return err
	}
	return ensureFeatureDir(c.Enabled["snippet-manager"], "snippet-root", c.SnippetRoot)
}

func ensureFeatureDir(enabled bool, flagName, dir string) error {
	if !enabled {
		return nil
	}
	if dir == "" {
		return fmt.Errorf("%s must not be empty when feature is enabled", flagName)
	}
	if err := ensureWritableDir(dir); err != nil {
		return fmt.Errorf("%s (%s): %w", flagName, dir, err)
	}
	return nil
}

// ToEnv converts Config into a feature Env.
func (c *Config) ToEnv(logger *slog.Logger, clock Clock) Env {
	if clock == nil {
		clock = SystemClock{}
	}
	return Env{
		Logger:            logger,
		Clock:             clock,
		StateDir:          c.StateDir,
		CacheDir:          c.CacheDir,
		CheckpointDir:     c.CheckpointRoot,
		SnippetDir:        c.SnippetRoot,
		MaxOutputSize:     c.MaxOutputSize,
		CacheMaxTTL:       c.CacheMaxTTL,
		CacheMaxEntrySize: c.CacheMaxEntrySize,
		CacheMaxTotalSize: c.CacheMaxTotalSize,
	}
}

// resolveXDGDir implements the XDG Base Directory specification requirement that
// XDG environment paths must be absolute; relative paths are rejected.
func resolveXDGDir(key, fallback string) (string, error) {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		if !filepath.IsAbs(val) {
			return "", fmt.Errorf("%s must be an absolute path (got %q)", key, val)
		}
		return filepath.Clean(val), nil
	}
	return filepath.Clean(fallback), nil
}

func ensureWritableDir(dir string) error {
	if dir == "" {
		return fmt.Errorf("directory path is empty")
	}
	if err := guard.CheckPathAllowed(dir); err != nil {
		return err
	}
	info, err := os.Stat(dir)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		if mkErr := os.MkdirAll(dir, 0o700); mkErr != nil {
			return mkErr
		}
	} else if !info.IsDir() {
		return fmt.Errorf("path exists and is not a directory")
	}

	probe, err := os.CreateTemp(dir, ".write_probe_*")
	if err != nil {
		return fmt.Errorf("directory is not writable: %w", err)
	}
	probeName := probe.Name()
	_ = probe.Close()
	_ = os.Remove(probeName)
	return nil
}
