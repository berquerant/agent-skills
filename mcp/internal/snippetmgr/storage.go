package snippetmgr

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/berquerant/agent-skills/mcp/internal/pathutil"
	"github.com/berquerant/agent-skills/mcp/internal/store"
)

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
