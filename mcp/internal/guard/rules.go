package guard

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

//go:embed rules.json
var embeddedRulesJSON []byte

type namedPatternConfig struct {
	Name    string `json:"name"`
	Pattern string `json:"pattern"`
}

type rulesConfig struct {
	SensitiveFilenames     []string             `json:"sensitive_filenames"`
	SensitiveExtensions    []string             `json:"sensitive_extensions"`
	SensitivePathPatterns  []namedPatternConfig `json:"sensitive_path_patterns"`
	SensitiveKeySubstrings []string             `json:"sensitive_key_substrings"`
	SecretValuePatterns    []namedPatternConfig `json:"secret_value_patterns"`
}

type compiledRule struct {
	name    string
	pattern *regexp.Regexp
}

// Loaded rule sets populated from embedded rules.json at package initialization.
var (
	// SensitiveFilenames lists exact filenames blocked from reading, caching, or writing.
	SensitiveFilenames map[string]struct{}
	// SensitiveExtensions lists file extensions associated with private keys, certificates, and keystores.
	SensitiveExtensions map[string]struct{}
	// SensitiveKeySubstrings lists key/variable name substrings treated as sensitive during log scrubbing.
	SensitiveKeySubstrings []string

	sensitivePathPatterns []compiledRule
	secretValuePatterns   []compiledRule
	sensitiveKeyRegex     *regexp.Regexp
)

func init() {
	if err := loadEmbeddedRules(); err != nil {
		panic(fmt.Sprintf("guard: failed to load embedded rules.json: %v", err))
	}
}

func loadEmbeddedRules() error {
	var cfg rulesConfig
	if err := json.Unmarshal(embeddedRulesJSON, &cfg); err != nil {
		return fmt.Errorf("decode rules.json: %w", err)
	}

	SensitiveFilenames = toStringSet(cfg.SensitiveFilenames)
	SensitiveExtensions = toStringSet(cfg.SensitiveExtensions)
	SensitiveKeySubstrings = cfg.SensitiveKeySubstrings

	var err error
	if sensitivePathPatterns, err = compileNamedRules("sensitive_path_patterns", cfg.SensitivePathPatterns); err != nil {
		return err
	}
	if secretValuePatterns, err = compileNamedRules("secret_value_patterns", cfg.SecretValuePatterns); err != nil {
		return err
	}
	if len(SensitiveKeySubstrings) == 0 {
		return fmt.Errorf("sensitive_key_substrings must not be empty")
	}
	sensitiveKeyRegex = compileSensitiveKeyRegex(SensitiveKeySubstrings)
	return nil
}

func toStringSet(items []string) map[string]struct{} {
	set := make(map[string]struct{}, len(items))
	for _, item := range items {
		if trimmed := strings.ToLower(strings.TrimSpace(item)); trimmed != "" {
			set[trimmed] = struct{}{}
		}
	}
	return set
}

func compileNamedRules(section string, items []namedPatternConfig) ([]compiledRule, error) {
	rules := make([]compiledRule, 0, len(items))
	for i, item := range items {
		if strings.TrimSpace(item.Name) == "" {
			return nil, fmt.Errorf("%s[%d]: name must not be empty", section, i)
		}
		re, err := regexp.Compile(item.Pattern)
		if err != nil {
			return nil, fmt.Errorf("%s[%d] (%s): compile pattern: %w", section, i, item.Name, err)
		}
		rules = append(rules, compiledRule{name: item.Name, pattern: re})
	}
	return rules, nil
}
