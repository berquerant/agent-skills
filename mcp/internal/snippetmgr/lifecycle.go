package snippetmgr

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/berquerant/agent-skills/mcp/internal/guard"
	"github.com/berquerant/agent-skills/mcp/internal/pathutil"
	"github.com/berquerant/agent-skills/mcp/internal/store"
)

const maxInvocationHistory = 200

// Verify checks that a snippet version is active, executable, write-protected, and matches its SHA-256 digest and inode.
func (m *Manager) Verify(in VerifyInput) (VerifyOutput, error) {
	name, scope, proj, branch, err := resolveSnippetTarget(in.Name, in.Scope, in.Project, in.Branch, in.WorkDir)
	if err != nil {
		return VerifyOutput{}, err
	}

	var out VerifyOutput
	err = m.locks.WithRLock(m.lockKey(proj, branch, scope, name), func() error {
		meta, ver, err := m.loadTargetVersion(proj, branch, scope, name, in.Version)
		if err != nil {
			return err
		}
		out = m.verifyVersionFile(meta, *ver)
		return nil
	})
	return out, err
}

func (m *Manager) verifyVersionFile(meta SnippetMeta, ver VersionRecord) VerifyOutput {
	out := VerifyOutput{
		Name:           meta.Name,
		Version:        ver.Version,
		Status:         ver.Status,
		Path:           ver.Path,
		DualPath:       ver.DualPath,
		ExpectedSHA256: ver.SHA256,
		ExpectedInode:  ver.Inode,
		Issues:         []string{},
	}
	if ver.Status != StatusActive {
		out.Issues = append(out.Issues, fmt.Sprintf("version %d status is %q (expected active)", ver.Version, ver.Status))
	}

	relPath := filepath.Join(meta.Project, meta.Branch, string(meta.Scope), ver.Filename)
	if err := m.withRoot(func(root *os.Root) error {
		info, statErr := root.Stat(relPath)
		if statErr != nil {
			out.Issues = append(out.Issues, fmt.Sprintf("snippet file missing or unreadable: %v", statErr))
			return nil
		}
		mode := info.Mode().Perm()
		out.Executable = mode&ownerExecBit != 0
		out.Immutable = mode&anyWriteBits == 0
		if !out.Executable {
			out.Issues = append(out.Issues, fmt.Sprintf("snippet file is not executable (mode %04o)", mode))
		}
		if !out.Immutable {
			out.Issues = append(out.Issues, fmt.Sprintf("snippet file is writable (mode %04o; expected write-protected 0500)", mode))
		}
		if ino, ok := fileInode(info); ok {
			out.ActualInode = ino
			if ver.Inode > 0 && ino != ver.Inode {
				out.Issues = append(out.Issues, fmt.Sprintf("inode mismatch (expected %d, got %d; file replaced or moved)", ver.Inode, ino))
			}
		}

		raw, readErr := root.ReadFile(relPath)
		if readErr != nil {
			out.Issues = append(out.Issues, fmt.Sprintf("read snippet file failed: %v", readErr))
			return nil
		}
		out.ActualSHA256 = sha256Hex(raw)
		if out.ActualSHA256 != ver.SHA256 {
			out.Issues = append(out.Issues, "SHA-256 digest mismatch (file modified outside snippet-manager)")
		}
		return nil
	}); err != nil {
		out.Issues = append(out.Issues, fmt.Sprintf("open snippet root failed: %v", err))
	}

	out.Valid = len(out.Issues) == 0
	return out
}

// RecordInvocation appends an execution record to the snippet metadata ledger.
func (m *Manager) RecordInvocation(in RecordInvocationInput, now time.Time) (RecordInvocationOutput, error) {
	if in.DurationMs < 0 {
		return RecordInvocationOutput{}, fmt.Errorf("duration_ms must be >= 0 (got %d)", in.DurationMs)
	}
	name, scope, proj, branch, err := resolveSnippetTarget(in.Name, in.Scope, in.Project, in.Branch, in.WorkDir)
	if err != nil {
		return RecordInvocationOutput{}, err
	}
	for _, a := range in.Args {
		if err := guard.CheckNoSecrets(a); err != nil {
			return RecordInvocationOutput{}, fmt.Errorf("invocation arg failed secret check: %w", err)
		}
	}
	if in.Note != "" {
		if err := guard.CheckNoSecrets(in.Note); err != nil {
			return RecordInvocationOutput{}, fmt.Errorf("invocation note failed secret check: %w", err)
		}
	}

	var out RecordInvocationOutput
	err = m.locks.WithLock(m.lockKey(proj, branch, scope, name), func() error {
		meta, ver, err := m.loadTargetVersion(proj, branch, scope, name, in.Version)
		if err != nil {
			return err
		}
		ts := now.UTC().Format(time.RFC3339)
		rec := InvocationRecord{
			Version:    ver.Version,
			ExecutedAt: ts,
			Args:       slices.Clone(in.Args),
			ExitCode:   in.ExitCode,
			DurationMs: in.DurationMs,
			Note:       strings.TrimSpace(in.Note),
		}
		meta.Invocations = append(meta.Invocations, rec)
		if len(meta.Invocations) > maxInvocationHistory {
			meta.Invocations = meta.Invocations[len(meta.Invocations)-maxInvocationHistory:]
		}
		meta.UpdatedAt = ts
		if err := m.saveMeta(meta); err != nil {
			return err
		}
		out = RecordInvocationOutput{
			Recorded:        true,
			Name:            meta.Name,
			Version:         ver.Version,
			TotalRecorded:   len(meta.Invocations),
			LastInvocation:  rec,
			MetaStoragePath: m.metaAbsPath(proj, branch, scope, name),
		}
		return nil
	})
	return out, err
}

