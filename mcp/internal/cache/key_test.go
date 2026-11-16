package cache_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/berquerant/agent-skills/mcp/internal/cache"
	"github.com/stretchr/testify/assert"
)

func TestComputeCacheKeyDeterminism(t *testing.T) {
	tmp := t.TempDir()
	strNull := "null"

	tests := []struct {
		name        string
		t1, n1, d1  string
		args1       []string
		stdin1      *string
		t2, n2, d2  string
		args2       []string
		stdin2      *string
		shouldMatch bool
	}{
		{
			name: "empty args vs nil args produce identical key",
			t1:   "cmd", n1: "ls", d1: tmp, args1: []string{},
			t2: "cmd", n2: "ls", d2: tmp, args2: nil,
			shouldMatch: true,
		},
		{
			name: "trailing slash and non-existent leaf normalize identically",
			t1:   "cmd", n1: "ls", d1: filepath.Join(tmp, "nonexistent", "leaf"), args1: []string{"-la"},
			t2: "cmd", n2: "ls", d2: filepath.Join(tmp, "nonexistent", "leaf") + "/", args2: []string{"-la"},
			shouldMatch: true,
		},
		{
			name: "nil stdin vs literal 'null' string produce distinct keys",
			t1:   "cmd", n1: "cat", d1: tmp, stdin1: nil,
			t2: "cmd", n2: "cat", d2: tmp, stdin2: &strNull,
			shouldMatch: false,
		},
		{
			name: "delimiter-like characters in args do not collide",
			t1:   "cmd", n1: "echo", d1: tmp, args1: []string{"a__k__b", "c"},
			t2: "cmd", n2: "echo", d2: tmp, args2: []string{"a", "b__k__c"},
			shouldMatch: false,
		},
		{
			name: "argument order changes key",
			t1:   "cmd", n1: "echo", d1: tmp, args1: []string{"1", "2"},
			t2: "cmd", n2: "echo", d2: tmp, args2: []string{"2", "1"},
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k1 := cache.ComputeCacheKey(tt.t1, tt.n1, tt.d1, tt.args1, tt.stdin1)
			k2 := cache.ComputeCacheKey(tt.t2, tt.n2, tt.d2, tt.args2, tt.stdin2)
			assert.Len(t, k1, 64)
			assert.Len(t, k2, 64)
			if tt.shouldMatch {
				assert.Equal(t, k1, k2)
			} else {
				assert.NotEqual(t, k1, k2)
			}
		})
	}
}

func TestKeyValidationAndVersionParsing(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		wantOK      bool
		wantBase    string
		wantVerKind cache.VersionKind
		wantVerNum  int
	}{
		{
			name:        "unversioned base key",
			key:         "abc_123-xyz",
			wantOK:      true,
			wantBase:    "abc_123-xyz",
			wantVerKind: cache.VersionNone,
		},
		{
			name:        "versioned positive int",
			key:         "abc_123@42",
			wantOK:      true,
			wantBase:    "abc_123",
			wantVerKind: cache.VersionSpecific,
			wantVerNum:  42,
		},
		{
			name:        "versioned latest",
			key:         "abc_123@latest",
			wantOK:      true,
			wantBase:    "abc_123",
			wantVerKind: cache.VersionLatest,
		},
		{name: "invalid zero version", key: "abc_123@0", wantOK: false},
		{name: "invalid negative version", key: "abc_123@-1", wantOK: false},
		{name: "invalid path traversal key", key: "../secret@1", wantOK: false},
		{name: "invalid single char key", key: "a", wantOK: false},
		{name: "invalid 65 char key", key: strings.Repeat("a", 65), wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base, ver, ok := cache.ParseVersionedKey(tt.key)
			assert.Equal(t, tt.wantOK, ok)
			if !tt.wantOK {
				return
			}
			assert.Equal(t, tt.wantBase, base)
			assert.Equal(t, tt.wantVerKind, ver.Kind)
			assert.Equal(t, tt.wantVerNum, ver.Number)
			assert.Equal(t, tt.key, cache.FormatVersionedKey(base, ver))
		})
	}
}

func TestResolveVersionTarget(t *testing.T) {
	suffixLatest := cache.VersionTarget{Kind: cache.VersionLatest}
	tests := []struct {
		name     string
		explicit *int
		suffix   cache.VersionTarget
		wantOK   bool
		wantKind cache.VersionKind
		wantNum  int
	}{
		{name: "nil explicit uses suffix", explicit: nil, suffix: suffixLatest, wantOK: true, wantKind: cache.VersionLatest},
		{name: "positive int overrides", explicit: new(3), suffix: suffixLatest, wantOK: true, wantKind: cache.VersionSpecific, wantNum: 3},
		{name: "zero int rejected", explicit: new(0), wantOK: false},
		{name: "negative int rejected", explicit: new(-1), wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := cache.ResolveVersionTarget(tt.explicit, tt.suffix)
			assert.Equal(t, tt.wantOK, ok)
			if !tt.wantOK {
				return
			}
			assert.Equal(t, tt.wantKind, got.Kind)
			assert.Equal(t, tt.wantNum, got.Number)
		})
	}
}
