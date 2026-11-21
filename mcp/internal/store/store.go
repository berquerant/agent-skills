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
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

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
