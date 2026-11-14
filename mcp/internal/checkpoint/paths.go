package checkpoint

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/berquerant/agent-skills/mcp/internal/guard"
)

var unsafeSlugChars = regexp.MustCompile(`[^a-zA-Z0-9_-]+`)

// SanitizeSlug replaces characters outside [a-zA-Z0-9_-] with '-' and trims leading/trailing hyphens.
func SanitizeSlug(raw string) (string, error) {
	cleaned := unsafeSlugChars.ReplaceAllString(strings.TrimSpace(raw), "-")
	cleaned = strings.Trim(cleaned, "-_")
	if cleaned == "" {
		return "", fmt.Errorf("invalid empty identifier after sanitization from %q", raw)
	}
	if len(cleaned) > 128 {
		cleaned = cleaned[:128]
	}
	return cleaned, nil
}

// NormalizeDocType validates and normalizes a DocType, defaulting to DocTypeCheckpoint.
func NormalizeDocType(dt DocType) (DocType, error) {
	s := strings.ToLower(strings.TrimSpace(string(dt)))
	if s == "" {
		return DocTypeCheckpoint, nil
	}
	switch DocType(s) {
	case DocTypeCheckpoint, DocTypeRunbook, DocTypeReport:
		return DocType(s), nil
	default:
		return "", fmt.Errorf("unsupported doc_type %q (expected checkpoint, runbook, or report)", dt)
	}
}

// FormatDualPath formats an absolute file path according to the output-presentation convention:
// [<basename>](file://<abs_path>) (<abs_path>).
func FormatDualPath(absPath string) string {
	slashPath := filepath.ToSlash(absPath)
	base := filepath.Base(slashPath)
	return fmt.Sprintf("[%s](file://%s) (%s)", base, slashPath, slashPath)
}

// ResolveProjectAndBranch resolves project and branch from explicit inputs or git in workDir.
func ResolveProjectAndBranch(project, branch, workDir string) (string, string, error) {
	dir := strings.TrimSpace(workDir)
	if dir != "" {
		if err := guard.CheckPathAllowed(dir); err != nil {
			return "", "", err
		}
	} else {
		if wd, err := os.Getwd(); err == nil {
			dir = wd
		}
	}

	rawProj := strings.TrimSpace(project)
	if rawProj == "" {
		top, err := runGit(dir, "rev-parse", "--show-toplevel")
		if err != nil {
			return "", "", fmt.Errorf("auto-derive project via git failed (provide project explicitly): %w", err)
		}
		rawProj = filepath.Base(top)
	}

	rawBranch := strings.TrimSpace(branch)
	if rawBranch == "" {
		br, err := runGit(dir, "branch", "--show-current")
		if err == nil && br != "" {
			rawBranch = br
		} else {
			sha, shaErr := runGit(dir, "rev-parse", "--short", "HEAD")
			if shaErr != nil {
				return "", "", fmt.Errorf("auto-derive branch via git failed (provide branch explicitly): %w", shaErr)
			}
			rawBranch = sha
		}
	}

	projSlug, err := SanitizeSlug(rawProj)
	if err != nil {
		return "", "", fmt.Errorf("invalid project: %w", err)
	}
	branchSlug, err := SanitizeSlug(rawBranch)
	if err != nil {
		return "", "", fmt.Errorf("invalid branch: %w", err)
	}
	return projSlug, branchSlug, nil
}

func runGit(dir string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// ResolveRelativeWithinRoot resolves targetPath (absolute or root-relative) to a clean relative path
// inside rootDir, rejecting any path that escapes rootDir or matches sensitive path rules.
func ResolveRelativeWithinRoot(rootDir, targetPath string) (string, string, error) {
	if strings.TrimSpace(targetPath) == "" {
		return "", "", fmt.Errorf("path is required")
	}
	if err := guard.CheckPathAllowed(targetPath); err != nil {
		return "", "", err
	}
	absRoot, err := filepath.Abs(rootDir)
	if err != nil {
		return "", "", fmt.Errorf("resolve checkpoint root: %w", err)
	}
	var candidate string
	if filepath.IsAbs(targetPath) {
		candidate = filepath.Clean(targetPath)
	} else {
		candidate = filepath.Clean(filepath.Join(absRoot, targetPath))
	}

	rel, err := filepath.Rel(absRoot, candidate)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", "", fmt.Errorf("path %q is outside checkpoint root %q", targetPath, absRoot)
	}
	if !strings.HasSuffix(rel, ".md") {
		return "", "", fmt.Errorf("checkpoint file path must end with .md: %q", targetPath)
	}
	return rel, candidate, nil
}

// FindLatestGeneration returns the absolute path of the newest <docType>-*.md file in the context directory,
// or "" if none exists.
func FindLatestGeneration(rootDir, projSlug, branchSlug, ctxSlug string, docType DocType) (string, error) {
	relDir := filepath.Join(projSlug, branchSlug, ctxSlug)
	root, err := os.OpenRoot(rootDir)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	defer func() { _ = root.Close() }()

	dirPath := filepath.Join(rootDir, relDir)
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}

	prefix := string(docType) + "-"
	var matches []string
	for _, e := range entries {
		name := e.Name()
		if !e.IsDir() && strings.HasPrefix(name, prefix) && strings.HasSuffix(name, ".md") {
			if _, statErr := root.Stat(filepath.Join(relDir, name)); statErr == nil {
				matches = append(matches, name)
			}
		}
	}
	if len(matches) == 0 {
		return "", nil
	}
	slices.Sort(matches)
	latest := matches[len(matches)-1]
	return filepath.Join(dirPath, latest), nil
}
