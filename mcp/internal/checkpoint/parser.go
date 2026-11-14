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
)

var stepHeadingRe = regexp.MustCompile(`^##\s+\[([^\]]+)\]\s+Step\s+([0-9]+):\s+(.+)$`)

// Resume resolves a checkpoint log (by explicit path or latest generation for context) and parses its recovery state.
func (m *Manager) Resume(in ResumeInput) (ResumeOutput, error) {
	absPath, err := m.resolveResumePath(in)
	if err != nil {
		return ResumeOutput{}, err
	}
	relPath, absPath, err := ResolveRelativeWithinRoot(m.rootDir, absPath)
	if err != nil {
		return ResumeOutput{}, err
	}

	lock := m.lockFor(relPath)
	if err := lock.RLock(); err != nil {
		return ResumeOutput{}, err
	}
	defer func() { _ = lock.Unlock() }()

	root, err := os.OpenRoot(m.rootDir)
	if err != nil {
		return ResumeOutput{}, err
	}
	defer func() { _ = root.Close() }()

	raw, err := root.ReadFile(relPath)
	if err != nil {
		return ResumeOutput{}, fmt.Errorf("read checkpoint log %s: %w", absPath, err)
	}

	header, entries, warnings := ParseLogMarkdown(string(raw))
	return buildResumeOutput(absPath, header, entries, warnings), nil
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
	ctxSlug, err := SanitizeSlug(in.Context)
	if err != nil {
		return "", err
	}
	projSlug, branchSlug, err := ResolveProjectAndBranch(in.Project, in.Branch, in.WorkDir)
	if err != nil {
		return "", err
	}
	latest, err := FindLatestGeneration(m.rootDir, projSlug, branchSlug, ctxSlug, docType)
	if err != nil {
		return "", err
	}
	if latest == "" {
		return "", fmt.Errorf("no %s log found for %s/%s/%s", docType, projSlug, branchSlug, ctxSlug)
	}
	return latest, nil
}

func buildResumeOutput(absPath string, header Header, entries []StepEntry, warnings []string) ResumeOutput {
	completed := []string{}
	unresolved := []StepEntry{}
	var lastCompleted *StepEntry
	var latestEntry *StepEntry
	nextSteps := []string{}

	for i := range entries {
		e := entries[i]
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
	if warnings == nil {
		warnings = []string{}
	}

	return ResumeOutput{
		OK:               true,
		Path:             absPath,
		DualPath:         FormatDualPath(absPath),
		Header:           header,
		TotalSteps:       len(entries),
		CompletedSteps:   completed,
		UnresolvedSteps:  unresolved,
		LastCompleted:    lastCompleted,
		LatestEntry:      latestEntry,
		CurrentNextSteps: nextSteps,
		ParseWarnings:    warnings,
	}
}

// ParseLogMarkdown parses the header block and milestone entries from a checkpoint log Markdown document.
func ParseLogMarkdown(content string) (Header, []StepEntry, []string) {
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	header, bodyStart, warnings := parseHeaderSection(lines)
	entries, stepWarnings := parseStepEntries(lines[bodyStart:])
	warnings = append(warnings, stepWarnings...)
	return header, entries, warnings
}

func parseHeaderSection(lines []string) (Header, int, []string) {
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
	return h, idx, warnings
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
		section = parseStepLine(curr, raw, trimmed, section)
	}
	flush()
	return entries, warnings
}

func parseStepLine(curr *StepEntry, raw, trimmed, section string) string {
	switch {
	case strings.HasPrefix(trimmed, "- **Status**:"):
		curr.Status = StepStatus(trimHeaderVal(trimmed, "- **Status**:"))
		return ""
	case strings.HasPrefix(trimmed, "- **Executed Actions**:"):
		return "actions"
	case strings.HasPrefix(trimmed, "- **Verification & Findings**:"):
		return "findings"
	case strings.HasPrefix(trimmed, "- **Decisions & Rationale**:"):
		return "decisions"
	case strings.HasPrefix(trimmed, "- **Next Step**:"):
		return "next"
	case strings.HasPrefix(raw, "  - ") || strings.HasPrefix(raw, "\t- "):
		item := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(raw), "-"))
		if item != "(none)" && item != "" {
			appendSectionItem(curr, section, item)
		}
	}
	return section
}

func appendSectionItem(curr *StepEntry, section, item string) {
	switch section {
	case "actions":
		curr.Actions = append(curr.Actions, item)
	case "findings":
		curr.Findings = append(curr.Findings, item)
	case "decisions":
		curr.Decisions = append(curr.Decisions, item)
	case "next":
		curr.NextSteps = append(curr.NextSteps, item)
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
		if walkErr != nil || d.IsDir() || !strings.HasSuffix(d.Name(), ".md") || strings.HasPrefix(d.Name(), ".tmp_") {
			return nil
		}
		summary, ok := m.inspectLogSummary(root, relPath, d, in)
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

func (m *Manager) inspectLogSummary(root *os.Root, relPath string, d fs.DirEntry, in ListInput) (LogSummary, bool) {
	if in.DocType != "" && !strings.HasPrefix(d.Name(), string(in.DocType)+"-") {
		return LogSummary{}, false
	}
	raw, err := root.ReadFile(relPath)
	if err != nil {
		return LogSummary{}, false
	}
	header, entries, _ := ParseLogMarkdown(string(raw))
	if !matchesListFilter(header, in) {
		return LogSummary{}, false
	}
	absPath := filepath.Join(m.rootDir, filepath.FromSlash(relPath))
	updatedAt := header.Created
	if info, statErr := d.Info(); statErr == nil {
		updatedAt = info.ModTime().UTC().Format(time.RFC3339)
	}
	return LogSummary{
		Path:      absPath,
		DualPath:  FormatDualPath(absPath),
		Project:   header.Project,
		Branch:    header.Branch,
		Context:   header.Context,
		Title:     header.Title,
		Status:    header.Status,
		StepCount: len(entries),
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
	slug, err := SanitizeSlug(trimmed)
	return err == nil && actual == slug
}
