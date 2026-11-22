package pathutil

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/berquerant/agent-skills/mcp/internal/guard"
)

var unsafeSlugChars = regexp.MustCompile(`[^a-zA-Z0-9_-]+`)

// SanitizeSlug replaces characters outside [a-zA-Z0-9_-] with '-' and trims leading/trailing hyphens and underscores.
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

// ResolveProjectAndBranch resolves and sanitizes project and branch slugs from explicit inputs
// or by inspecting git metadata in workDir (defaulting to the current working directory).
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
