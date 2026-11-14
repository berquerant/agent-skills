package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/berquerant/agent-skills/mcp/internal/feature"
	"github.com/berquerant/agent-skills/mcp/internal/server"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/pflag"
)

func main() {
	os.Exit(mainExit())
}

func mainExit() int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, os.Args[1:], "", os.Stdout, os.Stderr, nil); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s: %v\n", server.ServerName, err)
		return 1
	}
	return 0
}

func run(
	ctx context.Context,
	args []string,
	homeDir string,
	out io.Writer,
	errOut io.Writer,
	transport mcp.Transport,
) error {
	sub := "serve"
	rest := args
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		sub = args[0]
		rest = args[1:]
	}

	switch sub {
	case "version":
		_, err := fmt.Fprintf(out, "%s %s\n", server.ServerName, server.BuildVersion())
		return err

	case "help":
		_, err := fmt.Fprintf(out, "%s\n", feature.Usage())
		return err

	case "serve":
		cfg, err := feature.ParseConfig(rest, homeDir)
		if errors.Is(err, pflag.ErrHelp) {
			_, wErr := fmt.Fprintf(out, "%s\n", feature.Usage())
			return wErr
		}
		if err != nil {
			return err
		}
		inst, err := server.New(cfg, nil, nil, errOut)
		if err != nil {
			return err
		}
		defer func() {
			_ = inst.Close()
		}()
		t := transport
		if t == nil {
			t = &mcp.StdioTransport{}
		}
		return inst.MCPServer.Run(ctx, t)

	default:
		return fmt.Errorf("unknown subcommand %q (supported: serve, version, help)", sub)
	}
}
