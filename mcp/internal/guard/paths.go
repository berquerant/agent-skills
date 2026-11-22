// Package guard enforces security boundaries across all MCP server features by rejecting
// access to sensitive files, system paths, private keys, and high-entropy secret patterns.
//
// All rule definitions are maintained declaratively in embedded rules.json.
// See mcp/references/security-rules.md for inclusion criteria and maintenance procedures.
package guard

import (
	"fmt"
	"path/filepath"
	"strings"
)

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
