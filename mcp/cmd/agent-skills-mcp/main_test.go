package main

import (
	"bytes"
	"context"
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
				serverTransport, clientTransport := mcp.NewInMemoryTransports()
				done := make(chan struct{})
				go func() {
					defer close(done)
					_ = run(ctx, tt.args, home, &out, &errOut, serverTransport)
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
