package server

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime/debug"
	"strings"

	"github.com/berquerant/agent-skills/mcp/internal/feature"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ServerName is the canonical MCP server implementation name.
const ServerName = "agent-skills-mcp"

// BuildVersion returns the module build version from runtime/debug.ReadBuildInfo,
// falling back to "dev" when built from a local workspace without version stamping.
func BuildVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "dev"
	}
	if info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	for _, s := range info.Settings {
		if s.Key == "vcs.revision" && len(s.Value) >= 7 {
			return "dev-" + s.Value[:7]
		}
	}
	return "dev"
}

// Instance bundles a configured MCP server and its active features.
type Instance struct {
	MCPServer *mcp.Server
	Active    *feature.Active
	Config    *feature.Config
	logFile   *os.File
}

// Close releases active features and closes any open log file.
func (inst *Instance) Close() error {
	var firstErr error
	if inst.Active != nil {
		if err := inst.Active.Close(); err != nil {
			firstErr = err
		}
	}
	if inst.logFile != nil {
		if err := inst.logFile.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// New constructs and validates a server Instance from Config, registering all enabled features
// from reg (or feature.DefaultFeatures if reg is nil) and writing a single startup summary line to errOut.
func New(cfg *feature.Config, reg *feature.Registry, clock feature.Clock, errOut io.Writer) (*Instance, error) {
	if errOut == nil {
		errOut = os.Stderr
	}

	if err := cfg.ValidateDirs(); err != nil {
		return nil, fmt.Errorf("validate directories: %w", err)
	}

	logger, logWriter, openedFile, err := setupLogger(cfg, errOut)
	if err != nil {
		return nil, err
	}

	reg, err = ensureRegistry(reg, cfg.CacheToolPrefix)
	if err != nil {
		closeLogFile(openedFile)
		return nil, err
	}

	mcpSrv := mcp.NewServer(&mcp.Implementation{
		Name:    ServerName,
		Version: BuildVersion(),
	}, nil)

	env := cfg.ToEnv(logger, clock)
	active, err := reg.RegisterEnabled(mcpSrv, env, cfg.Enabled)
	if err != nil {
		closeLogFile(openedFile)
		return nil, err
	}

	enabledNames := make([]string, 0, len(active.Features))
	for _, f := range active.Features {
		enabledNames = append(enabledNames, f.Name())
	}

	// Emit single startup summary line to logWriter (stderr or log file; never stdout).
	_, _ = fmt.Fprintf(
		logWriter,
		"%s (%s) started: features=[%s] tools=[%s]\n",
		ServerName,
		BuildVersion(),
		strings.Join(enabledNames, ","),
		strings.Join(active.ToolNames, ","),
	)

	return &Instance{
		MCPServer: mcpSrv,
		Active:    active,
		Config:    cfg,
		logFile:   openedFile,
	}, nil
}

func setupLogger(cfg *feature.Config, errOut io.Writer) (*slog.Logger, io.Writer, *os.File, error) {
	logWriter := errOut
	var openedFile *os.File
	if cfg.LogFile != "" {
		// #nosec G304 -- user-configured log file path
		f, err := os.OpenFile(cfg.LogFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("open log file %s: %w", cfg.LogFile, err)
		}
		openedFile = f
		logWriter = f
	}

	level, err := parseLogLevel(cfg.LogLevel)
	if err != nil {
		closeLogFile(openedFile)
		return nil, nil, nil, err
	}

	logger := slog.New(slog.NewTextHandler(logWriter, &slog.HandlerOptions{Level: level}))
	return logger, logWriter, openedFile, nil
}

func ensureRegistry(reg *feature.Registry, cachePrefix string) (*feature.Registry, error) {
	if reg != nil {
		return reg, nil
	}
	reg = feature.NewRegistry()
	for _, f := range feature.DefaultFeatures(cachePrefix) {
		if err := reg.Add(f); err != nil {
			return nil, err
		}
	}
	return reg, nil
}

func closeLogFile(f *os.File) {
	if f != nil {
		_ = f.Close()
	}
}

func parseLogLevel(raw string) (slog.Level, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn", "warning", "":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelWarn, fmt.Errorf("invalid --log-level %q (expected debug, info, warn, error)", raw)
	}
}
