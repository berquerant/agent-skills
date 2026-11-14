package cache_test

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/berquerant/agent-skills/mcp/internal/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type manualClock struct {
	mu sync.Mutex
	t  time.Time
}

func newManualClock(unixSec int64) *manualClock {
	return &manualClock{t: time.Unix(unixSec, 0).UTC()}
}

func (c *manualClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.t
}

func (c *manualClock) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.t = c.t.Add(d)
}

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
			name:        "empty args vs nil args produce identical key",
			t1:          "cmd",
			n1:          "ls",
			d1:          tmp,
			args1:       []string{},
			t2:          "cmd",
			n2:          "ls",
			d2:          tmp,
			args2:       nil,
			shouldMatch: true,
		},
		{
			name:        "trailing slash and non-existent leaf normalize identically",
			t1:          "cmd",
			n1:          "ls",
			d1:          filepath.Join(tmp, "nonexistent", "leaf"),
			args1:       []string{"-la"},
			t2:          "cmd",
			n2:          "ls",
			d2:          filepath.Join(tmp, "nonexistent", "leaf") + "/",
			args2:       []string{"-la"},
			shouldMatch: true,
		},
		{
			name:        "nil stdin vs literal 'null' string produce distinct keys",
			t1:          "cmd",
			n1:          "cat",
			d1:          tmp,
			stdin1:      nil,
			t2:          "cmd",
			n2:          "cat",
			d2:          tmp,
			stdin2:      &strNull,
			shouldMatch: false,
		},
		{
			name:        "delimiter-like characters in args do not collide",
			t1:          "cmd",
			n1:          "echo",
			d1:          tmp,
			args1:       []string{"a__k__b", "c"},
			t2:          "cmd",
			n2:          "echo",
			d2:          tmp,
			args2:       []string{"a", "b__k__c"},
			shouldMatch: false,
		},
		{
			name:        "argument order changes key",
			t1:          "cmd",
			n1:          "echo",
			d1:          tmp,
			args1:       []string{"1", "2"},
			t2:          "cmd",
			n2:          "echo",
			d2:          tmp,
			args2:       []string{"2", "1"},
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
		{
			name:   "invalid zero version",
			key:    "abc_123@0",
			wantOK: false,
		},
		{
			name:   "invalid negative version",
			key:    "abc_123@-1",
			wantOK: false,
		},
		{
			name:   "invalid path traversal key",
			key:    "../secret@1",
			wantOK: false,
		},
		{
			name:   "invalid single char key",
			key:    "a",
			wantOK: false,
		},
		{
			name:   "invalid 65 char key",
			key:    strings.Repeat("a", 65),
			wantOK: false,
		},
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
		explicit any
		suffix   cache.VersionTarget
		wantOK   bool
		wantKind cache.VersionKind
		wantNum  int
	}{
		{
			name:     "nil explicit falls back to suffix",
			explicit: nil,
			suffix:   suffixLatest,
			wantOK:   true,
			wantKind: cache.VersionLatest,
		},
		{
			name:     "empty string explicit falls back to suffix",
			explicit: "",
			suffix:   suffixLatest,
			wantOK:   true,
			wantKind: cache.VersionLatest,
		},
		{
			name:     "positive int overrides suffix",
			explicit: 3,
			suffix:   suffixLatest,
			wantOK:   true,
			wantKind: cache.VersionSpecific,
			wantNum:  3,
		},
		{
			name:     "positive int64 overrides suffix",
			explicit: int64(5),
			suffix:   suffixLatest,
			wantOK:   true,
			wantKind: cache.VersionSpecific,
			wantNum:  5,
		},
		{
			name:     "integral float64 overrides suffix",
			explicit: float64(2),
			suffix:   suffixLatest,
			wantOK:   true,
			wantKind: cache.VersionSpecific,
			wantNum:  2,
		},
		{
			name:     "numeric string overrides suffix",
			explicit: "4",
			suffix:   suffixLatest,
			wantOK:   true,
			wantKind: cache.VersionSpecific,
			wantNum:  4,
		},
		{
			name:     "latest string overrides suffix",
			explicit: "latest",
			suffix:   cache.VersionTarget{Kind: cache.VersionSpecific, Number: 1},
			wantOK:   true,
			wantKind: cache.VersionLatest,
		},
		{
			name:     "zero int rejected",
			explicit: 0,
			wantOK:   false,
		},
		{
			name:     "negative int64 rejected",
			explicit: int64(-1),
			wantOK:   false,
		},
		{
			name:     "zero float64 rejected",
			explicit: float64(0),
			wantOK:   false,
		},
		{
			name:     "fractional float64 rejected",
			explicit: 1.5,
			wantOK:   false,
		},
		{
			name:     "zero string rejected",
			explicit: "0",
			wantOK:   false,
		},
		{
			name:     "non-numeric string rejected",
			explicit: "invalid",
			wantOK:   false,
		},
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

func makeEntry(stdout string, createdTS int64, ttl int) cache.CacheEntry {
	sc := 0
	ss := "OK"
	return cache.CacheEntry{
		Type:             "command",
		Name:             "test",
		Args:             []string{"arg1"},
		Dir:              "/tmp",
		CreatedTimestamp: createdTS,
		TTL:              ttl,
		TTLTimestamp:     createdTS + int64(ttl),
		Stdout:           stdout,
		Stderr:           "",
		Success:          true,
		StatusCode:       &sc,
		StatusString:     &ss,
	}
}

func TestStorageVersioningTTLAndEviction(t *testing.T) {
	t.Run("immutable versioning, @latest, specific version, and delete", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "cache")
		clk := newManualClock(1000)
		st, err := cache.NewStorage(cache.Config{
			Dir:                         dir,
			MaxTTL:                      1800,
			MaxCacheSize:                1024 * 1024,
			MaxTotalCompressedCacheSize: 10 * 1024 * 1024,
			Clock:                       clk,
		})
		require.NoError(t, err)
		defer func() { _ = st.Close() }()

		key := "imm_key"
		for i, out := range []string{"v1", "v2", "v3"} {
			res, setErr := st.Set(key, makeEntry(out, 1000+int64(i), 300))
			require.NoError(t, setErr)
			require.True(t, res.Success)
			require.NotNil(t, res.Version)
			assert.Equal(t, i+1, *res.Version)
			require.NotNil(t, res.Key)
			assert.Equal(t, fmt.Sprintf("%s@%d", key, i+1), *res.Key)
		}

		for _, qKey := range []string{key, key + "@latest"} {
			got, getErr := st.Get(qKey, nil)
			require.NoError(t, getErr)
			require.True(t, got.Hit)
			require.NotNil(t, got.Version)
			assert.Equal(t, 3, *got.Version)
			require.NotNil(t, got.Data)
			assert.Equal(t, "v3", got.Data.Stdout)
		}

		gotV1, err := st.Get(key+"@1", nil)
		require.NoError(t, err)
		require.True(t, gotV1.Hit)
		assert.Equal(t, "v1", gotV1.Data.Stdout)

		gotV2, err := st.Get(key, 2)
		require.NoError(t, err)
		require.True(t, gotV2.Hit)
		assert.Equal(t, "v2", gotV2.Data.Stdout)

		verRes, err := st.Version(key)
		require.NoError(t, err)
		assert.True(t, verRes.Hit)
		require.Len(t, verRes.Versions, 3)

		delV2, err := st.Delete(key, 2)
		require.NoError(t, err)
		assert.True(t, delV2.Deleted)
		assert.Equal(t, 1, delV2.DeletedCount)

		existV2, err := st.Exist(key+"@2", nil)
		require.NoError(t, err)
		assert.False(t, existV2)

		delLatest, err := st.Delete(key+"@latest", nil)
		require.NoError(t, err)
		assert.True(t, delLatest.Deleted)
		assert.Equal(t, 3, delLatest.Version)

		delAll, err := st.Delete(key, nil)
		require.NoError(t, err)
		assert.True(t, delAll.Deleted)
		assert.Equal(t, 1, delAll.DeletedCount)
	})

	t.Run("lazy expiration cleanup and purge", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "cache")
		clk := newManualClock(1000)
		st, err := cache.NewStorage(cache.Config{
			Dir:                         dir,
			MaxTTL:                      1800,
			MaxCacheSize:                1024 * 1024,
			MaxTotalCompressedCacheSize: 10 * 1024 * 1024,
			Clock:                       clk,
		})
		require.NoError(t, err)
		defer func() { _ = st.Close() }()

		_, err = st.Set("exp_key", makeEntry("short_ttl", 1000, 10))
		require.NoError(t, err)
		_, err = st.Set("live_key", makeEntry("long_ttl", 1000, 600))
		require.NoError(t, err)

		clk.Advance(20 * time.Second)

		purgeRes, err := st.Purge(false)
		require.NoError(t, err)
		assert.True(t, purgeRes.Success)
		assert.Equal(t, 1, purgeRes.DeletedCount)

		exExp, _ := st.Exist("exp_key", nil)
		exLive, _ := st.Exist("live_key", nil)
		assert.False(t, exExp)
		assert.True(t, exLive)

		purgeAll, err := st.Purge(true)
		require.NoError(t, err)
		assert.True(t, purgeAll.Success)
		assert.Equal(t, 1, purgeAll.DeletedCount)
		assert.Equal(t, int64(0), st.TotalCompressedSize())
	})

	t.Run("capacity limit and FIFO eviction", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "cache")
		clk := newManualClock(1000)
		st, err := cache.NewStorage(cache.Config{
			Dir:                         dir,
			MaxTTL:                      1800,
			MaxCacheSize:                300,
			MaxTotalCompressedCacheSize: 450,
			Clock:                       clk,
		})
		require.NoError(t, err)
		defer func() { _ = st.Close() }()

		tooBig, err := st.Set("big_key", makeEntry(strings.Repeat("A", 300), 1000, 300))
		require.NoError(t, err)
		assert.False(t, tooBig.Success)
		require.NotNil(t, tooBig.Message)
		assert.Contains(t, *tooBig.Message, "Exceeds max_cache_size")

		r1, err := st.Set("k_old", makeEntry("old", 980, 300))
		require.NoError(t, err)
		require.True(t, r1.Success)

		r2, err := st.Set("k_mid", makeEntry("mid", 990, 300))
		require.NoError(t, err)
		require.True(t, r2.Success)

		r3, err := st.Set("k_new", makeEntry("new", 1000, 300))
		require.NoError(t, err)
		require.True(t, r3.Success)

		exOld, _ := st.Exist("k_old", nil)
		exMid, _ := st.Exist("k_mid", nil)
		exNew, _ := st.Exist("k_new", nil)
		assert.False(t, exOld, "oldest entry should be evicted first")
		assert.True(t, exMid)
		assert.True(t, exNew)
	})

	t.Run("concurrent set and get serialize safely under -race", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "cache")
		clk := newManualClock(2000)
		st, err := cache.NewStorage(cache.Config{
			Dir:                         dir,
			MaxTTL:                      1800,
			MaxCacheSize:                1024 * 1024,
			MaxTotalCompressedCacheSize: 10 * 1024 * 1024,
			Clock:                       clk,
		})
		require.NoError(t, err)
		defer func() { _ = st.Close() }()

		const workers = 8
		var wg sync.WaitGroup
		wg.Add(workers)
		for i := range workers {
			go func(id int) {
				defer wg.Done()
				res, setErr := st.Set("shared_key", makeEntry(fmt.Sprintf("out-%d", id), 2000+int64(id), 300))
				require.NoError(t, setErr)
				assert.True(t, res.Success)

				got, getErr := st.Get("shared_key", nil)
				require.NoError(t, getErr)
				assert.True(t, got.Hit)
			}(i)
		}
		wg.Wait()

		verRes, err := st.Version("shared_key")
		require.NoError(t, err)
		assert.Len(t, verRes.Versions, workers)
	})
}