// Deprecate marks one or all active versions of a snippet as deprecated and strips execute permissions.
func (m *Manager) Deprecate(in DeprecateInput, now time.Time) (DeprecateOutput, error) {
	name, scope, proj, branch, err := resolveSnippetTarget(in.Name, in.Scope, in.Project, in.Branch, in.WorkDir)
	if err != nil {
		return DeprecateOutput{}, err
	}

	var out DeprecateOutput
	err = m.locks.WithLock(m.lockKey(proj, branch, scope, name), func() error {
		meta, exists, err := m.loadMeta(proj, branch, scope, name)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("snippet %q not found under %s/%s/%s", name, proj, branch, scope)
		}
		ts := now.UTC().Format(time.RFC3339)
		deprecated := markVersionsDeprecated(&meta, in.Version, ts)
		if in.Version > 0 && len(deprecated) == 0 {
			return fmt.Errorf("active version %d not found for snippet %q", in.Version, name)
		}
		if len(deprecated) > 0 {
			meta.UpdatedAt = ts
			if err := m.saveMeta(meta); err != nil {
				return err
			}
			if err := m.chmodDeprecatedFiles(meta, deprecated); err != nil {
				return err
			}
		}
		if deprecated == nil {
			deprecated = []int{}
		}
		out = DeprecateOutput{
			Name:               name,
			DeprecatedVersions: deprecated,
		}
		return nil
	})
	return out, err
}

// Cleanup removes deprecated snippet files (and optionally active one-off files) while preserving metadata history.
func (m *Manager) Cleanup(in CleanupInput, now time.Time) (CleanupOutput, error) {
	proj := in.Project
	if !in.AllProjects && strings.TrimSpace(proj) == "" {
		var err error
		if proj, _, err = pathutil.ResolveProjectAndBranch("", "placeholder", in.WorkDir); err != nil {
			return CleanupOutput{}, err
		}
	}
	flt, err := newSnippetFilter(proj, in.Branch, in.Scope, in.Name, "")
	if err != nil {
		return CleanupOutput{}, err
	}

	metaRelPaths, err := m.findMetaFiles()
	if err != nil {
		return CleanupOutput{}, err
	}

	ts := now.UTC().Format(time.RFC3339)
	var removed []string
	for _, relMeta := range metaRelPaths {
		meta, readErr := store.ReadJSONRoot[SnippetMeta](m.rootDir, relMeta)
		if readErr != nil || !flt.matches(meta) {
			continue
		}
		cleaned, err := m.cleanupSingleSnippet(meta, in.IncludeActiveOneOff, in.DryRun, ts)
		if err != nil {
			return CleanupOutput{}, err
		}
		removed = append(removed, cleaned...)
	}

	if removed == nil {
		removed = []string{}
	}
	return CleanupOutput{
		DryRun:       in.DryRun,
		RemovedFiles: removed,
		RemovedCount: len(removed),
	}, nil
}

func (m *Manager) cleanupSingleSnippet(meta SnippetMeta, includeActiveOneOff, dryRun bool, ts string) ([]string, error) {
	var removed []string
	err := m.locks.WithLock(m.lockKey(meta.Project, meta.Branch, meta.Scope, meta.Name), func() error {
		cur, exists, err := m.loadMeta(meta.Project, meta.Branch, meta.Scope, meta.Name)
		if err != nil || !exists {
			return err
		}
		changed := false
		for i := range cur.Versions {
			v := &cur.Versions[i]
			shouldRemove := v.Status == StatusDeprecated || (includeActiveOneOff && cur.Scope == ScopeOneOff && v.Status == StatusActive)
			if !shouldRemove {
				continue
			}
			relPath := filepath.Join(cur.Project, cur.Branch, string(cur.Scope), v.Filename)
			removed = append(removed, v.Path)
			if dryRun {
				continue
			}
			if err := m.removeWithinRoot(relPath); err != nil {
				return err
			}
			v.Status = StatusCleaned
			v.CleanedAt = ts
			changed = true
		}
		if changed {
			cur.UpdatedAt = ts
			return m.saveMeta(cur)
		}
		return nil
	})
	return removed, err
}
