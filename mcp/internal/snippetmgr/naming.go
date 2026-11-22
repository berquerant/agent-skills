package snippetmgr

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/berquerant/agent-skills/mcp/internal/guard"
	"github.com/berquerant/agent-skills/mcp/internal/snippet"
)

const (
	maxSnippetNameLen   = 80
	maxSnippetCodeBytes = 1024 * 1024
	headerScanMaxLines  = 40
)

var (
	validSnippetNameRe = regexp.MustCompile(`^[a-z][a-z0-9_-]*$`)
	versionSuffixRe    = regexp.MustCompile(`_v[0-9]+$`)
)

// ValidateName enforces structural and filesystem-safe snippet names
// (lowercase alphanumeric with _ or -, no path separators, no extensions, no manual _vN suffix).
// Semantic naming quality (self-descriptive action+target names) is governed by the command-snippet skill.
func ValidateName(raw string) (string, error) {
	name := strings.TrimSpace(raw)
	if name == "" {
		return "", fmt.Errorf("snippet name is required")
	}
	if len(name) > maxSnippetNameLen {
		return "", fmt.Errorf("snippet name exceeds %d characters: %q", maxSnippetNameLen, name)
	}
	if strings.ContainsAny(name, `/\.`) {
		return "", fmt.Errorf("snippet name must not contain path separators or file extensions: %q", name)
	}
	if !validSnippetNameRe.MatchString(name) {
		return "", fmt.Errorf("snippet name must start with a lowercase letter and use only [a-z0-9_-]: %q", name)
	}
	if versionSuffixRe.MatchString(name) {
		return "", fmt.Errorf("snippet name must not end with a manual _vN version suffix (versioning is managed automatically): %q", name)
	}
	return name, nil
}

// NormalizeScope validates and normalizes snippet scope, defaulting to ScopeOneOff.
func NormalizeScope(raw Scope) (Scope, error) {
	s := strings.ToLower(strings.TrimSpace(string(raw)))
	if s == "" {
		return ScopeOneOff, nil
	}
	switch Scope(s) {
	case ScopeOneOff, ScopeReusable:
		return Scope(s), nil
	default:
		return "", fmt.Errorf("unsupported scope %q (expected 'one-off' or 'reusable')", raw)
	}
}

// NormalizeStatus validates and normalizes a non-empty version status filter.
func NormalizeStatus(raw VersionStatus) (VersionStatus, error) {
	s := strings.ToLower(strings.TrimSpace(string(raw)))
	if s == "" {
		return "", nil
	}
	switch VersionStatus(s) {
	case StatusActive, StatusDeprecated, StatusCleaned:
		return VersionStatus(s), nil
	default:
		return "", fmt.Errorf("unsupported status %q (expected 'active', 'deprecated', or 'cleaned')", raw)
	}
}

// NormalizeLanguage validates language via snippet.NormalizeLanguage and returns the canonical language and extension.
func NormalizeLanguage(raw string) (string, string, error) {
	lang, err := snippet.NormalizeLanguage(raw)
	if err != nil {
		return "", "", err
	}
	if lang == "python" {
		return lang, ".py", nil
	}
	return lang, ".sh", nil
}

// ValidateContent checks secret hygiene, LF line endings, shebang, documentation header, and shell safety flags.
func ValidateContent(code, lang string) error {
	if strings.TrimSpace(code) == "" {
		return fmt.Errorf("snippet code must not be empty")
	}
	if len(code) > maxSnippetCodeBytes {
		return fmt.Errorf("snippet code exceeds %d bytes", maxSnippetCodeBytes)
	}
	if err := guard.CheckNoSecrets(code); err != nil {
		return fmt.Errorf("snippet code failed secret check: %w", err)
	}
	if strings.Contains(code, "\r") {
		return fmt.Errorf("snippet code must use LF (Unix) line endings, not CRLF")
	}

	lines := strings.Split(code, "\n")
	if !strings.HasPrefix(lines[0], "#!") {
		return fmt.Errorf("snippet must start with a shebang (e.g. #!/usr/bin/env bash or #!/usr/bin/env python3)")
	}

	if err := checkDocumentationHeader(code); err != nil {
		return err
	}
	return checkShellHardening(code, lang)
}

func checkDocumentationHeader(code string) error {
	lines := strings.Split(code, "\n")
	limit := min(len(lines), headerScanMaxLines)
	header := strings.ToLower(strings.Join(lines[:limit], "\n"))

	hasDesc := strings.Contains(header, "description:") || strings.Contains(header, "purpose:")
	hasUsage := strings.Contains(header, "usage:")
	if !hasDesc || !hasUsage {
		return fmt.Errorf("snippet must include a documentation header with Description (or Purpose) and Usage sections")
	}
	return nil
}

func checkShellHardening(code, lang string) error {
	if lang == "python" {
		return nil
	}
	target := "set -euo pipefail"
	if lang == "sh" {
		target = "set -eu"
	}
	for line := range strings.SplitSeq(code, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if strings.Contains(trimmed, target) {
			return nil
		}
	}
	if lang == "sh" {
		return fmt.Errorf("POSIX sh snippet must include 'set -eu' (or 'set -euo pipefail') for fail-fast hardening")
	}
	return fmt.Errorf("%s snippet must include 'set -euo pipefail' for fail-fast hardening", lang)
}
