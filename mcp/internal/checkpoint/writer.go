package checkpoint

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/berquerant/agent-skills/mcp/internal/guard"
	"github.com/berquerant/agent-skills/mcp/internal/pathutil"
	"github.com/berquerant/agent-skills/mcp/internal/store"
)

const (
	maxEntryBytes   = 16 * 1024
	fileLockTimeout = 10 * time.Second
)

// Manager handles checkpoint log initialization, append-only milestone entries, and context recovery under rootDir.
type Manager struct {
	rootDir string
	locks   *store.LockPool
}

// NewManager creates a checkpoint Manager rooted at rootDir.
func NewManager(rootDir string) (*Manager, error) {
	if err := guard.CheckPathAllowed(rootDir); err != nil {
		return nil, err
	}
	absRoot, err := filepath.Abs(rootDir)
	if err != nil {
		return nil, fmt.Errorf("resolve checkpoint root: %w", err)
	}
	if err := os.MkdirAll(absRoot, 0o700); err != nil {
		return nil, fmt.Errorf("create checkpoint root: %w", err)
	}
	return &Manager{
		rootDir: absRoot,
		locks:   store.NewLockPool(absRoot, fileLockTimeout),
	}, nil
}

func (m *Manager) withLockedRoot(relPath string, fn func(root *os.Root) error) error {
	return m.locks.WithLock(relPath, func() error {
		return m.withRoot(fn)
	})
}

func (m *Manager) withRLockedRoot(relPath string, fn func(root *os.Root) error) error {
	return m.locks.WithRLock(relPath, func() error {
		return m.withRoot(fn)
	})
}

func (m *Manager) withRoot(fn func(root *os.Root) error) error {
	root, err := os.OpenRoot(m.rootDir)
	if err != nil {
		return err
	}
	defer func() { _ = root.Close() }()
	return fn(root)
}

// Init resolves the generation path and, unless in.ProposeOnly is true, creates the checkpoint log header.
func (m *Manager) Init(in InitInput, now time.Time) (InitOutput, error) {
	if strings.TrimSpace(in.Title) == "" || strings.TrimSpace(in.Objective) == "" {
		return InitOutput{}, fmt.Errorf("title and objective are required")
	}
	if err := checkNoSecrets(in.Title, in.Objective); err != nil {
		return InitOutput{}, err
	}
	docType, err := NormalizeDocType(in.DocType)
	if err != nil {
		return InitOutput{}, err
	}
	ctxSlug, err := pathutil.SanitizeSlug(in.Context)
	if err != nil {
		return InitOutput{}, fmt.Errorf("invalid context: %w", err)
	}
	pb, err := pathutil.ResolveProjectBranch(pathutil.ProjectBranchInput{
		Project: in.Project,
		Branch:  in.Branch,
		WorkDir: in.WorkDir,
	})
	if err != nil {
		return InitOutput{}, err
	}

	latestGen, _ := FindLatestGeneration(GenerationQuery{
		RootDir: m.rootDir,
		Project: pb.Project,
		Branch:  pb.Branch,
		Context: ctxSlug,
		DocType: docType,
	})
	tsSlug := now.UTC().Format("20060102-150405")
	relPath := filepath.Join(pb.Project, pb.Branch, ctxSlug, fmt.Sprintf("%s-%s.md", docType, tsSlug))
	absPath := filepath.Join(m.rootDir, relPath)

	out := InitOutput{
		OK:               true,
		Path:             absPath,
		DualPath:         pathutil.FormatDualPath(absPath),
		Project:          pb.Project,
		Branch:           pb.Branch,
		Context:          ctxSlug,
		DocType:          string(docType),
		ProposeOnly:      in.ProposeOnly,
		LatestGeneration: latestGen,
	}
	if in.ProposeOnly {
		return out, nil
	}

	if err := m.writeHeaderFile(headerWriteParams{
		relPath:    relPath,
		absPath:    absPath,
		projSlug:   pb.Project,
		branchSlug: pb.Branch,
		ctxSlug:    ctxSlug,
		in:         in,
		now:        now,
	}); err != nil {
		return InitOutput{}, err
	}
	out.Created = true
	return out, nil
}

