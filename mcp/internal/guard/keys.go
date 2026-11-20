package guard

import (
	"encoding/json"
	"regexp"
	"strings"
)

// RedactedMask is the replacement string used when scrubbing sensitive values from logs.
const RedactedMask = "***REDACTED***"

// SensitiveKeySubstrings lists key/variable name substrings treated as sensitive during log scrubbing.
// Derived from standard sanitizer rules and Django HiddenSettings (API|TOKEN|KEY|SECRET|PASS|SIGNATURE|HTTP_COOKIE).
var SensitiveKeySubstrings = []string{
	"password",
	"passwd",
	"passphrase",
	"secret",
	"token",
	"api_key",
	"apikey",
	"access_key",
	"private_key",
	"client_secret",
	"auth",
	"credential",
	"session_id",
	"cookie",
	"http_cookie",
	"signature",
	"webhook",
	"database_url",
	"connection_string",
}

var sensitiveKeyRegex = compileSensitiveKeyRegex(SensitiveKeySubstrings)

func compileSensitiveKeyRegex(substrings []string) *regexp.Regexp {
	escaped := make([]string, len(substrings))
	for i, s := range substrings {
		escaped[i] = regexp.QuoteMeta(s)
	}
	return regexp.MustCompile(`(?i)(` + strings.Join(escaped, "|") + `)`)
}

// IsSensitiveKey reports whether a map key or field name contains a sensitive keyword.
func IsSensitiveKey(key string) bool {
	return sensitiveKeyRegex.MatchString(key)
}

// SanitizeForLog converts an arbitrary value into a JSON-compatible structure with all sensitive keys
// and embedded secret values replaced by RedactedMask.
func SanitizeForLog(v any) any {
	if v == nil {
		return nil
	}
	if errVal, ok := v.(error); ok {
		return RedactSecretsInString(errVal.Error())
	}
	raw, err := json.Marshal(v)
	if err != nil {
		return RedactedMask
	}
	var generic any
	if err := json.Unmarshal(raw, &generic); err != nil {
		return RedactSecretsInString(string(raw))
	}
	return sanitizeNode(generic)
}

func sanitizeNode(node any) any {
	switch x := node.(type) {
	case map[string]any:
		out := make(map[string]any, len(x))
		for k, val := range x {
			if IsSensitiveKey(k) {
				out[k] = RedactedMask
			} else {
				out[k] = sanitizeNode(val)
			}
		}
		return out
	case []any:
		out := make([]any, len(x))
		for i, elem := range x {
			out[i] = sanitizeNode(elem)
		}
		return out
	case string:
		return RedactSecretsInString(x)
	default:
		return x
	}
}
