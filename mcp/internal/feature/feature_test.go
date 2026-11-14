package feature_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/berquerant/agent-skills/mcp/internal/feature"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fixedClock struct {
	t time.Time
}

func (c fixedClock) Now() time.Time { return c.t }

type mockFeature struct {
	name      string
	tools     []string
	regErr    error
	closeErr  error
	closed    bool
	clockSeen time.Time
}

func (m *mockFeature) Name() string    { return m.name }
func (m *mockFeature) Tools() []string { return m.tools }
func (m *mockFeature) Register(_ *mcp.Server, env feature.Env) error {
	if env.Clock != nil {
		m.clockSeen = env.Clock.Now()
	}
	return m.regErr
}
func (m *mockFeature) Close() error {
	m.closed = true
	return m.closeErr
}

func TestRegistryAdd(t *testing.T) {
	tests := []struct {
		name      string
		features  []feature.Feature
		wantErr   string
		wantNames []string
	}{
		{
			name:     "nil feature returns error",
			features: []feature.Feature{nil},
			wantErr:  "nil feature",
		},
		{
			name:     "empty feature name returns error",
			features: []feature.Feature{&mockFeature{name: ""}},
			wantErr:  "empty",
		},
		{
			name: "duplicate feature name returns error",
			features: []feature.Feature{
				&mockFeature{name: "alpha", tools: []string{"t1"}},
				&mockFeature{name: "alpha", tools: []string{"t2"}},
			},
			wantErr: "duplicate",
		},
		{
			name: "registers distinct features in order",
			features: []feature.Feature{
				&mockFeature{name: "alpha", tools: []string{"t1"}},
				&mockFeature{name: "beta", tools: []string{"t2"}},
			},
			wantNames: []string{"alpha", "beta"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := feature.NewRegistry()
			var err error
			for _, f := range tt.features {
				if addErr := reg.Add(f); addErr != nil {
					err = addErr
					break
				}
			}
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantNames, reg.Names())
		})
	}
}

func TestRegistryRegisterEnabled(t *testing.T) {
	now := time.Date(2026, 11, 15, 1, 30, 0, 0, time.UTC)

	tests := []struct {
		name           string
		f1             *mockFeature
		f2             *mockFeature
		enabled        map[string]bool
		wantErr        string
		wantTools      []string
		wantF1Closed   bool
		wantF2Closed   bool
		wantF1ClockSet bool
	}{
		{
			name:           "registers enabled feature and injects clock",
			f1:             &mockFeature{name: "f1", tools: []string{"t1"}},
			f2:             &mockFeature{name: "f2", tools: []string{"t2"}},
			enabled:        map[string]bool{"f1": true, "f2": false},
			wantTools:      []string{"t1"},
			wantF1Closed:   true,
			wantF2Closed:   false,
			wantF1ClockSet: true,
		},
		{
			name:    "unknown feature name in enabled returns error",
			f1:      &mockFeature{name: "f1", tools: []string{"t1"}},
			f2:      &mockFeature{name: "f2", tools: []string{"t2"}},
			enabled: map[string]bool{"unknown": true},
			wantErr: "unknown feature",
		},
		{
			name:    "tool name collision returns error",
			f1:      &mockFeature{name: "f1", tools: []string{"dup"}},
			f2:      &mockFeature{name: "f2", tools: []string{"dup"}},
			enabled: map[string]bool{"f1": true, "f2": true},
			wantErr: "collision",
		},
		{
			name:         "registration error closes already registered features",
			f1:           &mockFeature{name: "f1", tools: []string{"t1"}},
			f2:           &mockFeature{name: "f2", tools: []string{"t2"}, regErr: errors.New("boom")},
			enabled:      map[string]bool{"f1": true, "f2": true},
			wantErr:      "boom",
			wantF1Closed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := feature.NewRegistry()
			require.NoError(t, reg.Add(tt.f1))
			require.NoError(t, reg.Add(tt.f2))

			srv := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "1"}, nil)
			active, err := reg.RegisterEnabled(srv, feature.Env{Clock: fixedClock{t: now}}, tt.enabled)
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				assert.Equal(t, tt.wantF1Closed, tt.f1.closed)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantTools, active.ToolNames)
			if tt.wantF1ClockSet {
				assert.Equal(t, now, tt.f1.clockSeen)
			}
			require.NoError(t, active.Close())
			assert.Equal(t, tt.wantF1Closed, tt.f1.closed)
			assert.Equal(t, tt.wantF2Closed, tt.f2.closed)
		})
	}
}

type parseConfigTestCase struct {
	name    string
	args    []string
	env     map[string]string
	wantErr string
	check   func(t *testing.T, cfg *feature.Config)
}