type headerWriteParams struct {
	relPath    string
	absPath    string
	projSlug   string
	branchSlug string
	ctxSlug    string
	in         InitInput
	now        time.Time
}

func (m *Manager) writeHeaderFile(p headerWriteParams) error {
	return m.withLockedRoot(p.relPath, func(root *os.Root) error {
		if err := os.MkdirAll(filepath.Dir(p.absPath), 0o700); err != nil {
			return fmt.Errorf("create context dir: %w", err)
		}
		if _, err := root.Stat(p.relPath); err == nil {
			return fmt.Errorf("checkpoint generation file already exists: %s", p.absPath)
		}

		var b strings.Builder
		_, _ = fmt.Fprintf(&b, "# Task Checkpoint Log: %s\n\n", sanitizeSingleLine(p.in.Title))
		_, _ = fmt.Fprintf(&b, "- **Created**: %s\n", p.now.UTC().Format("2006-01-02 15:04:05"))
		_, _ = fmt.Fprintf(&b, "- **Log Path**: %s\n", pathutil.FormatDualPath(p.absPath))
		_, _ = fmt.Fprintf(&b, "- **Project**: `%s`\n", p.projSlug)
		_, _ = fmt.Fprintf(&b, "- **Branch**: `%s`\n", p.branchSlug)
		_, _ = fmt.Fprintf(&b, "- **Context**: `%s`\n", p.ctxSlug)
		_, _ = fmt.Fprintf(&b, "- **Objective**: %s\n", sanitizeSingleLine(p.in.Objective))
		_, _ = fmt.Fprintf(&b, "- **Status**: %s\n\n---\n", LogStatusInProgress)

		return store.WriteAtomic(p.absPath, []byte(b.String()), 0o600)
	})
}

// Append appends a milestone entry to the end of an existing checkpoint log file using O_APPEND.
func (m *Manager) Append(in AppendInput, now time.Time) (AppendOutput, error) {
	if err := validateAppendInput(in); err != nil {
		return AppendOutput{}, err
	}
	resolved, err := ResolveRelativeWithinRoot(m.rootDir, in.Path)
	if err != nil {
		return AppendOutput{}, err
	}

	var out AppendOutput
	err = m.withLockedRoot(resolved.Rel, func(root *os.Root) error {
		existingBytes, err := root.ReadFile(resolved.Rel)
		if err != nil {
			return fmt.Errorf("read existing checkpoint log: %w", err)
		}

		stepNum := in.Step
		if stepNum <= 0 {
			parsed := ParseLogMarkdown(string(existingBytes))
			stepNum = len(parsed.Entries) + 1
			for _, e := range parsed.Entries {
				if e.Step >= stepNum {
					stepNum = e.Step + 1
				}
			}
		}

		tsStr := now.UTC().Format("2006-01-02 15:04:05")
		block := formatStepEntryMarkdown(stepEntryParams{
			step:  stepNum,
			tsStr: tsStr,
			in:    in,
		})
		if len(block) > maxEntryBytes {
			return fmt.Errorf("checkpoint entry size (%d bytes) exceeds maximum (%d bytes)", len(block), maxEntryBytes)
		}

		f, err := root.OpenFile(resolved.Rel, os.O_WRONLY|os.O_APPEND, 0o600)
		if err != nil {
			return fmt.Errorf("open checkpoint log for append: %w", err)
		}
		defer func() { _ = f.Close() }()

		if _, err := f.WriteString(block); err != nil {
			return fmt.Errorf("append checkpoint entry: %w", err)
		}
		if err := f.Sync(); err != nil {
			return fmt.Errorf("sync checkpoint log: %w", err)
		}

		out = AppendOutput{
			OK:        true,
			Path:      resolved.Abs,
			DualPath:  pathutil.FormatDualPath(resolved.Abs),
			Step:      stepNum,
			Timestamp: tsStr,
			Status:    in.Status,
		}
		return nil
	})
	return out, err
}

