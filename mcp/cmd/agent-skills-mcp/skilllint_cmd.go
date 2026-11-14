package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"text/template"

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
	if len(targets) == 0 {
		targets = []string{"."}
	}

	batch, err := executeSkillLint(targets, opts.strict)
	if err != nil {
		return err
	}

	if err := renderSkillLintOutput(out, batch, opts.format); err != nil {
		return err
	}

	if !batch.Passed {
		return errors.New("skill-lint failed")
	}
	return nil
}

func executeSkillLint(targets []string, strict bool) (skilllint.BatchReport, error) {
	var allReports []skilllint.SkillReport
	seenDirs := make(map[string]bool)

	for _, target := range targets {
		reports, err := collectTargetReports(target, strict, seenDirs)
		if err != nil {
			return skilllint.BatchReport{}, err
		}
		allReports = append(allReports, reports...)
	}

	passedSkills := 0
	failedSkills := 0
	totalErrors := 0
	totalWarnings := 0

	for _, rep := range allReports {
		totalErrors += rep.Errors
		totalWarnings += rep.Warnings
		if rep.Passed {
			passedSkills++
		} else {
			failedSkills++
		}
	}

	return skilllint.BatchReport{
		OK:            true,
		Passed:        failedSkills == 0,
		Strict:        strict,
		RootDir:       strings.Join(targets, ", "),
		TotalSkills:   len(allReports),
		PassedSkills:  passedSkills,
		FailedSkills:  failedSkills,
		TotalErrors:   totalErrors,
		TotalWarnings: totalWarnings,
		Reports:       allReports,
	}, nil
}

func collectTargetReports(target string, strict bool, seenDirs map[string]bool) ([]skilllint.SkillReport, error) {
	abs, err := filepath.Abs(target)
	if err != nil {
		return nil, fmt.Errorf("resolve path %s: %w", target, err)
	}
	info, err := os.Stat(abs)
	if err != nil {
		return nil, fmt.Errorf("stat target %s: %w", target, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("target is not a directory: %s", target)
	}

	if _, err := os.Stat(filepath.Join(abs, "SKILL.md")); err == nil {
		if seenDirs[abs] {
			return nil, nil
		}
		seenDirs[abs] = true
		rep, lintErr := skilllint.LintSkill(abs, strict)
		if lintErr != nil {
			return nil, lintErr
		}
		return []skilllint.SkillReport{rep}, nil
	}

	batch, err := skilllint.LintAll(abs, strict)
	if err != nil {
		return nil, err
	}
	var deduped []skilllint.SkillReport
	for _, rep := range batch.Reports {
		repAbs, absErr := filepath.Abs(rep.TargetDir)
		if absErr != nil {
			return nil, fmt.Errorf("resolve report path %s: %w", rep.TargetDir, absErr)
		}
		if !seenDirs[repAbs] {
			seenDirs[repAbs] = true
			deduped = append(deduped, rep)
		}
	}
	return deduped, nil
}

func renderSkillLintOutput(out io.Writer, batch skilllint.BatchReport, format string) error {
	if format == "json" {
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		return enc.Encode(batch)
	}
	return writeMarkdownReport(out, batch)
}

const markdownReportTemplate = `# Skill Lint Report

{{if not .Reports}}No skills found.

{{else}}| Skill | Lines | Errors | Warnings | Status |
|---|---|---|---|---|
{{range .Reports}}| {{.Name}} (` + "`" + `{{.DisplayDir}}` + "`" + `) | {{.Lines}} | {{.Errors}} | {{.Warnings}} | {{.Status}} |
{{end}}
{{end}}{{if .HasFindings}}## Findings

{{range .Findings}}### {{.Name}} (` + "`" + `{{.DisplayDir}}` + "`" + `)

{{range .Items}}- ` + "`" + `[Check {{.Check}}]` + "`" + ` ({{.Severity}}){{.Location}}: {{.Message}}
{{end}}
{{end}}{{end}}## Summary

- **Skills**: {{.TotalSkills}} total ({{.PassedSkills}} passed, {{.FailedSkills}} failed)
- **Findings**: {{.TotalErrors}} errors, {{.TotalWarnings}} warnings
- **Result**: {{.ResultStatus}}
`

var (
	markdownTmplOnce sync.Once
	markdownTmpl     *template.Template
)

func getMarkdownTemplate() *template.Template {
	markdownTmplOnce.Do(func() {
		markdownTmpl = template.Must(template.New("markdownReport").Parse(markdownReportTemplate))
	})
	return markdownTmpl
}

type markdownTemplateData struct {
	Reports       []markdownSkillRow
	HasFindings   bool
	Findings      []markdownFindingGroup
	TotalSkills   int
	PassedSkills  int
	FailedSkills  int
	TotalErrors   int
	TotalWarnings int
	ResultStatus  string
}

type markdownSkillRow struct {
	Name       string
	DisplayDir string
	Lines      int
	Errors     int
	Warnings   int
	Status     string
}

type markdownFindingItem struct {
	Check    string
	Severity string
	Location string
	Message  string
}

type markdownFindingGroup struct {
	Name       string
	DisplayDir string
	Items      []markdownFindingItem
}

func toDisplayDir(targetDir, cwd string) string {
	if cwd != "" {
		if rel, err := filepath.Rel(cwd, targetDir); err == nil {
			if !strings.HasPrefix(rel, "..") {
				return filepath.ToSlash(rel)
			}
		}
	}
	return filepath.ToSlash(targetDir)
}

func buildMarkdownTemplateData(batch skilllint.BatchReport, cwd string) markdownTemplateData {
	rows := make([]markdownSkillRow, 0, len(batch.Reports))
	var findingGroups []markdownFindingGroup

	for _, rep := range batch.Reports {
		displayDir := toDisplayDir(rep.TargetDir, cwd)
		status := "PASS"
		if !rep.Passed {
			status = "**FAIL**"
		}
		rows = append(rows, markdownSkillRow{
			Name:       rep.SkillName,
			DisplayDir: displayDir,
			Lines:      rep.SkillLines,
			Errors:     rep.Errors,
			Warnings:   rep.Warnings,
			Status:     status,
		})

		if len(rep.Findings) > 0 {
			var items []markdownFindingItem
			for _, f := range rep.Findings {
				loc := ""
				if f.File != "" && f.Line > 0 {
					loc = fmt.Sprintf(" `%s:%d`", f.File, f.Line)
				} else if f.File != "" {
					loc = fmt.Sprintf(" `%s`", f.File)
				}
				items = append(items, markdownFindingItem{
					Check:    string(f.Check),
					Severity: string(f.Severity),
					Location: loc,
					Message:  f.Message,
				})
			}
			findingGroups = append(findingGroups, markdownFindingGroup{
				Name:       rep.SkillName,
				DisplayDir: displayDir,
				Items:      items,
			})
		}
	}

	resultStatus := "PASS"
	if !batch.Passed {
		resultStatus = "**FAIL**"
	}

	return markdownTemplateData{
		Reports:       rows,
		HasFindings:   len(findingGroups) > 0,
		Findings:      findingGroups,
		TotalSkills:   batch.TotalSkills,
		PassedSkills:  batch.PassedSkills,
		FailedSkills:  batch.FailedSkills,
		TotalErrors:   batch.TotalErrors,
		TotalWarnings: batch.TotalWarnings,
		ResultStatus:  resultStatus,
	}
}

func writeMarkdownReport(out io.Writer, batch skilllint.BatchReport) error {
	cwd, _ := os.Getwd()
	data := buildMarkdownTemplateData(batch, cwd)
	return getMarkdownTemplate().Execute(out, data)
}
