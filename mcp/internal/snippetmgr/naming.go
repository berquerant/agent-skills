package snippetmgr

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/berquerant/agent-skills/mcp/internal/guard"
)

const (
	maxSnippetNameLen   = 80
	maxSnippetCodeBytes = 1024 * 1024
)

var (
	validSnippetNameRe = regexp.MustCompile(`^[a-z][a-z0-9_-]*$`)
	versionSuffixRe    = regexp.MustCompile(`_v[0-9]+$`)
)

// genericSnippetNames lists vague single-word names disallowed by command-snippet naming rules.
var genericSnippetNames = map[string]struct{}{
	"run": {}, "script": {}, "helper": {}, "test": {}, "tmp": {}, "temp": {},
	"cmd": {}, "exec": {}, "tool": {}, "util": {}, "utils": {}, "main": {},
	"do": {}, "check": {}, "fix": {}, "resolve": {}, "verify": {}, "update": {},
	"snippet": {}, "foo": {}, "bar": {}, "work": {}, "task": {}, "job": {},
}

// ValidateName enforces self-descriptive, lowercase, extension-free, version-free snippet names.
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
	if _, generic := genericSnippetNames[name]; generic {
		return "", fmt.Errorf("snippet name %q is too generic; use a self-descriptive action+target name (e.g. rebase_and_test_worktree)", name)
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

// NormalizeLanguage validates language and returns canonical language name and file extension.
func NormalizeLanguage(raw string) (string, string, error) {
	lang := strings.ToLower(strings.TrimSpace(raw))
	if lang == "" {
		lang = "bash"
	}
	switch lang {
	case "bash", "sh", "zsh", "mksh":
		return lang, ".sh", nil
	case "python", "python3":
		return "python", ".py", nil
	default:
		return "", "", fmt.Errorf("unsupported language %q (supported: bash, sh, zsh, mksh, python)", raw)
	}
}

// ValidateContent checks secret hygiene, shebang, documentation header, and shell safety flags.
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
	// Inspect the leading header block (first 40 lines) for Description/Purpose and Usage.
	lines := strings.Split(code, "\n")
	limit := min(len(lines), 40)
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
