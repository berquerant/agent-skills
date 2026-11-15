package skilllint

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"sync"
	"text/template"
)

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

func buildMarkdownTemplateData(batch BatchReport, cwd string) markdownTemplateData {
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

// RenderMarkdown writes a formatted Markdown lint report to out.
func RenderMarkdown(out io.Writer, batch BatchReport, cwd string) error {
	data := buildMarkdownTemplateData(batch, cwd)
	return getMarkdownTemplate().Execute(out, data)
}

// RenderJSON writes a formatted JSON lint report to out with 2-space indentation.
func RenderJSON(out io.Writer, batch BatchReport) error {
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return enc.Encode(batch)
}

// RenderOutput renders the batch report in the requested format ("markdown" or "json").
func RenderOutput(out io.Writer, batch BatchReport, format, cwd string) error {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "json":
		return RenderJSON(out, batch)
	case "markdown", "":
		return RenderMarkdown(out, batch, cwd)
	default:
		return fmt.Errorf("invalid format %q: must be 'markdown' or 'json'", format)
	}
}
