package store

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	defaultLockTimeout = 10 * time.Second
	defaultLockPoll    = 10 * time.Millisecond
)

// ErrLockTimeout is returned when acquiring a file lock exceeds the configured timeout.
var ErrLockTimeout = errors.New("acquire file lock timed out")

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
		timeout = defaultLockTimeout
	}
	return &FileLock{
		path:         lockPath,
		timeout:      timeout,
		pollInterval: defaultLockPoll,
	}
}

func (l *FileLock) effectivePoll() time.Duration {
	if l.pollInterval <= 0 {
		return defaultLockPoll
	}
	return l.pollInterval
}

// Lock acquires an exclusive lock (`LOCK_EX`), blocking up to timeout.
func (l *FileLock) Lock() error {
	deadline := time.Now().Add(l.timeout)
	poll := l.effectivePoll()

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
	f, err := l.acquireOSFlock(deadline, poll, flockExclusive)
	if err != nil {
		l.mu.Lock()
		l.isExclusive = false
		l.mu.Unlock()
		return err
	}
	l.mu.Lock()
	l.f = f
	l.mu.Unlock()
	return nil
}

// RLock acquires a shared lock (`LOCK_SH`), blocking up to timeout.
func (l *FileLock) RLock() error {
	deadline := time.Now().Add(l.timeout)
	poll := l.effectivePoll()

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
	f, err := l.acquireOSFlock(deadline, poll, flockShared)
	if err != nil {
		l.mu.Lock()
		l.readerAcquiring = false
		l.mu.Unlock()
		return err
	}
	l.mu.Lock()
	l.f = f
	l.readers = 1
	l.readerAcquiring = false
	l.mu.Unlock()
	return nil
}

func (l *FileLock) acquireOSFlock(
	deadline time.Time,
	poll time.Duration,
	flockFn func(fd int) error,
) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(l.path), 0o700); err != nil {
		return nil, fmt.Errorf("create lock dir: %w", err)
	}
	f, err := os.OpenFile(l.path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, fmt.Errorf("open lock file: %w", err)
	}

	for {
		err := flockFn(int(f.Fd()))
		if err == nil {
			return f, nil
		}
		if !isWouldBlock(err) {
			_ = f.Close()
			return nil, fmt.Errorf("flock %s: %w", l.path, err)
		}
		if time.Now().After(deadline) {
			_ = f.Close()
			return nil, fmt.Errorf("%w: flock on %s", ErrLockTimeout, l.path)
		}
		time.Sleep(poll)
	}
}

func releaseOSFlock(f *os.File) error {
	unlockErr := flockUnlock(int(f.Fd()))
	closeErr := f.Close()
	if unlockErr != nil {
		return fmt.Errorf("flock unlock: %w", unlockErr)
	}
	return closeErr
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
		return releaseOSFlock(f)
	}

	if l.readers > 0 {
		l.readers--
		if l.readers == 0 {
			f := l.f
			l.f = nil
			if f != nil {
				return releaseOSFlock(f)
			}
		}
		return nil
	}

	return errors.New("cannot unlock un-acquired FileLock")
}

// LockPool manages a thread-safe map of per-key FileLocks rooted at baseDir.
// Callers must pass validated, relative keys (without traversal segments).
type LockPool struct {
	baseDir string
	timeout time.Duration
	mu      sync.Mutex
	locks   map[string]*FileLock
}

// NewLockPool creates a LockPool that stores `<key>.lock` files under baseDir.
func NewLockPool(baseDir string, timeout time.Duration) *LockPool {
	if timeout <= 0 {
		timeout = defaultLockTimeout
	}
	return &LockPool{
		baseDir: baseDir,
		timeout: timeout,
		locks:   make(map[string]*FileLock),
	}
}

// LockFor returns the shared in-process FileLock instance for key.
func (p *LockPool) LockFor(key string) *FileLock {
	p.mu.Lock()
	defer p.mu.Unlock()
	if l, ok := p.locks[key]; ok {
		return l
	}
	l := NewFileLock(filepath.Join(p.baseDir, key+".lock"), p.timeout)
	p.locks[key] = l
	return l
}

// WithLock executes fn while holding an exclusive FileLock for key.
func (p *LockPool) WithLock(key string, fn func() error) error {
	lock := p.LockFor(key)
	if err := lock.Lock(); err != nil {
		return err
	}
	defer func() { _ = lock.Unlock() }()
	return fn()
}

// DeleteLock removes the lock entry and underlying .lock file for key.
func (p *LockPool) DeleteLock(key string) {
	p.mu.Lock()
	delete(p.locks, key)
	p.mu.Unlock()
	_ = os.Remove(filepath.Join(p.baseDir, key+".lock"))
}

// WithRLock executes fn while holding a shared FileLock for key.
func (p *LockPool) WithRLock(key string, fn func() error) error {
	lock := p.LockFor(key)
	if err := lock.RLock(); err != nil {
		return err
	}
	defer func() { _ = lock.Unlock() }()
	return fn()
}
