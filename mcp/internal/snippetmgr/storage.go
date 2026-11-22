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
	"syscall"

	"github.com/berquerant/agent-skills/mcp/internal/pathutil"
	"github.com/berquerant/agent-skills/mcp/internal/store"
)

func (m *Manager) withRoot(fn func(root *os.Root) error) error {
	root, err := os.OpenRoot(m.rootDir)
	if err != nil {
		return err
	}
	defer func() { _ = root.Close() }()
	return fn(root)
}

func (m *Manager) findMetaFiles() ([]string, error) {
	var paths []string
	err := m.withRoot(func(root *os.Root) error {
		return fs.WalkDir(root.FS(), ".", func(relPath string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				if os.IsNotExist(walkErr) {
					return nil
				}
				return walkErr
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
	})
	if os.IsNotExist(err) {
		return nil, nil
	}
	return paths, err
}

// markVersionsDeprecated marks matching active versions (targetVersion == 0 means all active)
// as deprecated in meta and returns the newly deprecated version numbers.
func markVersionsDeprecated(meta *SnippetMeta, targetVersion int, ts string) []int {
	var deprecated []int
	for i := range meta.Versions {
		v := &meta.Versions[i]
		if targetVersion > 0 && v.Version != targetVersion {
			continue
		}
		if v.Status != StatusActive {
			continue
		}
		v.Status = StatusDeprecated
		v.DeprecatedAt = ts
		deprecated = append(deprecated, v.Version)
	}
	return deprecated
}

// chmodDeprecatedFiles strips execute permissions (mode 0400) from the specified versions on disk.
func (m *Manager) chmodDeprecatedFiles(meta SnippetMeta, versions []int) error {
	for _, verNum := range versions {
		v := selectVersion(&meta, verNum)
		if v == nil {
			continue
		}
		relPath := filepath.Join(meta.Project, meta.Branch, string(meta.Scope), v.Filename)
		if err := m.chmodWithinRoot(relPath, deprecatedFilePerm); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

func (m *Manager) writeImmutableFile(relPath string, data []byte, perm os.FileMode) (uint64, error) {
	parentAbs := filepath.Dir(filepath.Join(m.rootDir, relPath))
	if err := os.MkdirAll(parentAbs, dirPerm); err != nil {
		return 0, fmt.Errorf("create snippet directory: %w", err)
	}

	var inode uint64
	err := m.withRoot(func(root *os.Root) error {
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
		if info, err := f.Stat(); err == nil {
			if ino, ok := fileInode(info); ok {
				inode = ino
			}
		}
		if err := f.Sync(); err != nil {
			cleanup()
			return fmt.Errorf("sync snippet %s: %w", relPath, err)
		}
		return f.Close()
	})
	return inode, err
}

func (m *Manager) chmodWithinRoot(relPath string, perm os.FileMode) error {
	return m.withRoot(func(root *os.Root) error {
		f, err := root.Open(relPath)
		if err != nil {
			return err
		}
		defer func() { _ = f.Close() }()
		return f.Chmod(perm)
	})
}

func (m *Manager) removeWithinRoot(relPath string) error {
	err := m.withRoot(func(root *os.Root) error {
		if rmErr := root.Remove(relPath); rmErr != nil && !os.IsNotExist(rmErr) {
			return fmt.Errorf("remove snippet %s: %w", relPath, rmErr)
		}
		return nil
	})
	if os.IsNotExist(err) {
		return nil
	}
	return err
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

func (m *Manager) loadTargetVersion(proj, branch string, scope Scope, name string, version int) (SnippetMeta, *VersionRecord, error) {
	meta, exists, err := m.loadMeta(proj, branch, scope, name)
	if err != nil {
		return SnippetMeta{}, nil, err
	}
	if !exists {
		return SnippetMeta{}, nil, fmt.Errorf("snippet %q not found under %s/%s/%s", name, proj, branch, scope)
	}
	ver := selectVersion(&meta, version)
	if ver == nil {
		return SnippetMeta{}, nil, fmt.Errorf("snippet %q version %d not found", name, version)
	}
	return meta, ver, nil
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

type snippetFilter struct {
	project string
	branch  string
	scope   Scope
	name    string
	status  VersionStatus
}

func newSnippetFilter(project, branch string, scope Scope, name string, status VersionStatus) (snippetFilter, error) {
	var f snippetFilter
	var err error
	if strings.TrimSpace(project) != "" {
		if f.project, err = pathutil.SanitizeSlug(project); err != nil {
			return snippetFilter{}, fmt.Errorf("invalid project filter: %w", err)
		}
	}
	if strings.TrimSpace(branch) != "" {
		if f.branch, err = pathutil.SanitizeSlug(branch); err != nil {
			return snippetFilter{}, fmt.Errorf("invalid branch filter: %w", err)
		}
	}
	if strings.TrimSpace(string(scope)) != "" {
		if f.scope, err = NormalizeScope(scope); err != nil {
			return snippetFilter{}, err
		}
	}
	if strings.TrimSpace(name) != "" {
		if f.name, err = ValidateName(name); err != nil {
			return snippetFilter{}, err
		}
	}
	if f.status, err = NormalizeStatus(status); err != nil {
		return snippetFilter{}, err
	}
	return f, nil
}

func (f snippetFilter) matches(meta SnippetMeta) bool {
	if f.project != "" && meta.Project != f.project {
		return false
	}
	if f.branch != "" && meta.Branch != f.branch {
		return false
	}
	if f.scope != "" && meta.Scope != f.scope {
		return false
	}
	if f.name != "" && meta.Name != f.name {
		return false
	}
	if f.status != "" {
		latest := selectVersion(&meta, 0)
		if latest == nil || latest.Status != f.status {
			return false
		}
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

func fileInode(info fs.FileInfo) (uint64, bool) {
	if info == nil {
		return 0, false
	}
	st, ok := info.Sys().(*syscall.Stat_t)
	if !ok || st == nil {
		return 0, false
	}
	return st.Ino, true
}
