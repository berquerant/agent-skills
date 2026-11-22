package snippetmgr

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/berquerant/agent-skills/mcp/internal/guard"
	"github.com/berquerant/agent-skills/mcp/internal/pathutil"
	"github.com/berquerant/agent-skills/mcp/internal/snippet"
	"github.com/berquerant/agent-skills/mcp/internal/store"
)

const (
	activeFilePerm     os.FileMode = 0o500
	deprecatedFilePerm os.FileMode = 0o400
	metaFilePerm       os.FileMode = 0o600
	dirPerm            os.FileMode = 0o700
	defaultLockTimeout             = 10 * time.Second
)

// Manager coordinates immutable snippet storage, versioning, verification, and lifecycle operations.
type Manager struct {
	rootDir string
	locks   *store.LockPool
}

// NewManager creates a Manager rooted at rootDir after validating path safety.
func NewManager(rootDir string) (*Manager, error) {
	trimmed := strings.TrimSpace(rootDir)
	if trimmed == "" {
		return nil, fmt.Errorf("snippet root directory is required")
	}
	if err := guard.CheckPathAllowed(trimmed); err != nil {
		return nil, err
	}
	absRoot, err := filepath.Abs(trimmed)
	if err != nil {
		return nil, fmt.Errorf("resolve snippet root: %w", err)
	}
	return &Manager{
		rootDir: absRoot,
		locks:   store.NewLockPool(filepath.Join(absRoot, ".locks"), defaultLockTimeout),
	}, nil
}

// Create validates, audits, and saves an immutable versioned snippet under the XDG snippet root.
func (m *Manager) Create(in CreateInput, now time.Time) (CreateOutput, error) {
	name, err := ValidateName(in.Name)
	if err != nil {
		return CreateOutput{}, err
	}
	scope, err := NormalizeScope(in.Scope)
	if err != nil {
		return CreateOutput{}, err
	}
	lang, ext, err := NormalizeLanguage(in.Language)
	if err != nil {
		return CreateOutput{}, err
	}
	if err := ValidateContent(in.Code, lang); err != nil {
		return CreateOutput{}, err
	}
	if in.Purpose != "" {
		if err := guard.CheckNoSecrets(in.Purpose); err != nil {
			return CreateOutput{}, fmt.Errorf("purpose failed secret check: %w", err)
		}
	}

	proj, branch, err := pathutil.ResolveProjectAndBranch(in.Project, in.Branch, in.WorkDir)
	if err != nil {
		return CreateOutput{}, err
	}

	audit := snippet.Analyze(in.Code, lang)
	if !audit.OK {
		return CreateOutput{}, fmt.Errorf("snippet audit failed: %s", audit.ParseError)
	}
	if audit.Level >= 3 && !in.AllowComplex {
		return CreateOutput{
			Saved:          false,
			ActionRequired: "simplify_or_consult",
			Name:           name,
			Project:        proj,
			Branch:         branch,
			Scope:          scope,
			Language:       lang,
			Audit:          audit,
		}, nil
	}

	return m.saveVersionedSnippet(createParams{
		name:         name,
		scope:        scope,
		lang:         lang,
		ext:          ext,
		proj:         proj,
		branch:       branch,
		code:         in.Code,
		purpose:      strings.TrimSpace(in.Purpose),
		allowComplex: in.AllowComplex,
		audit:        audit,
		now:          now,
	})
}

type createParams struct {
	name         string
	scope        Scope
	lang         string
	ext          string
	proj         string
	branch       string
	code         string
	purpose      string
	allowComplex bool
	audit        snippet.AuditReport
	now          time.Time
}

