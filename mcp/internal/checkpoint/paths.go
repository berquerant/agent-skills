// Package checkpoint implements structured, append-only milestone logging, task lifecycle
// tracking, and context recovery for the checkpoint-log skill.
//
// File & Path Conventions:
//   - Logs follow the hierarchical structure: `<root>/<project>/<branch>/<context>/<doc_type>-<timestamp>.md`
//   - Each file contains a Markdown header block (`# Task Checkpoint Log: ...` with metadata bullets)
//     followed by zero or more chronological, append-only milestone entries.
//   - Writes are guarded by cross-process file locks to prevent race conditions during parallel execution.
package checkpoint

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/berquerant/agent-skills/mcp/internal/guard"
	"github.com/berquerant/agent-skills/mcp/internal/pathutil"
)

// SanitizeSlug replaces characters outside [a-zA-Z0-9_-] with '-' and trims leading/trailing hyphens.
func SanitizeSlug(raw string) (string, error) {
	return pathutil.SanitizeSlug(raw)
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
	return pathutil.FormatDualPath(absPath)
}

// ResolveProjectAndBranch resolves project and branch from explicit inputs or git in workDir.
func ResolveProjectAndBranch(project, branch, workDir string) (string, string, error) {
	return pathutil.ResolveProjectAndBranch(project, branch, workDir)
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
