package cache

import (
	"cmp"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/berquerant/agent-skills/mcp/internal/store"
	"github.com/klauspost/compress/zstd"
)

const (
	defaultLockTimeout = 10 * time.Second
	staleTmpAgeSeconds = 60
)

// Storage manages zstd-compressed immutable cache files on disk.
type Storage struct {
	config              Config
	lock                *store.FileLock
	totalCompressedSize int64
	enc                 *zstd.Encoder
	dec                 *zstd.Decoder
}

// NewStorage creates a Storage instance and initializes the in-memory compressed size ledger.
func NewStorage(cfg Config) (*Storage, error) {
	if err := os.MkdirAll(cfg.Dir, 0o700); err != nil {
		return nil, fmt.Errorf("create cache dir: %w", err)
	}
	_ = os.Chmod(cfg.Dir, 0o700)

	enc, err := zstd.NewWriter(nil, zstd.WithEncoderLevel(zstd.SpeedDefault))
	if err != nil {
		return nil, fmt.Errorf("create zstd encoder: %w", err)
	}
	// Cap zstd decoder memory consumption to prevent decompression bomb attacks (zip bombs)
	// from consuming unbounded host memory.
	maxDecodeMem := uint64(max(cfg.MaxCacheSize*2, 16*1024*1024)) //nolint:gosec // G115: max() guarantees positive int64
	dec, err := zstd.NewReader(nil, zstd.WithDecoderMaxMemory(maxDecodeMem))
	if err != nil {
		_ = enc.Close()
		return nil, fmt.Errorf("create zstd decoder: %w", err)
	}

	s := &Storage{
		config: cfg,
		lock:   store.NewFileLock(filepath.Join(cfg.Dir, "locks", "lockfile"), defaultLockTimeout),
		enc:    enc,
		dec:    dec,
	}
	s.initScanTotalSize()
	return s, nil
}

// Close releases encoder and decoder resources.
func (s *Storage) Close() error {
	if s.dec != nil {
		s.dec.Close()
	}
	if s.enc != nil {
		return s.enc.Close()
	}
	return nil
}

// TotalCompressedSize returns the tracked total compressed size in bytes.
func (s *Storage) TotalCompressedSize() int64 {
	return s.totalCompressedSize
}

func (s *Storage) openRoot() (*os.Root, error) {
	return os.OpenRoot(s.config.Dir)
}

func (s *Storage) withLockedRoot(fn func(root *os.Root)) error {
	if err := s.lock.Lock(); err != nil {
		return err
	}
	defer func() { _ = s.lock.Unlock() }()

	root, err := s.openRoot()
	if err != nil {
		return err
	}
	defer func() { _ = root.Close() }()

	fn(root)
	return nil
}

func (s *Storage) unlinkAndDeduct(root *os.Root, relPath string, size int64) bool {
	if err := root.Remove(relPath); err != nil && !os.IsNotExist(err) {
		return false
	}
	s.totalCompressedSize = max(0, s.totalCompressedSize-size)
	return true
}

func (s *Storage) readAndDecompress(root *os.Root, relPath, resolvedKey string, resolvedVer int) GetResult {
	v := resolvedVer
	compressed, err := root.ReadFile(relPath)
	if err != nil {
		return GetResult{
			Hit:     false,
			Key:     resolvedKey,
			Version: &v,
			Reason:  fmt.Sprintf("io_error: %v", err),
		}
	}
	decompressed, err := s.dec.DecodeAll(compressed, nil)
	if err != nil {
		return GetResult{
			Hit:     false,
			Key:     resolvedKey,
			Version: &v,
			Reason:  fmt.Sprintf("corrupt: %v", err),
		}
	}
	var entry CacheEntry
	if err := json.Unmarshal(decompressed, &entry); err != nil {
		return GetResult{
			Hit:     false,
			Key:     resolvedKey,
			Version: &v,
			Reason:  fmt.Sprintf("corrupt: %v", err),
		}
	}
	entry.Normalize()
	return GetResult{
		Hit:     true,
		Key:     resolvedKey,
		Version: &v,
		Data:    &entry,
	}
}

// Get retrieves and decompresses a cached record by key and optional version.
func (s *Storage) Get(key string, version any) (GetResult, error) {
	baseKey, keyVer, ok := ParseVersionedKey(key)
	if !ok {
		return GetResult{Hit: false, Key: key, Reason: "invalid_key"}, nil
	}
	targetVer, ok := ResolveVersionTarget(version, keyVer)
	if !ok {
		return GetResult{Hit: false, Key: key, Reason: "invalid_version"}, nil
	}

	var out GetResult
	err := s.withLockedRoot(func(root *os.Root) {
		now := s.config.nowUnix()
		validFiles, expiredFiles := s.findAndCleanFiles(root, baseKey, now)
		if len(validFiles) == 0 && len(expiredFiles) == 0 {
			out = GetResult{
				Hit:     false,
				Key:     FormatVersionedKey(baseKey, targetVer),
				Version: targetVer.IntPtr(),
				Reason:  "miss",
			}
			return
		}
		targetFile, reason := findTargetFile(targetVer, validFiles, expiredFiles)
		if targetFile == nil {
			out = GetResult{
				Hit:     false,
				Key:     FormatVersionedKey(baseKey, targetVer),
				Version: targetVer.IntPtr(),
				Reason:  reason,
			}
			return
		}
		resolvedKey := FormatVersionedKeyNumber(baseKey, targetFile.version)
		out = s.readAndDecompress(root, targetFile.relPath, resolvedKey, targetFile.version)
	})
	return out, err
}

