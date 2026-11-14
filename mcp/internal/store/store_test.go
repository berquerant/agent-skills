package store_test

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/berquerant/agent-skills/mcp/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteAtomicAndCleanStaleTmp(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "sub", "state.json")

	require.NoError(t, store.WriteAtomic(target, []byte(`{"v":1}`), 0o600))
	require.NoError(t, store.WriteAtomic(target, []byte(`{"v":2}`), 0o600))

	got, err := os.ReadFile(target)
	require.NoError(t, err)
	assert.Equal(t, `{"v":2}`, string(got))

	// Create a stale .tmp_ file and a fresh .tmp_ file
	staleTmp := filepath.Join(dir, "sub", ".tmp_1234_stale.json")
	freshTmp := filepath.Join(dir, "sub", ".tmp_1234_fresh.json")
	require.NoError(t, os.WriteFile(staleTmp, []byte("old"), 0o600))
	require.NoError(t, os.WriteFile(freshTmp, []byte("new"), 0o600))

	now := time.Now()
	oldTime := now.Add(-2 * time.Minute)
	require.NoError(t, os.Chtimes(staleTmp, oldTime, oldTime))

	removed, err := store.CleanStaleTmpFiles(dir, now, 60*time.Second)
	require.NoError(t, err)
	assert.Equal(t, 1, removed)

	_, err = os.Stat(staleTmp)
	assert.True(t, os.IsNotExist(err))
	_, err = os.Stat(freshTmp)
	assert.NoError(t, err)
}

func TestFileLockConcurrencyAndTimeout(t *testing.T) {
	dir := t.TempDir()
	lockPath := filepath.Join(dir, "locks", "lockfile")

	t.Run("concurrent exclusive writers serialize cleanly", func(t *testing.T) {
		lock := store.NewFileLock(lockPath, 5*time.Second)
		counterFile := filepath.Join(dir, "counter.txt")
		require.NoError(t, os.WriteFile(counterFile, []byte("0"), 0o600))

		const workers = 8
		var wg sync.WaitGroup
		wg.Add(workers)
		for i := range workers {
			go func(id int) {
				defer wg.Done()
				require.NoError(t, lock.Lock())
				defer func() {
					require.NoError(t, lock.Unlock())
				}()
				data, err := os.ReadFile(counterFile)
				require.NoError(t, err)
				var n int
				_, _ = fmt.Sscanf(string(data), "%d", &n)
				require.NoError(t, store.WriteAtomic(counterFile, []byte(fmt.Sprintf("%d", n+1)), 0o600))
				_ = id
			}(i)
		}
		wg.Wait()

		final, err := os.ReadFile(counterFile)
		require.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("%d", workers), string(final))
	})

	t.Run("shared locks succeed concurrently while exclusive lock times out", func(t *testing.T) {
		r1 := store.NewFileLock(lockPath, 500*time.Millisecond)
		r2 := store.NewFileLock(lockPath, 500*time.Millisecond)
		w := store.NewFileLock(lockPath, 80*time.Millisecond)

		require.NoError(t, r1.RLock())
		defer func() { _ = r1.Unlock() }()

		require.NoError(t, r2.RLock())
		defer func() { _ = r2.Unlock() }()

		err := w.Lock()
		require.ErrorIs(t, err, store.ErrLockTimeout)
	})

	t.Run("unlocking an un-acquired lock returns error cleanly", func(t *testing.T) {
		unacquired := store.NewFileLock(lockPath, 100*time.Millisecond)
		err := unacquired.Unlock()
		require.ErrorContains(t, err, "cannot unlock un-acquired FileLock")
	})
}