func TestServiceOperations(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "cache")
	clk := newManualClock(5000)
	svc, err := cache.NewService(cache.Config{
		Dir:                         dir,
		MaxTTL:                      1800,
		MaxCacheSize:                1024 * 1024,
		MaxTotalCompressedCacheSize: 10 * 1024 * 1024,
		Clock:                       clk,
	})
	require.NoError(t, err)
	defer func() { _ = svc.Close() }()

	setTests := []struct {
		name        string
		in          cache.SetInput
		wantSuccess bool
		wantMsgSub  string
	}{
		{
			name: "rejects non-positive ttl",
			in: cache.SetInput{
				Type:   "command",
				Name:   "echo",
				Dir:    "/tmp",
				TTL:    0,
				Stdout: "hi",
			},
			wantSuccess: false,
			wantMsgSub:  "ttl must be > 0",
		},
		{
			name: "stores v1 by computed key",
			in: cache.SetInput{
				Type:    "command",
				Name:    "echo",
				Dir:     "/tmp",
				Args:    []string{"hello"},
				TTL:     300,
				Stdout:  "hello-1\n",
				Success: true,
			},
			wantSuccess: true,
		},
		{
			name: "stores v2 by explicit key",
			in: cache.SetInput{
				Key:     "svc_key",
				Type:    "command",
				Name:    "echo",
				Dir:     "/tmp",
				TTL:     300,
				Stdout:  "svc-v1",
				Success: true,
			},
			wantSuccess: true,
		},
		{
			name: "stores second version for explicit key",
			in: cache.SetInput{
				Key:     "svc_key",
				Type:    "command",
				Name:    "echo",
				Dir:     "/tmp",
				TTL:     300,
				Stdout:  "svc-v2",
				Success: true,
			},
			wantSuccess: true,
		},
	}

	for _, tt := range setTests {
		t.Run("Set/"+tt.name, func(t *testing.T) {
			res := svc.Set(tt.in)
			assert.Equal(t, tt.wantSuccess, res["success"])
			if tt.wantMsgSub != "" {
				assert.Contains(t, fmt.Sprint(res["message"]), tt.wantMsgSub)
			}
		})
	}

	queryTests := []struct {
		name       string
		op         string
		in         cache.QueryInput
		wantHit    bool
		wantReason string
	}{
		{
			name:    "get missing key and triple",
			op:      "get",
			in:      cache.QueryInput{},
			wantHit: false,
		},
		{
			name:       "get invalid explicit version 0",
			op:         "get",
			in:         cache.QueryInput{Key: "svc_key", Version: 0},
			wantHit:    false,
			wantReason: "invalid_version",
		},
		{
			name:    "get hit latest for svc_key",
			op:      "get",
			in:      cache.QueryInput{Key: "svc_key"},
			wantHit: true,
		},
		{
			name:    "exist missing key and triple",
			op:      "exist",
			in:      cache.QueryInput{},
			wantHit: false,
		},
		{
			name:    "exist hit by triple",
			op:      "exist",
			in:      cache.QueryInput{Type: "command", Name: "echo", Dir: "/tmp", Args: []string{"hello"}},
			wantHit: true,
		},
		{
			name:    "delete missing key and triple",
			op:      "delete",
			in:      cache.QueryInput{},
			wantHit: false,
		},
		{
			name:    "delete specific version 1 of svc_key",
			op:      "delete",
			in:      cache.QueryInput{Key: "svc_key", Version: 1},
			wantHit: true,
		},
	}

	for _, tt := range queryTests {
		t.Run("Query/"+tt.name, func(t *testing.T) {
			var res map[string]any
			switch tt.op {
			case "get":
				res = svc.Get(tt.in)
			case "exist":
				res = svc.Exist(tt.in)
			case "delete":
				res = svc.Delete(tt.in)
			}
			assert.Equal(t, tt.wantHit, res["hit"])
			if tt.wantReason != "" {
				assert.Equal(t, tt.wantReason, res["reason"])
			}
		})
	}

	verMissing := svc.Version(cache.VersionInput{})
	assert.Equal(t, false, verMissing["hit"])
	verHit := svc.Version(cache.VersionInput{Key: "svc_key"})
	assert.Equal(t, true, verHit["hit"])

	purgeRes := svc.Purge(cache.PurgeInput{All: true})
	assert.Equal(t, true, purgeRes["success"])
}
