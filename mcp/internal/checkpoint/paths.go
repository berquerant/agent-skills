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
)

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

// ResolvedPath holds both the root-relative path and the clean absolute path of a checkpoint log file.
type ResolvedPath struct {
	Rel string
	Abs string
}

// ResolveRelativeWithinRoot resolves targetPath (absolute or root-relative) to a clean relative path
// inside rootDir, rejecting any path that escapes rootDir or matches sensitive path rules.
func ResolveRelativeWithinRoot(rootDir, targetPath string) (ResolvedPath, error) {
	if strings.TrimSpace(targetPath) == "" {
		return ResolvedPath{}, fmt.Errorf("path is required")
	}
	if err := guard.CheckPathAllowed(targetPath); err != nil {
		return ResolvedPath{}, err
	}
	absRoot, err := filepath.Abs(rootDir)
	if err != nil {
		return ResolvedPath{}, fmt.Errorf("resolve checkpoint root: %w", err)
	}
	var candidate string
	if filepath.IsAbs(targetPath) {
		candidate = filepath.Clean(targetPath)
	} else {
		candidate = filepath.Clean(filepath.Join(absRoot, targetPath))
	}

	rel, err := filepath.Rel(absRoot, candidate)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return ResolvedPath{}, fmt.Errorf("path %q is outside checkpoint root %q", targetPath, absRoot)
	}
	if !strings.HasSuffix(rel, ".md") {
		return ResolvedPath{}, fmt.Errorf("checkpoint file path must end with .md: %q", targetPath)
	}
	return ResolvedPath{Rel: rel, Abs: candidate}, nil
}

// GenerationQuery specifies the target context directory and document type for generation discovery.
type GenerationQuery struct {
	RootDir string
	Project string
	Branch  string
	Context string
	DocType DocType
}

// FindLatestGeneration returns the absolute path of the newest <docType>-*.md file in the context directory,
// or "" if none exists.
func FindLatestGeneration(q GenerationQuery) (string, error) {
	relDir := filepath.Join(q.Project, q.Branch, q.Context)
	root, err := os.OpenRoot(q.RootDir)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	defer func() { _ = root.Close() }()

	dirPath := filepath.Join(q.RootDir, relDir)
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}

	prefix := string(q.DocType) + "-"
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
