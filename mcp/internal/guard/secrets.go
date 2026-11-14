package guard

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
)

type secretRule struct {
	name    string
	pattern *regexp.Regexp
}

// secretValuePatterns defines high-confidence secret value patterns based on Gitleaks (gitleaks.toml),
// TruffleHog, and git-secrets rules.
var secretValuePatterns = []secretRule{
	{
		name:    "AWS Access Key",
		pattern: regexp.MustCompile(`\b(A3T[A-Z0-9]|AKIA|AGPA|AIDA|AROA|AIPA|ANPA|ANVA|ASIA)[A-Z0-9]{16}\b`),
	},
	{
		name:    "GitHub Token",
		pattern: regexp.MustCompile(`\b(gh[pousr]_[A-Za-z0-9_]{20,255}|github_pat_[A-Za-z0-9_]{20,255})\b`),
	},
	{
		name:    "GitLab Token",
		pattern: regexp.MustCompile(`\bglpat-[A-Za-z0-9_\-=]{20,}\b`),
	},
	{
		name:    "Slack Token",
		pattern: regexp.MustCompile(`\bxox[baprs]-[0-9]{10,13}-[0-9]{10,13}-[a-zA-Z0-9]{24,32}\b`),
	},
	{
		name:    "OpenAI API Key",
		pattern: regexp.MustCompile(`\bsk-(proj|svcacct)-[A-Za-z0-9_-]{20,}\b`),
	},
	{
		name:    "Anthropic API Key",
		pattern: regexp.MustCompile(`\bsk-ant-api03-[A-Za-z0-9_-]{20,}\b`),
	},
	{
		name:    "GCP API Key",
		pattern: regexp.MustCompile(`\bAIza[0-9A-Za-z\-_]{35}\b`),
	},
	{
		name:    "Stripe Key",
		pattern: regexp.MustCompile(`\b(sk|rk)_live_[0-9a-zA-Z]{24,99}\b`),
	},
	{
		name:    "Private Key Block",
		pattern: regexp.MustCompile(`-----BEGIN (RSA |EC |DSA |OPENSSH |PGP |ENCRYPTED )?PRIVATE KEY-----`),
	},
	{
		name:    "JSON Web Token",
		pattern: regexp.MustCompile(`\beyJ[A-Za-z0-9_-]{10,}\.eyJ[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}\b`),
	},
	{
		name:    "URI with Credentials",
		pattern: regexp.MustCompile(`\b[a-zA-Z][a-zA-Z0-9+.-]*://[^/\s:@]+:[^/\s:@]+@[^/\s]+\b`),
	},
}

// ScanTextForSecrets returns the names of all secret patterns detected in text.
func ScanTextForSecrets(text string) []string {
	if text == "" {
		return nil
	}
	var detected []string
	for _, rule := range secretValuePatterns {
		if rule.pattern.MatchString(text) {
			detected = append(detected, rule.name)
		}
	}
	return detected
}

// CheckNoSecrets inspects one or more strings and returns an error if any secret pattern is matched.
func CheckNoSecrets(texts ...string) error {
	var found []string
	for _, t := range texts {
		for _, name := range ScanTextForSecrets(t) {
			if !slices.Contains(found, name) {
				found = append(found, name)
			}
		}
	}
	if len(found) > 0 {
		return fmt.Errorf("sensitive secret pattern detected (%s); writing or caching sensitive values is denied", strings.Join(found, ", "))
	}
	return nil
}

// RedactSecretsInString replaces any matched secret value in text with RedactedMask.
func RedactSecretsInString(text string) string {
	if text == "" {
		return ""
	}
	out := text
	for _, rule := range secretValuePatterns {
		out = rule.pattern.ReplaceAllString(out, RedactedMask)
	}
	return out
}
