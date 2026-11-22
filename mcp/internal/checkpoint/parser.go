package checkpoint

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/berquerant/agent-skills/mcp/internal/pathutil"
)

var stepHeadingRe = regexp.MustCompile(`^##\s+\[([^\]]+)\]\s+Step\s+([0-9]+):\s+(.+)$`)

// ParsedLog holds the parsed header block, milestone entries, and non-fatal parse warnings.
type ParsedLog struct {
	Header   Header
	Entries  []StepEntry
	Warnings []string
}

// Resume resolves a checkpoint log (by explicit path or latest generation for context) and parses its recovery state.
func (m *Manager) Resume(in ResumeInput) (ResumeOutput, error) {
	targetPath, err := m.resolveResumePath(in)
	if err != nil {
		return ResumeOutput{}, err
	}
	resolved, err := ResolveRelativeWithinRoot(m.rootDir, targetPath)
	if err != nil {
		return ResumeOutput{}, err
	}

	var out ResumeOutput
	err = m.withRLockedRoot(resolved.Rel, func(root *os.Root) error {
		raw, err := root.ReadFile(resolved.Rel)
		if err != nil {
			return fmt.Errorf("read checkpoint log %s: %w", resolved.Abs, err)
		}
		parsed := ParseLogMarkdown(string(raw))
		out = buildResumeOutput(resolved.Abs, parsed)
		return nil
	})
	return out, err
}

func (m *Manager) resolveResumePath(in ResumeInput) (string, error) {
	if strings.TrimSpace(in.Path) != "" {
		return in.Path, nil
	}
	if strings.TrimSpace(in.Context) == "" {
		return "", fmt.Errorf("either path or context must be provided")
	}
	docType, err := NormalizeDocType(in.DocType)
	if err != nil {
		return "", err
	}
	ctxSlug, err := pathutil.SanitizeSlug(in.Context)
	if err != nil {
		return "", err
	}
	pb, err := pathutil.ResolveProjectBranch(pathutil.ProjectBranchInput{
		Project: in.Project,
		Branch:  in.Branch,
		WorkDir: in.WorkDir,
	})
	if err != nil {
		return "", err
	}
	latest, err := FindLatestGeneration(GenerationQuery{
		RootDir: m.rootDir,
		Project: pb.Project,
		Branch:  pb.Branch,
		Context: ctxSlug,
		DocType: docType,
	})
	if err != nil {
		return "", err
	}
	if latest == "" {
		return "", fmt.Errorf("no %s log found for %s/%s/%s", docType, pb.Project, pb.Branch, ctxSlug)
	}
	return latest, nil
}

func buildResumeOutput(absPath string, parsed ParsedLog) ResumeOutput {
	completed := []string{}
	unresolved := []StepEntry{}
	var lastCompleted *StepEntry
	var latestEntry *StepEntry
	nextSteps := []string{}

	for i := range parsed.Entries {
		e := parsed.Entries[i]
		latestEntry = &e
		switch e.Status {
		case StepStatusCompleted:
			completed = append(completed, fmt.Sprintf("Step %d: %s", e.Step, e.Title))
			cp := e
			lastCompleted = &cp
		case StepStatusBlocked, StepStatusFailed:
			unresolved = append(unresolved, e)
		}
	}
	if latestEntry != nil && len(latestEntry.NextSteps) > 0 {
		nextSteps = latestEntry.NextSteps
	}
	warnings := parsed.Warnings
	if warnings == nil {
		warnings = []string{}
	}

	return ResumeOutput{
		OK:               true,
		Path:             absPath,
		DualPath:         pathutil.FormatDualPath(absPath),
		Header:           parsed.Header,
		TotalSteps:       len(parsed.Entries),
		CompletedSteps:   completed,
		UnresolvedSteps:  unresolved,
		LastCompleted:    lastCompleted,
		LatestEntry:      latestEntry,
		CurrentNextSteps: nextSteps,
		ParseWarnings:    warnings,
	}
}

