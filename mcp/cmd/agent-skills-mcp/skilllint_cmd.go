package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/berquerant/agent-skills/mcp/internal/skilllint"
	"github.com/spf13/pflag"
)

type skillLintOptions struct {
	strict bool
	format string
}

func runSkillLint(args []string, out io.Writer, _ io.Writer) error {
	fs := pflag.NewFlagSet("skill-lint", pflag.ContinueOnError)
	fs.SetOutput(out)

	var opts skillLintOptions
	fs.BoolVar(&opts.strict, "strict", false, "treat warnings as failures")
	fs.StringVar(&opts.format, "format", "markdown", "output format: markdown or json")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, pflag.ErrHelp) {
			return nil
		}
		return err
	}

	opts.format = strings.ToLower(strings.TrimSpace(opts.format))
	if opts.format != "markdown" && opts.format != "json" {
		return fmt.Errorf("invalid format %q: must be 'markdown' or 'json'", opts.format)
	}

	targets := fs.Args()
	batch, err := skilllint.LintTargets(targets, opts.strict)
	if err != nil {
		return err
	}

	cwd, _ := os.Getwd()
	if err := skilllint.RenderOutput(out, batch, opts.format, cwd); err != nil {
		return err
	}

	if !batch.Passed {
		return errors.New("skill-lint failed")
	}
	return nil
}