func parseConfigTestCases(home string) []parseConfigTestCase {
	return []parseConfigTestCase{
		{
			name: "defaults when no flags or env vars are set",
			check: func(t *testing.T, cfg *feature.Config) {
				for _, name := range feature.FeatureNames {
					assert.True(t, cfg.Enabled[name], "feature %s should default to true", name)
				}
				assert.Equal(t, filepath.Join(home, ".local", "state", "agent-skills-mcp"), cfg.StateDir)
				assert.Equal(t, filepath.Join(home, ".cache", "agent-skills-mcp", "cache"), cfg.CacheDir)
				assert.Equal(t, filepath.Join(home, ".cache", "agent-skills-mcp", "checkpoints"), cfg.CheckpointRoot)
				assert.True(t, cfg.CheckpointDefault)
				assert.Equal(t, 1800, cfg.CacheMaxTTL)
			},
		},
		{
			name: "XDG_CACHE_HOME and XDG_STATE_HOME override home defaults",
			env: map[string]string{
				"XDG_CACHE_HOME": "/custom/xdg/cache",
				"XDG_STATE_HOME": "/custom/xdg/state",
			},
			check: func(t *testing.T, cfg *feature.Config) {
				assert.Equal(t, "/custom/xdg/state/agent-skills-mcp", cfg.StateDir)
				assert.Equal(t, "/custom/xdg/cache/agent-skills-mcp/cache", cfg.CacheDir)
				assert.Equal(t, "/custom/xdg/cache/agent-skills-mcp/checkpoints", cfg.CheckpointRoot)
				assert.True(t, cfg.CheckpointDefault)
			},
		},
		{
			name: "rejects relative XDG_CACHE_HOME",
			env: map[string]string{
				"XDG_CACHE_HOME": "relative/cache",
			},
			wantErr: "XDG_CACHE_HOME must be an absolute path",
		},
		{
			name: "rejects relative XDG_STATE_HOME",
			env: map[string]string{
				"XDG_STATE_HOME": "relative/state",
			},
			wantErr: "XDG_STATE_HOME must be an absolute path",
		},
		{
			name: "precedence: CLI flag > env > default via structconfig",
			env: map[string]string{
				"AGENT_SKILLS_MCP_TASK_DAG":   "false",
				"AGENT_SKILLS_MCP_PACE_GUARD": "false",
				"AGENT_SKILLS_MCP_LOG_LEVEL":  "debug",
				"AGENT_SKILLS_MCP_CACHE_DIR":  "/from/env_cache_dir",
			},
			args: []string{
				"--task-dag=true",
				"--skill-lint=false",
				"--log-level=warn",
				"--checkpoint-root=/explicit/checkpoints",
			},
			check: func(t *testing.T, cfg *feature.Config) {
				assert.True(t, cfg.Enabled["task-dag"], "CLI flag should override env")
				assert.False(t, cfg.Enabled["pace-guard"], "env should override default")
				assert.False(t, cfg.Enabled["skill-lint"], "CLI flag should disable skill-lint")
				assert.Equal(t, "warn", cfg.LogLevel, "CLI flag matching default should override env")
				assert.Equal(t, "/from/env_cache_dir", cfg.CacheDir)
				assert.Equal(t, "/explicit/checkpoints", cfg.CheckpointRoot)
				assert.False(t, cfg.CheckpointDefault)

				cfg.TaskDAG = false
				cfg.SyncEnabledMap()
				assert.False(t, cfg.Enabled["task-dag"], "SyncEnabledMap should refresh canonical feature flags")
			},
		},
		{
			name: "invalid env boolean fails",
			env: map[string]string{
				"AGENT_SKILLS_MCP_CACHE": "not_a_bool",
			},
			wantErr: "invalid syntax",
		},
		{
			name:    "non-positive cache-max-ttl fails",
			args:    []string{"--cache-max-ttl=0"},
			wantErr: "--cache-max-ttl must be > 0",
		},
	}
}

func TestParseConfig(t *testing.T) {
	home := "/Users/testuser"

	for _, tt := range parseConfigTestCases(home) {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.env {
				t.Setenv(k, v)
			}
			cfg, err := feature.ParseConfig(tt.args, home)
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			if tt.check != nil {
				tt.check(t, cfg)
			}
		})
	}
}

func TestValidateDirs(t *testing.T) {
	tmp := t.TempDir()
	blockingFile := filepath.Join(tmp, "blocking_file")
	require.NoError(t, os.WriteFile(blockingFile, []byte("x"), 0o600))

	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name: "valid writable directories succeed",
			args: []string{
				"--state-dir=" + filepath.Join(tmp, "state"),
				"--cache-dir=" + filepath.Join(tmp, "cache"),
				"--checkpoint-root=" + filepath.Join(tmp, "cp"),
			},
		},
		{
			name:    "file blocking state-dir fails",
			args:    []string{"--state-dir=" + blockingFile},
			wantErr: "not a directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := feature.ParseConfig(tt.args, tmp)
			require.NoError(t, err)
			err = cfg.ValidateDirs()
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestUsageAndHelpMarkdown(t *testing.T) {
	u := feature.Usage()
	assert.Contains(t, u, "## Usage")
	assert.Contains(t, u, "## Configuration & Precedence")
	assert.Contains(t, u, "## Flags")
	assert.Contains(t, u, "--cache-max-ttl")
}