// ParseLogMarkdown parses the header block and milestone entries from a checkpoint log Markdown document.
func ParseLogMarkdown(content string) ParsedLog {
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	hdrSec := parseHeaderSection(lines)
	entries, stepWarnings := parseStepEntries(lines[hdrSec.bodyStart:])
	warnings := append(hdrSec.warnings, stepWarnings...)
	return ParsedLog{
		Header:   hdrSec.header,
		Entries:  entries,
		Warnings: warnings,
	}
}

type parsedHeaderSection struct {
	header    Header
	bodyStart int
	warnings  []string
}

func parseHeaderSection(lines []string) parsedHeaderSection {
	var h Header
	var warnings []string
	idx := 0
	for ; idx < len(lines); idx++ {
		line := strings.TrimSpace(lines[idx])
		if strings.HasPrefix(line, "## [") {
			break
		}
		if after, ok := strings.CutPrefix(line, "# Task Checkpoint Log:"); ok {
			h.Title = strings.TrimSpace(after)
			continue
		}
		parseHeaderBullet(&h, line)
	}
	if h.Title == "" {
		warnings = append(warnings, "missing '# Task Checkpoint Log:' header title")
	}
	return parsedHeaderSection{
		header:    h,
		bodyStart: idx,
		warnings:  warnings,
	}
}

func parseHeaderBullet(h *Header, line string) {
	switch {
	case strings.HasPrefix(line, "- **Created**:"):
		h.Created = trimHeaderVal(line, "- **Created**:")
	case strings.HasPrefix(line, "- **Log Path**:"):
		h.LogPath = trimHeaderVal(line, "- **Log Path**:")
	case strings.HasPrefix(line, "- **Project**:"):
		h.Project = strings.Trim(trimHeaderVal(line, "- **Project**:"), "`")
	case strings.HasPrefix(line, "- **Branch**:"):
		h.Branch = strings.Trim(trimHeaderVal(line, "- **Branch**:"), "`")
	case strings.HasPrefix(line, "- **Context**:"):
		h.Context = strings.Trim(trimHeaderVal(line, "- **Context**:"), "`")
	case strings.HasPrefix(line, "- **Objective**:"):
		h.Objective = trimHeaderVal(line, "- **Objective**:")
	case strings.HasPrefix(line, "- **Status**:"):
		h.Status = LogStatus(trimHeaderVal(line, "- **Status**:"))
	}
}

func trimHeaderVal(line, prefix string) string {
	return strings.TrimSpace(strings.TrimPrefix(line, prefix))
}

func parseStepEntries(lines []string) ([]StepEntry, []string) {
	var entries []StepEntry
	var warnings []string
	var curr *StepEntry
	section := ""

	flush := func() {
		if curr != nil {
			entries = append(entries, *curr)
			curr = nil
		}
	}

	for _, raw := range lines {
		trimmed := strings.TrimSpace(raw)
		if m := stepHeadingRe.FindStringSubmatch(trimmed); m != nil {
			flush()
			num, _ := strconv.Atoi(m[2])
			curr = &StepEntry{
				Timestamp: m[1],
				Step:      num,
				Title:     strings.TrimSpace(m[3]),
				Actions:   []string{},
				Findings:  []string{},
				Decisions: []string{},
				NextSteps: []string{},
			}
			section = ""
			continue
		}
		if curr == nil || trimmed == "" || trimmed == "---" {
			continue
		}
		section = parseStepLine(curr, stepLineInput{raw: raw, trimmed: trimmed, section: section})
	}
	flush()
	return entries, warnings
}

type stepLineInput struct {
	raw     string
	trimmed string
	section string
}

type sectionItem struct {
	section string
	text    string
}

