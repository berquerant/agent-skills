package guard

import (
	"fmt"
	"slices"
	"strings"
)

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
