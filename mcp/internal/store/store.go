// Package store provides foundational filesystem primitives for safe local persistence,
// atomic writes, cross-process/multi-goroutine file locking, and stale file cleanup.
//
// Security & Correctness Invariants:
//  1. All directory operations use [os.OpenRoot] to guarantee directory containment and
//     prevent symlink traversal attacks (TOCTOU).
//  2. Atomic writes write to a PID- and random ID-scoped temporary file in the target
//     directory before issuing an atomic rename, ensuring callers never observe partial writes.
//  3. File locking combines POSIX flock (for inter-process coordination) with sync.Mutex
//     (for goroutine-level synchronization within the same process).
package store

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ErrLockTimeout is returned when acquiring a file lock exceeds the configured timeout.
var ErrLockTimeout = errors.New("acquire file lock timed out")

// GenerateID returns a random 16-byte (32 hex chars) identifier.
func GenerateID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("read random bytes: %w", err)
	}
	return hex.EncodeToString(b[:]), nil
}

// WriteAtomic writes data to targetPath atomically by writing to a unique .tmp_<pid>_<id>
// file in the same directory, syncing, closing, and renaming over targetPath.
func WriteAtomic(targetPath string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(targetPath)
	base := filepath.Base(targetPath)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create parent dir: %w", err)
	}

	root, err := os.OpenRoot(dir)
	if err != nil {
		return fmt.Errorf("open parent root: %w", err)
	}
	defer func() { _ = root.Close() }()

	id, err := GenerateID()
	if err != nil {
		return err
	}
	ext := filepath.Ext(base)
	tmpName := fmt.Sprintf(".tmp_%d_%s%s", os.Getpid(), id, ext)

	f, err := root.OpenFile(tmpName, os.O_CREATE|os.O_EXCL|os.O_WRONLY, perm)
	if err != nil {
		return fmt.Errorf("open temp file: %w", err)
	}

	cleanup := func() {
		_ = f.Close()
		_ = root.Remove(tmpName)
	}

	if _, err := f.Write(data); err != nil {
		cleanup()
		return fmt.Errorf("write temp file: %w", err)
	}
	if err := f.Sync(); err != nil {
		cleanup()
		return fmt.Errorf("sync temp file: %w", err)
	}
	if err := f.Close(); err != nil {
		_ = root.Remove(tmpName)
		return fmt.Errorf("close temp file: %w", err)
	}

	if err := root.Rename(tmpName, base); err != nil {
		_ = root.Remove(tmpName)
		return fmt.Errorf("rename temp file to target: %w", err)
	}
	return nil
}