func (m *Manager) saveVersionedSnippet(p createParams) (CreateOutput, error) {
	lockKey := m.lockKey(p.proj, p.branch, p.scope, p.name)
	var out CreateOutput

	err := m.locks.WithLock(lockKey, func() error {
		meta, exists, err := m.loadMeta(p.proj, p.branch, p.scope, p.name)
		if err != nil {
			return err
		}
		ts := p.now.UTC().Format(time.RFC3339)
		if !exists {
			meta = SnippetMeta{
				Name:      p.name,
				Project:   p.proj,
				Branch:    p.branch,
				Scope:     p.scope,
				Language:  p.lang,
				Extension: p.ext,
				Purpose:   p.purpose,
				CreatedAt: ts,
			}
		} else {
			if p.purpose != "" {
				meta.Purpose = p.purpose
			}
			meta.Language = p.lang
			meta.Extension = p.ext
		}

		nextVer := meta.LatestVersion + 1
		filename := fmt.Sprintf("%s_v%d%s", p.name, nextVer, p.ext)
		relPath := filepath.Join(p.proj, p.branch, string(p.scope), filename)
		absPath := filepath.Join(m.rootDir, relPath)

		codeBytes := []byte(p.code)
		if err := m.writeImmutableFile(relPath, codeBytes, activeFilePerm); err != nil {
			return err
		}

		deprecatedVers, err := m.deprecateActiveVersions(&meta, ts)
		if err != nil {
			return err
		}

		digest := sha256Hex(codeBytes)
		dualPath := pathutil.FormatDualPath(absPath)
		rec := VersionRecord{
			Version:      nextVer,
			Filename:     filename,
			Path:         absPath,
			DualPath:     dualPath,
			SHA256:       digest,
			Status:       StatusActive,
			AuditLevel:   p.audit.Level,
			AuditRating:  p.audit.Rating,
			AuditDrivers: p.audit.Drivers,
			AllowComplex: p.allowComplex,
			CreatedAt:    ts,
			Audit:        p.audit,
		}
		meta.LatestVersion = nextVer
		meta.UpdatedAt = ts
		meta.Versions = append(meta.Versions, rec)

		if err := m.saveMeta(meta); err != nil {
			return err
		}

		out = CreateOutput{
			Saved:                true,
			Name:                 p.name,
			Project:              p.proj,
			Branch:               p.branch,
			Scope:                p.scope,
			Language:             p.lang,
			Version:              nextVer,
			Path:                 absPath,
			DualPath:             dualPath,
			SHA256:               digest,
			DeprecatedVersions:   deprecatedVers,
			Audit:                p.audit,
			PreExecutionNotice:   BuildPreExecutionNotice(dualPath, absPath, meta.Purpose, p.audit.Level, p.audit.Rating),
			DirectInvocationHint: absPath,
		}
		return nil
	})
	return out, err
}

// Get retrieves metadata and the requested (or latest) version record for a snippet.
func (m *Manager) Get(in GetInput) (GetOutput, error) {
	name, scope, proj, branch, err := resolveSnippetTarget(in.Name, in.Scope, in.Project, in.Branch, in.WorkDir)
	if err != nil {
		return GetOutput{}, err
	}

	var out GetOutput
	err = m.locks.WithRLock(m.lockKey(proj, branch, scope, name), func() error {
		meta, exists, err := m.loadMeta(proj, branch, scope, name)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("snippet %q not found under %s/%s/%s", name, proj, branch, scope)
		}
		ver := selectVersion(&meta, in.Version)
		if ver == nil {
			return fmt.Errorf("snippet %q version %d not found", name, in.Version)
		}
		verCopy := *ver
		out = GetOutput{
			Meta:            meta,
			SelectedVersion: &verCopy,
		}
		return nil
	})
	return out, err
}

