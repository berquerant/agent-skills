// Package guard enforces security boundaries across all MCP server features by rejecting
// access to sensitive files, system paths, private keys, and high-entropy secret patterns.
//
// Rules and patterns are derived from industry-standard tools (Gitleaks, TruffleHog, and
// git-secrets) to protect user credentials and host operating system configurations.
package guard

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// SensitiveFilenames lists exact filenames blocked from reading, caching, or writing.
// Derived from Gitleaks default rules (gitleaks.toml), TruffleHog, and git-secrets path rules.
var SensitiveFilenames = map[string]struct{}{
	".env":                     {},
	".env.local":               {},
	".env.development":         {},
	".env.staging":             {},
	".env.production":          {},
	".npmrc":                   {},
	".pypirc":                  {},
	".netrc":                   {},
	".htpasswd":                {},
	".git-credentials":         {},
	"id_rsa":                   {},
	"id_ed25519":               {},
	"id_ecdsa":                 {},
	"id_dsa":                   {},
	"credentials.json":         {},
	"service-account.json":     {},
	"terraform.tfstate":        {},
	"terraform.tfstate.backup": {},
	"terraform.tfvars":         {},
	"kubeconfig":               {},
}

// SensitiveExtensions lists file extensions associated with private keys, certificates, and keystores.
var SensitiveExtensions = map[string]struct{}{
	".pem":      {},
	".key":      {},
	".p12":      {},
	".pfx":      {},
	".jks":      {},
	".keystore": {},
	".asc":      {},
	".ovpn":     {},
	".kdbx":     {},
}

type pathRule struct {
	name    string
	pattern *regexp.Regexp
}

var sensitivePathPatterns = []pathRule{
	{name: "AWS config/credentials", pattern: regexp.MustCompile(`(?i)(^|/)(\.aws/credentials|\.aws/config)$`)},
	{name: "SSH directory", pattern: regexp.MustCompile(`(?i)(^|/)(\.ssh/[^/]+.*)$`)},
	{name: "GnuPG directory", pattern: regexp.MustCompile(`(?i)(^|/)(\.gnupg/[^/]+.*)$`)},
	{name: "Docker config", pattern: regexp.MustCompile(`(?i)(^|/)(\.docker/config\.json)$`)},
	{name: "Kubeconfig", pattern: regexp.MustCompile(`(?i)(^|/)(\.kube/config)$`)},
	{name: "Google Cloud config", pattern: regexp.MustCompile(`(?i)(^|/)(\.config/gcloud/[^/]+.*)$`)},
	{name: "System auth file", pattern: regexp.MustCompile(`(?i)(^|/)etc/(shadow|passwd|sudoers)$`)},
}

// IsSensitivePath checks whether filePath matches a sensitive filename, extension, or directory pattern.
// It returns true and the matched rule description if the path is sensitive.
func IsSensitivePath(filePath string) (bool, string) {
	trimmed := strings.TrimSpace(filePath)
	if trimmed == "" {
		return false, ""
	}
	if strings.HasPrefix(trimmed, "-") {
		if _, val, ok := strings.Cut(trimmed, "="); ok && val != "" {
			trimmed = val
		}
	}
	normalized := strings.ReplaceAll(filepath.Clean(trimmed), `\`, "/")
	baseLower := strings.ToLower(filepath.Base(normalized))

	if _, ok := SensitiveFilenames[baseLower]; ok {
		return true, fmt.Sprintf("sensitive filename %q", baseLower)
	}
	if strings.HasPrefix(baseLower, ".env.") {
		return true, fmt.Sprintf("dotenv file %q", baseLower)
	}
	ext := filepath.Ext(baseLower)
	if _, ok := SensitiveExtensions[ext]; ok {
		return true, fmt.Sprintf("sensitive extension %q", ext)
	}
	for _, rule := range sensitivePathPatterns {
		if rule.pattern.MatchString(normalized) {
			return true, rule.name
		}
	}
	return false, ""
}

// CheckPathAllowed returns an error if filePath matches any sensitive path rule.
func CheckPathAllowed(filePath string) error {
	if sensitive, reason := IsSensitivePath(filePath); sensitive {
		return fmt.Errorf("access to sensitive path %q is denied (%s)", filePath, reason)
	}
	return nil
}