func parseStepLine(curr *StepEntry, in stepLineInput) string {
	switch {
	case strings.HasPrefix(in.trimmed, "- **Status**:"):
		curr.Status = StepStatus(trimHeaderVal(in.trimmed, "- **Status**:"))
		return ""
	case strings.HasPrefix(in.trimmed, "- **Executed Actions**:"):
		return "actions"
	case strings.HasPrefix(in.trimmed, "- **Verification & Findings**:"):
		return "findings"
	case strings.HasPrefix(in.trimmed, "- **Decisions & Rationale**:"):
		return "decisions"
	case strings.HasPrefix(in.trimmed, "- **Next Step**:"):
		return "next"
	case strings.HasPrefix(in.raw, "  - ") || strings.HasPrefix(in.raw, "\t- "):
		item := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(in.raw), "-"))
		if item != "(none)" && item != "" {
			appendSectionItem(curr, sectionItem{section: in.section, text: item})
		}
	}
	return in.section
}

func appendSectionItem(curr *StepEntry, item sectionItem) {
	switch item.section {
	case "actions":
		curr.Actions = append(curr.Actions, item.text)
	case "findings":
		curr.Findings = append(curr.Findings, item.text)
	case "decisions":
		curr.Decisions = append(curr.Decisions, item.text)
	case "next":
		curr.NextSteps = append(curr.NextSteps, item.text)
	}
}

// List discovers checkpoint log files under rootDir matching optional filters.
func (m *Manager) List(in ListInput) (ListOutput, error) {
	root, err := os.OpenRoot(m.rootDir)
	if err != nil {
		if os.IsNotExist(err) {
			return ListOutput{OK: true, Logs: []LogSummary{}}, nil
		}
		return ListOutput{}, err
	}
	defer func() { _ = root.Close() }()

	logs := []LogSummary{}
	_ = fs.WalkDir(root.FS(), ".", func(relPath string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if d.IsDir() {
			if d.Name() == ".locks" {
				return fs.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".md") || strings.HasPrefix(d.Name(), ".tmp_") {
			return nil
		}
		summary, ok := m.inspectLogSummary(logInspectContext{root: root, relPath: relPath, entry: d}, in)
		if ok {
			logs = append(logs, summary)
		}
		return nil
	})

	slices.SortFunc(logs, func(a, b LogSummary) int {
		return strings.Compare(b.UpdatedAt, a.UpdatedAt)
	})
	return ListOutput{OK: true, Logs: logs}, nil
}

type logInspectContext struct {
	root    *os.Root
	relPath string
	entry   fs.DirEntry
}

func (m *Manager) inspectLogSummary(ctx logInspectContext, in ListInput) (LogSummary, bool) {
	if in.DocType != "" && !strings.HasPrefix(ctx.entry.Name(), string(in.DocType)+"-") {
		return LogSummary{}, false
	}
	var raw []byte
	err := m.locks.WithRLock(lockKey(ctx.relPath), func() error {
		var readErr error
		raw, readErr = ctx.root.ReadFile(ctx.relPath)
		return readErr
	})
	if err != nil {
		return LogSummary{}, false
	}
	parsed := ParseLogMarkdown(string(raw))
	if !matchesListFilter(parsed.Header, in) {
		return LogSummary{}, false
	}
	absPath := filepath.Join(m.rootDir, filepath.FromSlash(ctx.relPath))
	updatedAt := parsed.Header.Created
	if info, statErr := ctx.entry.Info(); statErr == nil {
		updatedAt = info.ModTime().UTC().Format(time.RFC3339)
	}
	return LogSummary{
		Path:      absPath,
		DualPath:  pathutil.FormatDualPath(absPath),
		Project:   parsed.Header.Project,
		Branch:    parsed.Header.Branch,
		Context:   parsed.Header.Context,
		Title:     parsed.Header.Title,
		Status:    parsed.Header.Status,
		StepCount: len(parsed.Entries),
		UpdatedAt: updatedAt,
	}, true
}

func matchesListFilter(h Header, in ListInput) bool {
	return matchesSlugFilter(h.Project, in.Project) &&
		matchesSlugFilter(h.Branch, in.Branch) &&
		matchesSlugFilter(h.Context, in.Context)
}

func matchesSlugFilter(actual, filter string) bool {
	trimmed := strings.TrimSpace(filter)
	if trimmed == "" {
		return true
	}
	if actual == trimmed {
		return true
	}
	slug, err := pathutil.SanitizeSlug(trimmed)
	return err == nil && actual == slug
}