// Exist checks whether a non-expired cache entry exists for key and version without decompressing payload.
func (s *Storage) Exist(key string, version any) (bool, error) {
	baseKey, keyVer, ok := ParseVersionedKey(key)
	if !ok {
		return false, nil
	}
	targetVer, ok := ResolveVersionTarget(version, keyVer)
	if !ok {
		return false, nil
	}

	var found bool
	err := s.withLockedRoot(func(root *os.Root) {
		validFiles, _ := s.findAndCleanFiles(root, baseKey, s.config.nowUnix())
		if targetVer.Kind == VersionNone || targetVer.Kind == VersionLatest {
			found = len(validFiles) > 0
			return
		}
		for _, f := range validFiles {
			if f.version == targetVer.Number {
				found = true
				return
			}
		}
	})
	return found, err
}

// Version lists all valid, non-expired versions for key sorted ascending by version.
func (s *Storage) Version(key string) (VersionResult, error) {
	baseKey, _, ok := ParseVersionedKey(key)
	if !ok {
		return VersionResult{
			Hit:      false,
			Key:      key,
			Versions: []VersionInfo{},
			Reason:   "invalid_key",
		}, nil
	}

	var out VersionResult
	err := s.withLockedRoot(func(root *os.Root) {
		validFiles, _ := s.findAndCleanFiles(root, baseKey, s.config.nowUnix())
		slices.SortFunc(validFiles, func(a, b cacheFileRecord) int {
			return cmp.Compare(a.version, b.version)
		})

		versions := make([]VersionInfo, 0, len(validFiles))
		for _, f := range validFiles {
			versions = append(versions, VersionInfo{
				Version:          f.version,
				Key:              FormatVersionedKeyNumber(baseKey, f.version),
				CreatedTimestamp: f.createdTS,
				TTLTimestamp:     f.ttlTS,
				Expired:          false,
			})
		}
		out = VersionResult{
			Hit:      len(versions) > 0,
			Key:      baseKey,
			Versions: versions,
		}
	})
	return out, err
}

// Set compresses and writes a new immutable version of entry for key.
func (s *Storage) Set(key string, entry CacheEntry) (StorageResult, error) {
	baseKey, _, ok := ParseVersionedKey(key)
	if !ok {
		msg := fmt.Sprintf("Invalid cache key format: %s", key)
		return StorageResult{Success: false, Key: &key, Message: &msg}, nil
	}

	entry.Normalize()
	rawJSON, err := json.Marshal(entry)
	if err != nil {
		msg := fmt.Sprintf("Failed to serialize entry: %v", err)
		return StorageResult{Success: false, Key: &key, Message: &msg}, nil
	}

	if int64(len(rawJSON)) > s.config.MaxCacheSize {
		msg := fmt.Sprintf("Exceeds max_cache_size (%d > %d)", len(rawJSON), s.config.MaxCacheSize)
		return StorageResult{Success: false, Key: &key, Message: &msg}, nil
	}

	compressed := s.enc.EncodeAll(rawJSON, nil)
	newSize := int64(len(compressed))

	var out StorageResult
	err = s.withLockedRoot(func(root *os.Root) {
		now := s.config.nowUnix()
		newVersion := 1
		for _, f := range s.findFilesByKey(baseKey) {
			if f.version >= newVersion {
				newVersion = f.version + 1
			}
		}

		s.findAndCleanFiles(root, baseKey, now)

		if !s.ensureCapacity(root, newSize, now) {
			msg := "Storage quota exceeded (max_total_compressed_cache_size)"
			out = StorageResult{
				Success: false,
				Key:     &baseKey,
				Version: &newVersion,
				Message: &msg,
			}
			return
		}
		out = s.writeEntryFile(baseKey, entry, newVersion, compressed, newSize)
	})
	return out, err
}

func (s *Storage) writeEntryFile(
	baseKey string,
	entry CacheEntry,
	newVersion int,
	compressed []byte,
	newSize int64,
) StorageResult {
	subDir := filepath.Join(s.config.Dir, baseKey[:2])
	if err := os.MkdirAll(subDir, 0o700); err != nil {
		msg := fmt.Sprintf("Failed to write file: %v", err)
		return StorageResult{Success: false, Key: &baseKey, Version: &newVersion, Message: &msg}
	}
	_ = os.Chmod(subDir, 0o700)

	fileName := fmt.Sprintf("%s_%d_%d_%d.zst", baseKey, entry.CreatedTimestamp, entry.TTLTimestamp, newVersion)
	targetPath := filepath.Join(subDir, fileName)
	if err := store.WriteAtomic(targetPath, compressed, 0o600); err != nil {
		msg := fmt.Sprintf("Failed to write file: %v", err)
		return StorageResult{Success: false, Key: &baseKey, Version: &newVersion, Message: &msg}
	}

	s.totalCompressedSize += newSize
	versionedKey := FormatVersionedKeyNumber(baseKey, newVersion)
	return StorageResult{
		Success: true,
		Key:     &versionedKey,
		Version: &newVersion,
	}
}
