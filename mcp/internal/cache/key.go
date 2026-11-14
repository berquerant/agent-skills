package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

var cacheKeyRegex = regexp.MustCompile(`^[0-9a-zA-Z_-]{2,64}$`)

// VersionKind identifies what kind of version target was requested.
type VersionKind int

const (
	VersionNone VersionKind = iota
	VersionLatest
	VersionSpecific
)

// VersionTarget represents a parsed version selector (omitted/None, "latest", or integer version).
type VersionTarget struct {
	Kind   VersionKind
	Number int
}

// IntPtr returns a pointer to Number when Kind is VersionSpecific, or nil otherwise.
func (v VersionTarget) IntPtr() *int {
	if v.Kind == VersionSpecific {
		n := v.Number
		return &n
	}
	return nil
}

// IsValidCacheKey reports whether key matches ^[0-9a-zA-Z_-]{2,64}$.
func IsValidCacheKey(key string) bool {
	return cacheKeyRegex.MatchString(key)
}

// ParseVersionedKey parses '<base_key>', '<base_key>@latest', or '<base_key>@<version>'.
func ParseVersionedKey(key string) (string, VersionTarget, bool) {
	if strings.Contains(key, "@") {
		parts := strings.Split(key, "@")
		if len(parts) != 2 {
			return "", VersionTarget{}, false
		}
		baseKey, verStr := parts[0], parts[1]
		if !IsValidCacheKey(baseKey) {
			return "", VersionTarget{}, false
		}
		if verStr == "latest" {
			return baseKey, VersionTarget{Kind: VersionLatest}, true
		}
		if isDigits(verStr) {
			n, err := strconv.Atoi(verStr)
			if err == nil && n > 0 {
				return baseKey, VersionTarget{Kind: VersionSpecific, Number: n}, true
			}
		}
		return "", VersionTarget{}, false
	}
	if IsValidCacheKey(key) {
		return key, VersionTarget{Kind: VersionNone}, true
	}
	return "", VersionTarget{}, false
}

// ResolveVersionTarget combines an explicit version argument with the version suffix parsed from key.
// Explicit version takes precedence over keySuffix. Returns false if the explicit version is malformed.
func ResolveVersionTarget(explicit any, keySuffix VersionTarget) (VersionTarget, bool) {
	if explicit == nil {
		return keySuffix, true
	}
	switch v := explicit.(type) {
	case int:
		if v <= 0 {
			return VersionTarget{}, false
		}
		return VersionTarget{Kind: VersionSpecific, Number: v}, true
	case int64:
		if v <= 0 {
			return VersionTarget{}, false
		}
		return VersionTarget{Kind: VersionSpecific, Number: int(v)}, true
	case float64:
		if v <= 0 || float64(int(v)) != v {
			return VersionTarget{}, false
		}
		return VersionTarget{Kind: VersionSpecific, Number: int(v)}, true
	case string:
		if v == "" {
			return keySuffix, true
		}
		if v == "latest" {
			return VersionTarget{Kind: VersionLatest}, true
		}
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 {
			return VersionTarget{}, false
		}
		return VersionTarget{Kind: VersionSpecific, Number: n}, true
	default:
		return VersionTarget{}, false
	}
}

// FormatVersionedKey formats baseKey with an @version suffix when target is non-empty.
func FormatVersionedKey(baseKey string, target VersionTarget) string {
	switch target.Kind {
	case VersionLatest:
		return baseKey + "@latest"
	case VersionSpecific:
		return fmt.Sprintf("%s@%d", baseKey, target.Number)
	default:
		return baseKey
	}
}

// FormatVersionedKeyNumber formats baseKey@ver.
func FormatVersionedKeyNumber(baseKey string, ver int) string {
	return fmt.Sprintf("%s@%d", baseKey, ver)
}

// NormalizeDir resolves dirPath to an absolute canonical path even when leaf directories do not exist.
func NormalizeDir(dirPath string) string {
	if dirPath == "" {
		dirPath = "."
	}
	abs, err := filepath.Abs(dirPath)
	if err != nil {
		abs = filepath.Clean(dirPath)
	}
	return resolveExistingPrefix(abs)
}

func resolveExistingPrefix(absPath string) string {
	curr := filepath.Clean(absPath)
	var tail []string
	for {
		resolved, err := filepath.EvalSymlinks(curr)
		if err == nil {
			for _, t := range slices.Backward(tail) {
				resolved = filepath.Join(resolved, t)
			}
			return filepath.Clean(resolved)
		}
		parent := filepath.Dir(curr)
		if parent == curr {
			return filepath.Clean(absPath)
		}
		tail = append(tail, filepath.Base(curr))
		curr = parent
	}
}

// ResolveKey returns explicitKey if non-empty, or computes the SHA-256 cache key when typ, name, and dir are non-empty.
func ResolveKey(explicitKey, typ, name, dir string, args []string, stdin *string) (string, bool) {
	if explicitKey != "" {
		return explicitKey, true
	}
	if typ != "" && name != "" && dir != "" {
		return ComputeCacheKey(typ, name, dir, args, stdin), true
	}
	return "", false
}

type keyPayload struct {
	Args  []string `json:"args"`
	Dir   string   `json:"dir"`
	Name  string   `json:"name"`
	Stdin *string  `json:"stdin"`
	Type  string   `json:"type"`
}

// ComputeCacheKey computes a deterministic 64-character SHA-256 hex digest from the invocation parameters.
func ComputeCacheKey(typ, name, dir string, args []string, stdin *string) string {
	if args == nil {
		args = []string{}
	}
	payload := keyPayload{
		Args:  args,
		Dir:   NormalizeDir(dir),
		Name:  name,
		Stdin: stdin,
		Type:  typ,
	}
	b, _ := json.Marshal(payload)
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for i := range len(s) {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}
