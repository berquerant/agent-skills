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

	t.Run("multiple goroutines sharing same FileLock can RLock concurrently", func(t *testing.T) {
		sharedLock := store.NewFileLock(lockPath, 500*time.Millisecond)
		const readers = 5
		var wg sync.WaitGroup
		wg.Add(readers)
		for range readers {
			go func() {
				defer wg.Done()
				require.NoError(t, sharedLock.RLock())
				time.Sleep(20 * time.Millisecond)
				require.NoError(t, sharedLock.Unlock())
			}()
		}
		wg.Wait()
	})

	t.Run("unlocking an un-acquired lock returns error cleanly", func(t *testing.T) {
		unacquired := store.NewFileLock(lockPath, 100*time.Millisecond)
		err := unacquired.Unlock()
		require.ErrorContains(t, err, "cannot unlock un-acquired FileLock")
	})

	t.Run("RLock times out when exclusive lock is held on same FileLock", func(t *testing.T) {
		l := store.NewFileLock(lockPath, 50*time.Millisecond)
		require.NoError(t, l.Lock())
		defer func() { _ = l.Unlock() }()

		err := l.RLock()
		require.ErrorIs(t, err, store.ErrLockTimeout)
	})

	t.Run("Lock times out when reader holds lock on same FileLock", func(t *testing.T) {
		l := store.NewFileLock(lockPath, 50*time.Millisecond)
		require.NoError(t, l.RLock())
		defer func() { _ = l.Unlock() }()

		err := l.Lock()
		require.ErrorIs(t, err, store.ErrLockTimeout)
	})
}

func TestStoreEdgeCases(t *testing.T) {
	t.Run("CleanStaleTmpFiles returns 0 on non-existent dir", func(t *testing.T) {
		removed, err := store.CleanStaleTmpFiles(filepath.Join(t.TempDir(), "nonexistent"), time.Now(), time.Minute)
		require.NoError(t, err)
		assert.Equal(t, 0, removed)
	})

	t.Run("CleanStaleTmpFiles skips directories with .tmp_ prefix", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.Mkdir(filepath.Join(dir, ".tmp_dir"), 0o700))
		removed, err := store.CleanStaleTmpFiles(dir, time.Now(), 0)
		require.NoError(t, err)
		assert.Equal(t, 0, removed)
	})

	t.Run("WriteAtomic returns error when parent dir is not a directory", func(t *testing.T) {
		fileAsDir := filepath.Join(t.TempDir(), "file")
		require.NoError(t, os.WriteFile(fileAsDir, []byte("x"), 0o600))
		err := store.WriteAtomic(filepath.Join(fileAsDir, "sub", "test.txt"), []byte("data"), 0o600)
		require.Error(t, err)
	})

	t.Run("Lock returns error when lock dir cannot be created", func(t *testing.T) {
		fileAsDir := filepath.Join(t.TempDir(), "file")
		require.NoError(t, os.WriteFile(fileAsDir, []byte("x"), 0o600))
		l := store.NewFileLock(filepath.Join(fileAsDir, "sub", "lock"), 50*time.Millisecond)
		err := l.Lock()
		require.Error(t, err)
	})

	t.Run("RLock returns error when lock dir cannot be created", func(t *testing.T) {
		fileAsDir := filepath.Join(t.TempDir(), "file")
		require.NoError(t, os.WriteFile(fileAsDir, []byte("x"), 0o600))
		l := store.NewFileLock(filepath.Join(fileAsDir, "sub", "lock"), 50*time.Millisecond)
		err := l.RLock()
		require.Error(t, err)
	})

	t.Run("GenerateID produces valid hex string", func(t *testing.T) {
		id, err := store.GenerateID()
		require.NoError(t, err)
		assert.Len(t, id, 32)
	})

	t.Run("NewFileLock with zero timeout sets default", func(t *testing.T) {
		l := store.NewFileLock("/tmp/dummy.lock", 0)
		require.NotNil(t, l)
	})
}