// CleanStaleTmpFiles removes any .tmp_* files under rootDir whose modification time
// is older than maxAge relative to now, using os.Root to prevent symlink TOCTOU traversal.
// It returns the number of removed files.
func CleanStaleTmpFiles(rootDir string, now time.Time, maxAge time.Duration) (int, error) {
	root, err := os.OpenRoot(rootDir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	defer func() { _ = root.Close() }()

	removed := 0
	walkErr := fs.WalkDir(root.FS(), ".", func(relPath string, d fs.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if d.IsDir() || !strings.HasPrefix(d.Name(), ".tmp_") {
			return nil
		}
		info, statErr := d.Info()
		if statErr != nil {
			return nil
		}
		if now.Sub(info.ModTime()) >= maxAge {
			if rmErr := root.Remove(relPath); rmErr == nil {
				removed++
			}
		}
		return nil
	})
	if os.IsNotExist(walkErr) {
		return removed, nil
	}
	return removed, walkErr
}

// FileLock provides combined in-process and cross-process (flock)
// mutual exclusion with a configurable timeout.
type FileLock struct {
	path         string
	timeout      time.Duration
	pollInterval time.Duration

	mu              sync.Mutex
	readers         int
	isExclusive     bool
	readerAcquiring bool
	f               *os.File
}

// NewFileLock creates a FileLock on lockPath with the given timeout.
func NewFileLock(lockPath string, timeout time.Duration) *FileLock {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &FileLock{
		path:         lockPath,
		timeout:      timeout,
		pollInterval: 10 * time.Millisecond,
	}
}

// Lock acquires an exclusive lock (`LOCK_EX`), blocking up to timeout.
func (l *FileLock) Lock() error {
	deadline := time.Now().Add(l.timeout)
	poll := l.pollInterval
	if poll <= 0 {
		poll = 10 * time.Millisecond
	}

	// Step 1: Wait until in-process exclusivity is available.
	for {
		l.mu.Lock()
		if !l.isExclusive && l.readers == 0 && !l.readerAcquiring {
			l.isExclusive = true
			l.mu.Unlock()
			break
		}
		l.mu.Unlock()
		if time.Now().After(deadline) {
			return fmt.Errorf("%w: in-process lock on %s", ErrLockTimeout, l.path)
		}
		time.Sleep(poll)
	}

	// Step 2: Acquire OS-level exclusive flock.
	dir := filepath.Dir(l.path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		l.mu.Lock()
		l.isExclusive = false
		l.mu.Unlock()
		return fmt.Errorf("create lock dir: %w", err)
	}

	f, err := os.OpenFile(l.path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		l.mu.Lock()
		l.isExclusive = false
		l.mu.Unlock()
		return fmt.Errorf("open lock file: %w", err)
	}

	for {
		err := flockExclusive(int(f.Fd()))
		if err == nil {
			l.mu.Lock()
			l.f = f
			l.mu.Unlock()
			return nil
		}
		if !isWouldBlock(err) {
			_ = f.Close()
			l.mu.Lock()
			l.isExclusive = false
			l.mu.Unlock()
			return fmt.Errorf("flock %s: %w", l.path, err)
		}
		if time.Now().After(deadline) {
			_ = f.Close()
			l.mu.Lock()
			l.isExclusive = false
			l.mu.Unlock()
			return fmt.Errorf("%w: flock on %s", ErrLockTimeout, l.path)
		}
		time.Sleep(poll)
	}
}

// RLock acquires a shared lock (`LOCK_SH`), blocking up to timeout.
func (l *FileLock) RLock() error {
	deadline := time.Now().Add(l.timeout)
	poll := l.pollInterval
	if poll <= 0 {
		poll = 10 * time.Millisecond
	}

	// Step 1: Wait until in-process shared access is available.
	for {
		l.mu.Lock()
		if !l.isExclusive {
			if l.readers > 0 && !l.readerAcquiring {
				l.readers++
				l.mu.Unlock()
				return nil
			}
			if !l.readerAcquiring && l.readers == 0 {
				l.readerAcquiring = true
				l.mu.Unlock()
				break
			}
		}
		l.mu.Unlock()
		if time.Now().After(deadline) {
			return fmt.Errorf("%w: in-process lock on %s", ErrLockTimeout, l.path)
		}
		time.Sleep(poll)
	}

	// Step 2: First reader acquires OS-level shared flock.
	dir := filepath.Dir(l.path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		l.mu.Lock()
		l.readerAcquiring = false
		l.mu.Unlock()
		return fmt.Errorf("create lock dir: %w", err)
	}

	f, err := os.OpenFile(l.path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		l.mu.Lock()
		l.readerAcquiring = false
		l.mu.Unlock()
		return fmt.Errorf("open lock file: %w", err)
	}

	for {
		err := flockShared(int(f.Fd()))
		if err == nil {
			l.mu.Lock()
			l.f = f
			l.readers = 1
			l.readerAcquiring = false
			l.mu.Unlock()
			return nil
		}
		if !isWouldBlock(err) {
			_ = f.Close()
			l.mu.Lock()
			l.readerAcquiring = false
			l.mu.Unlock()
			return fmt.Errorf("flock %s: %w", l.path, err)
		}
		if time.Now().After(deadline) {
			_ = f.Close()
			l.mu.Lock()
			l.readerAcquiring = false
			l.mu.Unlock()
			return fmt.Errorf("%w: flock on %s", ErrLockTimeout, l.path)
		}
		time.Sleep(poll)
	}
}

// Unlock releases the held lock (exclusive or shared).
func (l *FileLock) Unlock() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.isExclusive {
		f := l.f
		l.f = nil
		l.isExclusive = false
		if f == nil {
			return errors.New("cannot unlock un-acquired FileLock")
		}
		unlockErr := flockUnlock(int(f.Fd()))
		closeErr := f.Close()
		if unlockErr != nil {
			return fmt.Errorf("flock unlock: %w", unlockErr)
		}
		return closeErr
	}

	if l.readers > 0 {
		l.readers--
		if l.readers == 0 {
			f := l.f
			l.f = nil
			if f != nil {
				unlockErr := flockUnlock(int(f.Fd()))
				closeErr := f.Close()
				if unlockErr != nil {
					return fmt.Errorf("flock unlock: %w", unlockErr)
				}
				return closeErr
			}
		}
		return nil
	}

	return errors.New("cannot unlock un-acquired FileLock")
}
