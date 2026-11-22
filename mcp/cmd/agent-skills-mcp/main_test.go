package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCLISubcommands(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		useTransport   bool
		wantErr        string
		wantOutContain []string
		wantOutEmpty   bool
		wantErrContain string
		wantErrEmpty   bool
	}{
		{
			name:           "version outputs server name and version",
			args:           []string{"version"},
			wantOutContain: []string{"agent-skills-mcp "},
			wantErrEmpty:   true,
		},
		{
			name: "help subcommand outputs usage and flags",
			args: []string{"help"},
			wantOutContain: []string{
				"## Usage",
				"## Configuration & Precedence",
				"## Flags",
				"--task-dag",
			},
			wantErrEmpty: true,
		},
		{
			name: "--help flag on default serve outputs usage and flags",
			args: []string{"--help"},
			wantOutContain: []string{
				"## Usage",
				"--cache-max-ttl",
			},
			wantErrEmpty: true,
		},
		{
			name:    "unknown subcommand fails",
			args:    []string{"nonexistent"},
			wantErr: "unknown subcommand",
		},
		{
			name:           "skill-lint subcommand help",
			args:           []string{"skill-lint", "--help"},
			wantOutContain: []string{"Usage of skill-lint:", "--strict", "--format"},
			wantErrEmpty:   true,
		},
		{
			name:           "serve default subcommand runs over transport without polluting stdout",
			args:           []string{"--task-dag=false"},
			useTransport:   true,
			wantOutEmpty:   true,
			wantErrContain: "agent-skills-mcp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			var out, errOut bytes.Buffer
			home := t.TempDir()

			if tt.useTransport {
				runServeOverInMemoryTransport(ctx, cancel, t, tt.args, home, &out, &errOut)
			} else {
				err := run(ctx, tt.args, home, &out, &errOut, nil)
				if tt.wantErr != "" {
					require.ErrorContains(t, err, tt.wantErr)
					return
				}
				require.NoError(t, err)
			}

			if tt.wantOutEmpty {
				assert.Empty(t, out.String(), "stdout must not contain non-MCP logs")
			}
			for _, sub := range tt.wantOutContain {
				assert.Contains(t, out.String(), sub)
			}
			if tt.wantErrEmpty {
				assert.Empty(t, errOut.String())
			}
			if tt.wantErrContain != "" {
				assert.Contains(t, errOut.String(), tt.wantErrContain)
			}
		})
	}
}

func runServeOverInMemoryTransport(
	ctx context.Context,
	cancel context.CancelFunc,
	t *testing.T,
	args []string,
	home string,
	out, errOut *bytes.Buffer,
) {
	t.Helper()
	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = run(ctx, args, home, out, errOut, serverTransport)
	}()

	client := mcp.NewClient(&mcp.Implementation{Name: "cli-test", Version: "1"}, nil)
	session, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)

	res, err := session.ListTools(ctx, nil)
	require.NoError(t, err)
	assert.NotEmpty(t, res.Tools)

	_ = session.Close()
	cancel()
	<-done
}

func TestSkillLintSubcommandExecution(t *testing.T) {
	ctx := context.Background()

	t.Run("markdown output for repository skills", func(t *testing.T) {
		var out, errOut bytes.Buffer
		err := run(ctx, []string{"skill-lint", "--strict", "../../../skills/skill-review"}, t.TempDir(), &out, &errOut, nil)
		require.NoError(t, err)
		assert.Contains(t, out.String(), "# Skill Lint Report")
		assert.Contains(t, out.String(), "skill-review")
		assert.Contains(t, out.String(), "PASS")
	})

	t.Run("json output format", func(t *testing.T) {
		var out, errOut bytes.Buffer
		err := run(ctx, []string{"skill-lint", "--format=json", "../../../skills/skill-review"}, t.TempDir(), &out, &errOut, nil)
		require.NoError(t, err)
		assert.Contains(t, out.String(), `"ok": true`)
		assert.Contains(t, out.String(), `"passed": true`)
		assert.Contains(t, out.String(), `"skill_name": "skill-review"`)
	})

	t.Run("invalid format option returns error", func(t *testing.T) {
		var out, errOut bytes.Buffer
		err := run(ctx, []string{"skill-lint", "--format=yaml"}, t.TempDir(), &out, &errOut, nil)
		require.ErrorContains(t, err, "invalid format")
	})

	t.Run("nonexistent target returns error", func(t *testing.T) {
		var out, errOut bytes.Buffer
		err := run(ctx, []string{"skill-lint", "/nonexistent/target/dir"}, t.TempDir(), &out, &errOut, nil)
		require.Error(t, err)
	})

	t.Run("invalid skill fails and reports findings", func(t *testing.T) {
		tempDir := t.TempDir()
		invalidSkillDir := filepath.Join(tempDir, "bad-skill")
		require.NoError(t, os.MkdirAll(invalidSkillDir, 0o750))
		require.NoError(t, os.WriteFile(filepath.Join(invalidSkillDir, "SKILL.md"), []byte("invalid content\n"), 0o600))

		var out, errOut bytes.Buffer
		err := run(ctx, []string{"skill-lint", "--strict", invalidSkillDir}, t.TempDir(), &out, &errOut, nil)
		require.Error(t, err)
		assert.Contains(t, out.String(), "FAIL")
		assert.Contains(t, out.String(), "## Findings")
	})
}