func validateAppendInput(in AppendInput) error {
	if strings.TrimSpace(in.Title) == "" {
		return fmt.Errorf("step title is required")
	}
	switch in.Status {
	case StepStatusCompleted, StepStatusFailed, StepStatusBlocked, StepStatusAdjusted:
	default:
		return fmt.Errorf("invalid step status %q (expected COMPLETED, FAILED, BLOCKED, or ADJUSTED)", in.Status)
	}
	allTexts := []string{in.Title}
	allTexts = append(allTexts, in.Actions...)
	allTexts = append(allTexts, in.Findings...)
	allTexts = append(allTexts, in.Decisions...)
	allTexts = append(allTexts, in.NextSteps...)
	return checkNoSecrets(allTexts...)
}

type stepEntryParams struct {
	step  int
	tsStr string
	in    AppendInput
}

type bulletSection struct {
	heading string
	items   []string
}

func formatStepEntryMarkdown(p stepEntryParams) string {
	var b strings.Builder
	_, _ = fmt.Fprintf(&b, "\n## [%s] Step %d: %s\n\n", p.tsStr, p.step, sanitizeSingleLine(p.in.Title))
	_, _ = fmt.Fprintf(&b, "- **Status**: %s\n", p.in.Status)
	writeBulletSection(&b, bulletSection{heading: "Executed Actions", items: p.in.Actions})
	writeBulletSection(&b, bulletSection{heading: "Verification & Findings", items: p.in.Findings})
	writeBulletSection(&b, bulletSection{heading: "Decisions & Rationale", items: p.in.Decisions})
	writeBulletSection(&b, bulletSection{heading: "Next Step", items: p.in.NextSteps})
	b.WriteString("\n---\n")
	return b.String()
}

func writeBulletSection(b *strings.Builder, sec bulletSection) {
	_, _ = fmt.Fprintf(b, "- **%s**:\n", sec.heading)
	if len(sec.items) == 0 {
		b.WriteString("  - (none)\n")
		return
	}
	for _, item := range sec.items {
		_, _ = fmt.Fprintf(b, "  - %s\n", sanitizeSingleLine(item))
	}
}

func sanitizeSingleLine(s string) string {
	s = strings.ReplaceAll(s, "\r\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	return strings.TrimSpace(s)
}

// SetStatus updates only the - **Status**: line in the header block of an existing checkpoint log.
func (m *Manager) SetStatus(in SetStatusInput) (SetStatusOutput, error) {
	switch in.Status {
	case LogStatusInProgress, LogStatusCompleted, LogStatusAborted:
	default:
		return SetStatusOutput{}, fmt.Errorf("invalid log status %q (expected IN_PROGRESS, COMPLETED, or ABORTED)", in.Status)
	}
	resolved, err := ResolveRelativeWithinRoot(m.rootDir, in.Path)
	if err != nil {
		return SetStatusOutput{}, err
	}

	var out SetStatusOutput
	err = m.withLockedRoot(resolved.Rel, func(root *os.Root) error {
		raw, err := root.ReadFile(resolved.Rel)
		if err != nil {
			return fmt.Errorf("read checkpoint log: %w", err)
		}
		updated, replaced := replaceHeaderStatusLine(string(raw), in.Status)
		if !replaced {
			return fmt.Errorf("header Status line not found in %s", resolved.Abs)
		}
		if err := store.WriteAtomic(resolved.Abs, []byte(updated), 0o600); err != nil {
			return err
		}
		out = SetStatusOutput{
			OK:       true,
			Path:     resolved.Abs,
			DualPath: pathutil.FormatDualPath(resolved.Abs),
			Status:   in.Status,
		}
		return nil
	})
	return out, err
}

func replaceHeaderStatusLine(content string, newStatus LogStatus) (string, bool) {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if strings.TrimSpace(line) == "---" {
			break
		}
		if strings.HasPrefix(strings.TrimSpace(line), "- **Status**:") {
			lines[i] = fmt.Sprintf("- **Status**: %s", newStatus)
			return strings.Join(lines, "\n"), true
		}
	}
	return content, false
}

func checkNoSecrets(texts ...string) error {
	if err := guard.CheckNoSecrets(texts...); err != nil {
		return fmt.Errorf("secret or credential token pattern detected in checkpoint content; omit sensitive values per log-format.md §5: %w", err)
	}
	return nil
}
