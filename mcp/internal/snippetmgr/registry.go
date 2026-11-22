package snippetmgr

import (
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