// List scans managed snippets matching optional project, branch, scope, and status filters.
func (m *Manager) List(in ListInput) (ListOutput, error) {
	projFilter, branchFilter, scopeFilter, err := normalizeListFilters(in)
	if err != nil {
		return ListOutput{}, err
	}

	root, err := os.OpenRoot(m.rootDir)
	if err != nil {
		if os.IsNotExist(err) {
			return ListOutput{Snippets: []SnippetSummary{}, Total: 0}, nil
		}
		return ListOutput{}, fmt.Errorf("open snippet root: %w", err)
	}
	defer func() { _ = root.Close() }()

	var summaries []SnippetSummary
	walkErr := fs.WalkDir(root.FS(), ".", func(relPath string, d fs.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if d.IsDir() {
			if d.Name() == ".locks" {
				return fs.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".meta.json") {
			return nil
		}
		meta, readErr := store.ReadJSONRoot[SnippetMeta](m.rootDir, relPath)
		if readErr != nil {
			return nil
		}
		if !matchesListFilter(meta, projFilter, branchFilter, scopeFilter, in.Status) {
			return nil
		}
		summaries = append(summaries, buildSnippetSummary(meta))
		return nil
	})
	if walkErr != nil {
		return ListOutput{}, fmt.Errorf("scan snippet root: %w", walkErr)
	}

	slices.SortFunc(summaries, func(a, b SnippetSummary) int {
		if c := strings.Compare(b.UpdatedAt, a.UpdatedAt); c != 0 {
			return c
		}
		return strings.Compare(a.Name, b.Name)
	})
	if summaries == nil {
		summaries = []SnippetSummary{}
	}
	return ListOutput{Snippets: summaries, Total: len(summaries)}, nil
}

// Verify checks that a snippet version is active, executable, write-protected, and matches its SHA-256 digest.
func (m *Manager) Verify(in VerifyInput) (VerifyOutput, error) {
	name, scope, proj, branch, err := resolveSnippetTarget(in.Name, in.Scope, in.Project, in.Branch, in.WorkDir)
	if err != nil {
		return VerifyOutput{}, err
	}

	var out VerifyOutput
	err = m.locks.WithRLock(m.lockKey(proj, branch, scope, name), func() error {
		meta, exists, err := m.loadMeta(proj, branch, scope, name)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("snippet %q not found under %s/%s/%s", name, proj, branch, scope)
		}
		ver := selectVersion(&meta, in.Version)
		if ver == nil {
			return fmt.Errorf("snippet %q version %d not found", name, in.Version)
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
		Issues:         []string{},
	}
	if ver.Status != StatusActive {
		out.Issues = append(out.Issues, fmt.Sprintf("version %d status is %q (expected active)", ver.Version, ver.Status))
	}

	relPath := filepath.Join(meta.Project, meta.Branch, string(meta.Scope), ver.Filename)
	root, err := os.OpenRoot(m.rootDir)
	if err != nil {
		out.Issues = append(out.Issues, fmt.Sprintf("open snippet root failed: %v", err))
		return out
	}
	defer func() { _ = root.Close() }()

	info, err := root.Stat(relPath)
	if err != nil {
		out.Issues = append(out.Issues, fmt.Sprintf("snippet file missing or unreadable: %v", err))
		return out
	}
	mode := info.Mode().Perm()
	out.Executable = mode&0o100 != 0
	out.Immutable = mode&0o222 == 0
	if !out.Executable {
		out.Issues = append(out.Issues, fmt.Sprintf("snippet file is not executable (mode %04o)", mode))
	}
	if !out.Immutable {
		out.Issues = append(out.Issues, fmt.Sprintf("snippet file is writable (mode %04o; expected write-protected 0500)", mode))
	}

	raw, err := root.ReadFile(relPath)
	if err != nil {
		out.Issues = append(out.Issues, fmt.Sprintf("read snippet file failed: %v", err))
		return out
	}
	out.ActualSHA256 = sha256Hex(raw)
	if out.ActualSHA256 != ver.SHA256 {
		out.Issues = append(out.Issues, "SHA-256 digest mismatch (file modified outside snippet-manager)")
	}
	out.Valid = len(out.Issues) == 0
	return out
}

// RecordInvocation appends an execution record to the snippet metadata ledger.
func (m *Manager) RecordInvocation(in RecordInvocationInput, now time.Time) (RecordInvocationOutput, error) {
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
		meta, exists, err := m.loadMeta(proj, branch, scope, name)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("snippet %q not found under %s/%s/%s", name, proj, branch, scope)
		}
		ver := selectVersion(&meta, in.Version)
		if ver == nil {
			return fmt.Errorf("snippet %q version %d not found", name, in.Version)
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
		var deprecated []int
		for i := range meta.Versions {
			v := &meta.Versions[i]
			if in.Version > 0 && v.Version != in.Version {
				continue
			}
			if v.Status != StatusActive {
				continue
			}
			relPath := filepath.Join(meta.Project, meta.Branch, string(meta.Scope), v.Filename)
			if err := m.chmodWithinRoot(relPath, deprecatedFilePerm); err != nil && !os.IsNotExist(err) {
				return err
			}
			v.Status = StatusDeprecated
			v.DeprecatedAt = ts
			deprecated = append(deprecated, v.Version)
		}
		if in.Version > 0 && len(deprecated) == 0 {
			return fmt.Errorf("active version %d not found for snippet %q", in.Version, name)
		}
		if len(deprecated) > 0 {
			meta.UpdatedAt = ts
			if err := m.saveMeta(meta); err != nil {
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
	projFilter, branchFilter, scopeFilter, nameFilter, err := normalizeCleanupFilters(in)
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
		if readErr != nil {
			continue
		}
		if !matchesCleanupTarget(meta, projFilter, branchFilter, scopeFilter, nameFilter) {
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

func (m *Manager) findMetaFiles() ([]string, error) {
	root, err := os.OpenRoot(m.rootDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("open snippet root: %w", err)
	}
	defer func() { _ = root.Close() }()

	var paths []string
	walkErr := fs.WalkDir(root.FS(), ".", func(relPath string, d fs.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if d.IsDir() {
			if d.Name() == ".locks" {
				return fs.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(d.Name(), ".meta.json") {
			paths = append(paths, relPath)
		}
		return nil
	})
	return paths, walkErr
}

func (m *Manager) deprecateActiveVersions(meta *SnippetMeta, ts string) ([]int, error) {
	var deprecated []int
	for i := range meta.Versions {
		v := &meta.Versions[i]
		if v.Status != StatusActive {
			continue
		}
		relPath := filepath.Join(meta.Project, meta.Branch, string(meta.Scope), v.Filename)
		if err := m.chmodWithinRoot(relPath, deprecatedFilePerm); err != nil && !os.IsNotExist(err) {
			return nil, err
		}
		v.Status = StatusDeprecated
		v.DeprecatedAt = ts
		deprecated = append(deprecated, v.Version)
	}
	return deprecated, nil
}

func (m *Manager) writeImmutableFile(relPath string, data []byte, perm os.FileMode) error {
	parentAbs := filepath.Dir(filepath.Join(m.rootDir, relPath))
	if err := os.MkdirAll(parentAbs, dirPerm); err != nil {
		return fmt.Errorf("create snippet directory: %w", err)
	}
	root, err := os.OpenRoot(m.rootDir)
	if err != nil {
		return fmt.Errorf("open snippet root: %w", err)
	}
	defer func() { _ = root.Close() }()

	f, err := root.OpenFile(relPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, perm)
	if err != nil {
		return fmt.Errorf("create immutable snippet %s: %w", relPath, err)
	}
	cleanup := func() {
		_ = f.Close()
		_ = root.Remove(relPath)
	}
	if _, err := f.Write(data); err != nil {
		cleanup()
		return fmt.Errorf("write snippet %s: %w", relPath, err)
	}
	if err := f.Chmod(perm); err != nil {
		cleanup()
		return fmt.Errorf("chmod snippet %s: %w", relPath, err)
	}
	if err := f.Sync(); err != nil {
		cleanup()
		return fmt.Errorf("sync snippet %s: %w", relPath, err)
	}
	return f.Close()
}

func (m *Manager) chmodWithinRoot(relPath string, perm os.FileMode) error {
	root, err := os.OpenRoot(m.rootDir)
	if err != nil {
		return err
	}
	defer func() { _ = root.Close() }()

	f, err := root.Open(relPath)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	return f.Chmod(perm)
}

func (m *Manager) removeWithinRoot(relPath string) error {
	root, err := os.OpenRoot(m.rootDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer func() { _ = root.Close() }()

	if err := root.Remove(relPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove snippet %s: %w", relPath, err)
	}
	return nil
}

func (m *Manager) loadMeta(proj, branch string, scope Scope, name string) (SnippetMeta, bool, error) {
	relMeta := filepath.Join(proj, branch, string(scope), name+".meta.json")
	meta, err := store.ReadJSONRoot[SnippetMeta](m.rootDir, relMeta)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return SnippetMeta{}, false, nil
		}
		return SnippetMeta{}, false, err
	}
	return meta, true, nil
}

func (m *Manager) saveMeta(meta SnippetMeta) error {
	target := m.metaAbsPath(meta.Project, meta.Branch, meta.Scope, meta.Name)
	return store.WriteJSONAtomic(target, meta, metaFilePerm)
}

func (m *Manager) metaAbsPath(proj, branch string, scope Scope, name string) string {
	return filepath.Join(m.rootDir, proj, branch, string(scope), name+".meta.json")
}

func (m *Manager) lockKey(proj, branch string, scope Scope, name string) string {
	return fmt.Sprintf("%s__%s__%s__%s", proj, branch, scope, name)
}

func resolveSnippetTarget(rawName string, rawScope Scope, project, branch, workDir string) (string, Scope, string, string, error) {
	name, err := ValidateName(rawName)
	if err != nil {
		return "", "", "", "", err
	}
	scope, err := NormalizeScope(rawScope)
	if err != nil {
		return "", "", "", "", err
	}
	proj, br, err := pathutil.ResolveProjectAndBranch(project, branch, workDir)
	if err != nil {
		return "", "", "", "", err
	}
	return name, scope, proj, br, nil
}

func selectVersion(meta *SnippetMeta, version int) *VersionRecord {
	if len(meta.Versions) == 0 {
		return nil
	}
	target := version
	if target <= 0 {
		target = meta.LatestVersion
	}
	for i := range meta.Versions {
		if meta.Versions[i].Version == target {
			return &meta.Versions[i]
		}
	}
	return nil
}

func normalizeListFilters(in ListInput) (string, string, Scope, error) {
	var projFilter, branchFilter string
	var err error
	if strings.TrimSpace(in.Project) != "" {
		if projFilter, err = pathutil.SanitizeSlug(in.Project); err != nil {
			return "", "", "", fmt.Errorf("invalid project filter: %w", err)
		}
	}
	if strings.TrimSpace(in.Branch) != "" {
		if branchFilter, err = pathutil.SanitizeSlug(in.Branch); err != nil {
			return "", "", "", fmt.Errorf("invalid branch filter: %w", err)
		}
	}
	var scopeFilter Scope
	if strings.TrimSpace(string(in.Scope)) != "" {
		if scopeFilter, err = NormalizeScope(in.Scope); err != nil {
			return "", "", "", err
		}
	}
	return projFilter, branchFilter, scopeFilter, nil
}

func normalizeCleanupFilters(in CleanupInput) (string, string, Scope, string, error) {
	var projFilter, branchFilter, nameFilter string
	var scopeFilter Scope
	var err error

	if !in.AllProjects {
		if strings.TrimSpace(in.Project) != "" {
			if projFilter, err = pathutil.SanitizeSlug(in.Project); err != nil {
				return "", "", "", "", fmt.Errorf("invalid project filter: %w", err)
			}
		} else {
			projFilter, _, err = pathutil.ResolveProjectAndBranch("", "placeholder", in.WorkDir)
			if err != nil {
				return "", "", "", "", err
			}
		}
	}
	if strings.TrimSpace(in.Branch) != "" {
		if branchFilter, err = pathutil.SanitizeSlug(in.Branch); err != nil {
			return "", "", "", "", fmt.Errorf("invalid branch filter: %w", err)
		}
	}
	if strings.TrimSpace(string(in.Scope)) != "" {
		if scopeFilter, err = NormalizeScope(in.Scope); err != nil {
			return "", "", "", "", err
		}
	}
	if strings.TrimSpace(in.Name) != "" {
		if nameFilter, err = ValidateName(in.Name); err != nil {
			return "", "", "", "", err
		}
	}
	return projFilter, branchFilter, scopeFilter, nameFilter, nil
}

func matchesListFilter(meta SnippetMeta, proj, branch string, scope Scope, status VersionStatus) bool {
	if proj != "" && meta.Project != proj {
		return false
	}
	if branch != "" && meta.Branch != branch {
		return false
	}
	if scope != "" && meta.Scope != scope {
		return false
	}
	if status != "" {
		latest := selectVersion(&meta, 0)
		if latest == nil || latest.Status != status {
			return false
		}
	}
	return true
}

func matchesCleanupTarget(meta SnippetMeta, proj, branch string, scope Scope, name string) bool {
	if proj != "" && meta.Project != proj {
		return false
	}
	if branch != "" && meta.Branch != branch {
		return false
	}
	if scope != "" && meta.Scope != scope {
		return false
	}
	if name != "" && meta.Name != name {
		return false
	}
	return true
}

func buildSnippetSummary(meta SnippetMeta) SnippetSummary {
	s := SnippetSummary{
		Name:            meta.Name,
		Project:         meta.Project,
		Branch:          meta.Branch,
		Scope:           meta.Scope,
		Language:        meta.Language,
		LatestVersion:   meta.LatestVersion,
		InvocationCount: len(meta.Invocations),
		UpdatedAt:       meta.UpdatedAt,
	}
	if latest := selectVersion(&meta, 0); latest != nil {
		s.LatestStatus = latest.Status
		s.LatestPath = latest.Path
		s.LatestDualPath = latest.DualPath
		s.LatestAuditLvl = latest.AuditLevel
	}
	return s
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
