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
	"github.com/berquerant/agent-skills/mcp/internal/snippet"
	"github.com/berquerant/agent-skills/mcp/internal/store"
)

const (
	activeFilePerm     os.FileMode = 0o500
	deprecatedFilePerm os.FileMode = 0o400
	metaFilePerm       os.FileMode = 0o600
	dirPerm            os.FileMode = 0o700
	ownerExecBit       os.FileMode = 0o100
	anyWriteBits       os.FileMode = 0o222
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
	target, err := resolveSnippetTarget(targetRefInput{
		Name:    in.Name,
		Scope:   in.Scope,
		Project: in.Project,
		Branch:  in.Branch,
		WorkDir: in.WorkDir,
	})
	if err != nil {
		return CreateOutput{}, err
	}
	langSpec, err := NormalizeLanguage(in.Language)
	if err != nil {
		return CreateOutput{}, err
	}
	if err := ValidateContent(in.Code, langSpec.Name); err != nil {
		return CreateOutput{}, err
	}
	if in.Purpose != "" {
		if err := guard.CheckNoSecrets(in.Purpose); err != nil {
			return CreateOutput{}, fmt.Errorf("purpose failed secret check: %w", err)
		}
	}

	audit := snippet.Analyze(in.Code, langSpec.Name)
	if !audit.OK {
		return CreateOutput{}, fmt.Errorf("snippet audit failed: %s", audit.ParseError)
	}
	if audit.Level >= snippet.LevelComplex && !in.AllowComplex {
		return CreateOutput{
			Saved:          false,
			ActionRequired: "simplify_or_consult",
			Name:           target.Name,
			Project:        target.Project,
			Branch:         target.Branch,
			Scope:          target.Scope,
			Language:       langSpec.Name,
			Audit:          audit,
		}, nil
	}

	return m.saveVersionedSnippet(createParams{
		target:       target,
		langSpec:     langSpec,
		code:         in.Code,
		purpose:      strings.TrimSpace(in.Purpose),
		allowComplex: in.AllowComplex,
		audit:        audit,
		now:          now,
	})
}

type createParams struct {
	target       snippetTarget
	langSpec     LanguageSpec
	code         string
	purpose      string
	allowComplex bool
	audit        snippet.AuditReport
	now          time.Time
}

func (m *Manager) saveVersionedSnippet(p createParams) (CreateOutput, error) {
	var out CreateOutput

	err := m.locks.WithLock(p.target.lockKey(), func() error {
		loaded, err := m.loadMeta(p.target)
		if err != nil {
			return err
		}
		ts := p.now.UTC().Format(time.RFC3339)
		var meta SnippetMeta
		if loaded == nil {
			meta = SnippetMeta{
				Name:      p.target.Name,
				Project:   p.target.Project,
				Branch:    p.target.Branch,
				Scope:     p.target.Scope,
				Language:  p.langSpec.Name,
				Extension: p.langSpec.Extension,
				Purpose:   p.purpose,
				CreatedAt: ts,
			}
		} else {
			meta = *loaded
			if p.purpose != "" {
				meta.Purpose = p.purpose
			}
			meta.Language = p.langSpec.Name
			meta.Extension = p.langSpec.Extension
		}

		nextVer := meta.LatestVersion + 1
		filename := fmt.Sprintf("%s_v%d%s", p.target.Name, nextVer, p.langSpec.Extension)
		relPath := filepath.Join(p.target.Project, p.target.Branch, string(p.target.Scope), filename)
		absPath := filepath.Join(m.rootDir, relPath)

		codeBytes := []byte(p.code)
		inode, err := m.writeImmutableFile(relPath, codeBytes)
		if err != nil {
			return err
		}

		deprecatedVers := meta.markVersionsDeprecated(0, ts)

		digest := sha256Hex(codeBytes)
		dualPath := pathutil.FormatDualPath(absPath)
		rec := VersionRecord{
			Version:      nextVer,
			Filename:     filename,
			Path:         absPath,
			DualPath:     dualPath,
			SHA256:       digest,
			Inode:        inode,
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
			_ = m.removeWithinRoot(relPath)
			return err
		}
		if err := m.chmodDeprecatedFiles(meta, deprecatedVers); err != nil {
			return err
		}

		out = CreateOutput{
			Saved:              true,
			Name:               p.target.Name,
			Project:            p.target.Project,
			Branch:             p.target.Branch,
			Scope:              p.target.Scope,
			Language:           p.langSpec.Name,
			Version:            nextVer,
			Path:               absPath,
			DualPath:           dualPath,
			SHA256:             digest,
			Inode:              inode,
			DeprecatedVersions: deprecatedVers,
			Audit:              p.audit,
			PreExecutionNotice: BuildPreExecutionNotice(NoticeParams{
				DualPath:    dualPath,
				AbsPath:     absPath,
				Purpose:     meta.Purpose,
				AuditLevel:  p.audit.Level,
				AuditRating: p.audit.Rating,
			}),
			DirectInvocationHint: absPath,
		}
		return nil
	})
	return out, err
}

// Get retrieves metadata and the requested (or latest) version record for a snippet.
func (m *Manager) Get(in GetInput) (GetOutput, error) {
	target, err := resolveSnippetTarget(targetRefInput{
		Name:    in.Name,
		Scope:   in.Scope,
		Project: in.Project,
		Branch:  in.Branch,
		WorkDir: in.WorkDir,
	})
	if err != nil {
		return GetOutput{}, err
	}

	var out GetOutput
	err = m.locks.WithRLock(target.lockKey(), func() error {
		res, err := m.loadTargetVersion(target, in.Version)
		if err != nil {
			return err
		}
		verCopy := *res.Version
		out = GetOutput{
			Meta:            res.Meta,
			SelectedVersion: &verCopy,
		}
		return nil
	})
	return out, err
}

// List scans managed snippets matching optional project, branch, scope, and status filters.
func (m *Manager) List(in ListInput) (ListOutput, error) {
	flt, err := newSnippetFilter(snippetFilterInput{
		Project: in.Project,
		Branch:  in.Branch,
		Scope:   in.Scope,
		Status:  in.Status,
	})
	if err != nil {
		return ListOutput{}, err
	}

	metaRelPaths, err := m.findMetaFiles()
	if err != nil {
		return ListOutput{}, err
	}

	var summaries []SnippetSummary
	for _, relPath := range metaRelPaths {
		meta, readErr := store.ReadJSONRoot[SnippetMeta](m.rootDir, relPath)
		if readErr != nil || !flt.matches(meta) {
			continue
		}
		summaries = append(summaries, buildSnippetSummary(meta))
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
