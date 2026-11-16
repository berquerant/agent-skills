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

func TestStorageVersioningAndDelete(t *testing.T) {
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

	gotV2, err := st.Get(key, new(2))
	require.NoError(t, err)
	require.True(t, gotV2.Hit)
	assert.Equal(t, "v2", gotV2.Data.Stdout)

	verRes, err := st.Version(key)
	require.NoError(t, err)
	assert.True(t, verRes.Hit)
	require.Len(t, verRes.Versions, 3)

	delV2, err := st.Delete(key, new(2))
	require.NoError(t, err)
	assert.True(t, delV2.Deleted)
	assert.Equal(t, 1, delV2.DeletedCount)

	existV2, err := st.Exist(key+"@2", nil)
	require.NoError(t, err)
	assert.False(t, existV2)

	delLatest, err := st.Delete(key+"@latest", nil)
	require.NoError(t, err)
	assert.True(t, delLatest.Deleted)
	require.NotNil(t, delLatest.Version)
	assert.Equal(t, 3, *delLatest.Version)

	delAll, err := st.Delete(key, nil)
	require.NoError(t, err)
	assert.True(t, delAll.Deleted)
	assert.Equal(t, 1, delAll.DeletedCount)
}

func TestStorageExpirationAndPurge(t *testing.T) {
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
}

func TestStorageCapacityAndEviction(t *testing.T) {
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

	for _, item := range []struct {
		key string
		out string
		ts  int64
	}{
		{key: "k_old", out: "old", ts: 980},
		{key: "k_mid", out: "mid", ts: 990},
		{key: "k_new", out: "new", ts: 1000},
	} {
		r, setErr := st.Set(item.key, makeEntry(item.out, item.ts, 300))
		require.NoError(t, setErr)
		require.True(t, r.Success)
	}

	exOld, _ := st.Exist("k_old", nil)
	exMid, _ := st.Exist("k_mid", nil)
	exNew, _ := st.Exist("k_new", nil)
	assert.False(t, exOld, "oldest entry should be evicted first")
	assert.True(t, exMid)
	assert.True(t, exNew)
}

func TestStorageConcurrency(t *testing.T) {
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
}
