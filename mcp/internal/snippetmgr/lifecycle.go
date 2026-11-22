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
	target, err := resolveSnippetTarget(targetRefInput{
		Name:    in.Name,
		Scope:   in.Scope,
		Project: in.Project,
		Branch:  in.Branch,
		WorkDir: in.WorkDir,
	})
	if err != nil {
		return VerifyOutput{}, err
	}

	var out VerifyOutput
	err = m.locks.WithRLock(target.lockKey(), func() error {
		res, err := m.loadTargetVersion(target, in.Version)
		if err != nil {
			return err
		}
		out = m.verifyVersionFile(res.Meta, *res.Version)
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
	target, err := resolveSnippetTarget(targetRefInput{
		Name:    in.Name,
		Scope:   in.Scope,
		Project: in.Project,
		Branch:  in.Branch,
		WorkDir: in.WorkDir,
	})
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
	err = m.locks.WithLock(target.lockKey(), func() error {
		res, err := m.loadTargetVersion(target, in.Version)
		if err != nil {
			return err
		}
		meta := res.Meta
		ver := res.Version
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
			MetaStoragePath: m.metaAbsPath(target),
		}
		return nil
	})
	return out, err
}

// Deprecate marks one or all active versions of a snippet as deprecated and strips execute permissions.
func (m *Manager) Deprecate(in DeprecateInput, now time.Time) (DeprecateOutput, error) {
	target, err := resolveSnippetTarget(targetRefInput{
		Name:    in.Name,
		Scope:   in.Scope,
		Project: in.Project,
		Branch:  in.Branch,
		WorkDir: in.WorkDir,
	})
	if err != nil {
		return DeprecateOutput{}, err
	}

	var out DeprecateOutput
	err = m.locks.WithLock(target.lockKey(), func() error {
		loaded, err := m.loadMeta(target)
		if err != nil {
			return err
		}
		if loaded == nil {
			return fmt.Errorf("snippet %q not found under %s/%s/%s", target.Name, target.Project, target.Branch, target.Scope)
		}
		meta := *loaded
		ts := now.UTC().Format(time.RFC3339)
		deprecated := meta.markVersionsDeprecated(in.Version, ts)
		if in.Version > 0 && len(deprecated) == 0 {
			return fmt.Errorf("active version %d not found for snippet %q", in.Version, target.Name)
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
			Name:               target.Name,
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
		pb, err := pathutil.ResolveProjectBranch(pathutil.ProjectBranchInput{
			Branch:  "placeholder",
			WorkDir: in.WorkDir,
		})
		if err != nil {
			return CleanupOutput{}, err
		}
		proj = pb.Project
	}
	flt, err := newSnippetFilter(snippetFilterInput{
		Project: proj,
		Branch:  in.Branch,
		Scope:   in.Scope,
		Name:    in.Name,
	})
	if err != nil {
		return CleanupOutput{}, err
	}

	metaRelPaths, err := m.findMetaFiles()
	if err != nil {
		return CleanupOutput{}, err
	}

	opts := cleanupOpts{
		IncludeActiveOneOff: in.IncludeActiveOneOff,
		DryRun:              in.DryRun,
		Timestamp:           now.UTC().Format(time.RFC3339),
	}
	var removed []string
	for _, relMeta := range metaRelPaths {
		meta, readErr := store.ReadJSONRoot[SnippetMeta](m.rootDir, relMeta)
		if readErr != nil || !flt.matches(meta) {
			continue
		}
		cleaned, err := m.cleanupSingleSnippet(meta, opts)
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

type cleanupOpts struct {
	IncludeActiveOneOff bool
	DryRun              bool
	Timestamp           string
}

func (m *Manager) cleanupSingleSnippet(meta SnippetMeta, opts cleanupOpts) ([]string, error) {
	target := snippetTarget{
		Name:    meta.Name,
		Scope:   meta.Scope,
		Project: meta.Project,
		Branch:  meta.Branch,
	}
	var removed []string
	err := m.locks.WithLock(target.lockKey(), func() error {
		loaded, err := m.loadMeta(target)
		if err != nil || loaded == nil {
			return err
		}
		cur := *loaded
		changed := false
		for i := range cur.Versions {
			v := &cur.Versions[i]
			shouldRemove := v.Status == StatusDeprecated || (opts.IncludeActiveOneOff && cur.Scope == ScopeOneOff && v.Status == StatusActive)
			if !shouldRemove {
				continue
			}
			relPath := filepath.Join(cur.Project, cur.Branch, string(cur.Scope), v.Filename)
			removed = append(removed, v.Path)
			if opts.DryRun {
				continue
			}
			if err := m.removeWithinRoot(relPath); err != nil {
				return err
			}
			v.Status = StatusCleaned
			v.CleanedAt = opts.Timestamp
			changed = true
		}
		if changed {
			cur.UpdatedAt = opts.Timestamp
			return m.saveMeta(cur)
		}
		return nil
	})
	return removed, err
}
