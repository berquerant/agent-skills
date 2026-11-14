package cache

import (
	"cmp"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/berquerant/agent-skills/mcp/internal/store"
	"github.com/klauspost/compress/zstd"
)

const (
	defaultLockTimeout = 10 * time.Second
	staleTmpAgeSeconds = 60
)

type cacheFileRecord struct {
	relPath   string
	key       string
	createdTS int64
	ttlTS     int64
	version   int
	size      int64
}

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

func isShardDirName(name string) bool {
	if len(name) != 2 {
		return false
	}
	for i := range 2 {
		c := name[i]
		isAlphaNum := (c >= '0' && c <= '9') || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
		if !isAlphaNum && c != '_' && c != '-' {
			return false
		}
	}
	return true
}

func (s *Storage) initScanTotalSize() {
	root, err := s.openRoot()
	if err != nil {
		return
	}
	defer func() { _ = root.Close() }()

	var total int64
	for _, shard := range listShardDirs(s.config.Dir) {
		entries, err := os.ReadDir(filepath.Join(s.config.Dir, shard))
		if err != nil {
			continue
		}
		for _, d := range entries {
			total += scanShardEntry(root, shard, d)
		}
	}
	s.totalCompressedSize = total
}

func scanShardEntry(root *os.Root, shard string, d os.DirEntry) int64 {
	if d.IsDir() || !strings.HasSuffix(d.Name(), ".zst") {
		return 0
	}
	rel := filepath.Join(shard, d.Name())
	if strings.HasPrefix(d.Name(), ".tmp_") {
		_ = root.Remove(rel)
		return 0
	}
	info, err := d.Info()
	if err != nil {
		return 0
	}
	return info.Size()
}

func listShardDirs(rootDir string) []string {
	entries, err := os.ReadDir(rootDir)
	if err != nil {
		return nil
	}
	var shards []string
	for _, d := range entries {
		if d.IsDir() && isShardDirName(d.Name()) {
			shards = append(shards, d.Name())
		}
	}
	return shards
}

func parseCacheFileInfo(relPath, fileName string, size int64) (cacheFileRecord, bool) {
	if !strings.HasSuffix(fileName, ".zst") || strings.HasPrefix(fileName, ".tmp_") {
		return cacheFileRecord{}, false
	}
	stem := strings.TrimSuffix(fileName, ".zst")

	if k, c, t, v, ok := parseStandardStem(stem); ok {
		return cacheFileRecord{
			relPath:   relPath,
			key:       k,
			createdTS: c,
			ttlTS:     t,
			version:   v,
			size:      size,
		}, true
	}
	return cacheFileRecord{}, false
}

func parseStandardStem(stem string) (string, int64, int64, int, bool) {
	parts := rsplitN(stem, "_", 3)
	if len(parts) != 4 {
		return "", 0, 0, 0, false
	}
	k, cStr, tStr, vStr := parts[0], parts[1], parts[2], parts[3]
	if !isDigits(cStr) || !isDigits(tStr) || !isDigits(vStr) || !IsValidCacheKey(k) {
		return "", 0, 0, 0, false
	}
	c, err1 := strconv.ParseInt(cStr, 10, 64)
	t, err2 := strconv.ParseInt(tStr, 10, 64)
	v, err3 := strconv.Atoi(vStr)
	if err1 != nil || err2 != nil || err3 != nil || v <= 0 {
		return "", 0, 0, 0, false
	}
	return k, c, t, v, true
}

func rsplitN(s, sep string, maxSplits int) []string {
	var rightParts []string
	rem := s
	for range maxSplits {
		idx := strings.LastIndex(rem, sep)
		if idx < 0 {
			break
		}
		rightParts = append(rightParts, rem[idx+len(sep):])
		rem = rem[:idx]
	}
	out := make([]string, 0, len(rightParts)+1)
	out = append(out, rem)
	for _, rightPart := range slices.Backward(rightParts) {
		out = append(out, rightPart)
	}
	return out
}

func (s *Storage) unlinkAndDeduct(root *os.Root, relPath string, size int64) bool {
	if err := root.Remove(relPath); err != nil && !os.IsNotExist(err) {
		return false
	}
	s.totalCompressedSize = max(0, s.totalCompressedSize-size)
	return true
}

func (s *Storage) findFilesByKey(baseKey string) []cacheFileRecord {
	shard := baseKey[:2]
	shardPath := filepath.Join(s.config.Dir, shard)
	entries, err := os.ReadDir(shardPath)
	if err != nil {
		return nil
	}
	prefix := baseKey + "_"
	var records []cacheFileRecord
	for _, d := range entries {
		name := d.Name()
		if d.IsDir() || !strings.HasPrefix(name, prefix) || !strings.HasSuffix(name, ".zst") {
			continue
		}
		info, statErr := d.Info()
		if statErr != nil {
			continue
		}
		rel := filepath.Join(shard, name)
		if rec, ok := parseCacheFileInfo(rel, name, info.Size()); ok && rec.key == baseKey {
			records = append(records, rec)
		}
	}
	return records
}

func (s *Storage) findAndCleanFiles(root *os.Root, baseKey string, now int64) ([]cacheFileRecord, []cacheFileRecord) {
	files := s.findFilesByKey(baseKey)
	var valid, expired []cacheFileRecord
	for _, item := range files {
		if item.ttlTS <= now {
			s.unlinkAndDeduct(root, item.relPath, item.size)
			expired = append(expired, item)
		} else {
			valid = append(valid, item)
		}
	}
	return valid, expired
}

func findTargetFile(
	targetVer VersionTarget,
	validFiles []cacheFileRecord,
	expiredFiles []cacheFileRecord,
) (*cacheFileRecord, string) {
	if targetVer.Kind == VersionNone || targetVer.Kind == VersionLatest {
		if len(validFiles) == 0 {
			return nil, "expired"
		}
		slices.SortFunc(validFiles, func(a, b cacheFileRecord) int {
			return cmp.Compare(b.version, a.version)
		})
		picked := validFiles[0]
		return &picked, ""
	}

	for i := range validFiles {
		if validFiles[i].version == targetVer.Number {
			return &validFiles[i], ""
		}
	}
	for i := range expiredFiles {
		if expiredFiles[i].version == targetVer.Number {
			return nil, "expired"
		}
	}
	return nil, "miss"
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

	if err := s.lock.Lock(); err != nil {
		return GetResult{}, err
	}
	defer func() { _ = s.lock.Unlock() }()

	root, err := s.openRoot()
	if err != nil {
		return GetResult{}, err
	}
	defer func() { _ = root.Close() }()

	now := s.config.nowUnix()
	validFiles, expiredFiles := s.findAndCleanFiles(root, baseKey, now)
	if len(validFiles) == 0 && len(expiredFiles) == 0 {
		return GetResult{
			Hit:     false,
			Key:     FormatVersionedKey(baseKey, targetVer),
			Version: targetVer.IntPtr(),
			Reason:  "miss",
		}, nil
	}

	targetFile, reason := findTargetFile(targetVer, validFiles, expiredFiles)
	if targetFile == nil {
		return GetResult{
			Hit:     false,
			Key:     FormatVersionedKey(baseKey, targetVer),
			Version: targetVer.IntPtr(),
			Reason:  reason,
		}, nil
	}

	resolvedKey := FormatVersionedKeyNumber(baseKey, targetFile.version)
	return s.readAndDecompress(root, targetFile.relPath, resolvedKey, targetFile.version), nil
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

	if err := s.lock.Lock(); err != nil {
		return false, err
	}
	defer func() { _ = s.lock.Unlock() }()

	root, err := s.openRoot()
	if err != nil {
		return false, err
	}
	defer func() { _ = root.Close() }()

	now := s.config.nowUnix()
	validFiles, _ := s.findAndCleanFiles(root, baseKey, now)
	if targetVer.Kind == VersionNone || targetVer.Kind == VersionLatest {
		return len(validFiles) > 0, nil
	}
	for _, f := range validFiles {
		if f.version == targetVer.Number {
			return true, nil
		}
	}
	return false, nil
}

// Delete removes cache files for key (all versions, "latest", or a specific integer version).
func (s *Storage) Delete(key string, version any) (DeleteResult, error) {
	baseKey, keyVer, ok := ParseVersionedKey(key)
	if !ok {
		return DeleteResult{Hit: false, Deleted: false, Key: key}, nil
	}
	targetVer, ok := ResolveVersionTarget(version, keyVer)
	if !ok {
		return DeleteResult{Hit: false, Deleted: false, Key: key}, nil
	}

	if err := s.lock.Lock(); err != nil {
		return DeleteResult{}, err
	}
	defer func() { _ = s.lock.Unlock() }()

	root, err := s.openRoot()
	if err != nil {
		return DeleteResult{}, err
	}
	defer func() { _ = root.Close() }()

	files := s.findFilesByKey(baseKey)
	if len(files) == 0 {
		var verOut any
		if targetVer.Kind == VersionSpecific {
			verOut = targetVer.Number
		}
		return DeleteResult{
			Hit:          false,
			Deleted:      false,
			Key:          FormatVersionedKey(baseKey, targetVer),
			Version:      verOut,
			DeletedCount: 0,
		}, nil
	}

	switch targetVer.Kind {
	case VersionNone:
		return s.deleteAllVersions(root, baseKey, files), nil
	case VersionLatest:
		return s.deleteLatestVersion(root, baseKey, files), nil
	case VersionSpecific:
		return s.deleteSpecificVersion(root, baseKey, files, targetVer.Number), nil
	default:
		return DeleteResult{Hit: false, Deleted: false, Key: key}, nil
	}
}

func (s *Storage) deleteAllVersions(root *os.Root, baseKey string, files []cacheFileRecord) DeleteResult {
	deletedCount := 0
	for _, f := range files {
		if s.unlinkAndDeduct(root, f.relPath, f.size) {
			deletedCount++
		}
	}
	return DeleteResult{
		Hit:          true,
		Deleted:      deletedCount > 0,
		Key:          baseKey,
		Version:      nil,
		DeletedCount: deletedCount,
	}
}

func (s *Storage) deleteLatestVersion(root *os.Root, baseKey string, files []cacheFileRecord) DeleteResult {
	slices.SortFunc(files, func(a, b cacheFileRecord) int {
		return cmp.Compare(b.version, a.version)
	})
	latest := files[0]
	deleted := s.unlinkAndDeduct(root, latest.relPath, latest.size)
	count := 0
	if deleted {
		count = 1
	}
	return DeleteResult{
		Hit:          true,
		Deleted:      deleted,
		Key:          FormatVersionedKeyNumber(baseKey, latest.version),
		Version:      latest.version,
		DeletedCount: count,
	}
}

func (s *Storage) deleteSpecificVersion(
	root *os.Root,
	baseKey string,
	files []cacheFileRecord,
	targetVer int,
) DeleteResult {
	targetKeyStr := FormatVersionedKeyNumber(baseKey, targetVer)
	var matched []cacheFileRecord
	for _, f := range files {
		if f.version == targetVer {
			matched = append(matched, f)
		}
	}
	if len(matched) == 0 {
		return DeleteResult{
			Hit:          false,
			Deleted:      false,
			Key:          targetKeyStr,
			Version:      targetVer,
			DeletedCount: 0,
		}
	}
	deletedCount := 0
	for _, f := range matched {
		if s.unlinkAndDeduct(root, f.relPath, f.size) {
			deletedCount++
		}
	}
	return DeleteResult{
		Hit:          true,
		Deleted:      deletedCount > 0,
		Key:          targetKeyStr,
		Version:      targetVer,
		DeletedCount: deletedCount,
	}
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

	if err := s.lock.Lock(); err != nil {
		return VersionResult{}, err
	}
	defer func() { _ = s.lock.Unlock() }()

	root, err := s.openRoot()
	if err != nil {
		return VersionResult{}, err
	}
	defer func() { _ = root.Close() }()

	now := s.config.nowUnix()
	validFiles, _ := s.findAndCleanFiles(root, baseKey, now)
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
	return VersionResult{
		Hit:      len(versions) > 0,
		Key:      baseKey,
		Versions: versions,
	}, nil
}

// Purge removes expired (or all) cache files across 2-character shard directories and removes empty shards.
// Non-shard subdirectories such as locks, logs, and checkpoints are never touched.
func (s *Storage) Purge(all bool) (PurgeResult, error) {
	if err := s.lock.Lock(); err != nil {
		return PurgeResult{}, err
	}
	defer func() { _ = s.lock.Unlock() }()

	root, err := s.openRoot()
	if err != nil {
		return PurgeResult{}, err
	}
	defer func() { _ = root.Close() }()

	now := s.config.nowUnix()
	deletedCount := 0
	var freedBytes int64

	shards := listShardDirs(s.config.Dir)
	for _, shard := range shards {
		c, b := s.purgeShard(root, shard, all, now)
		deletedCount += c
		freedBytes += b
		_ = root.Remove(shard) // Removes shard directory only if empty
	}

	if all {
		s.totalCompressedSize = max(0, s.totalCompressedSize-freedBytes)
	}

	return PurgeResult{
		Success:      true,
		DeletedCount: deletedCount,
		FreedBytes:   freedBytes,
		All:          all,
	}, nil
}

func (s *Storage) purgeShard(root *os.Root, shard string, all bool, now int64) (int, int64) {
	entries, err := os.ReadDir(filepath.Join(s.config.Dir, shard))
	if err != nil {
		return 0, 0
	}
	deletedCount := 0
	var freedBytes int64

	for _, d := range entries {
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".zst") {
			continue
		}
		info, statErr := d.Info()
		if statErr != nil {
			continue
		}
		rel := filepath.Join(shard, d.Name())
		sz := info.Size()

		if strings.HasPrefix(d.Name(), ".tmp_") {
			if all || (now-info.ModTime().Unix() >= staleTmpAgeSeconds) {
				if root.Remove(rel) == nil {
					deletedCount++
					freedBytes += sz
				}
			}
			continue
		}

		if all {
			if root.Remove(rel) == nil {
				deletedCount++
				freedBytes += sz
			}
			continue
		}

		if rec, ok := parseCacheFileInfo(rel, d.Name(), sz); ok && rec.ttlTS <= now {
			if s.unlinkAndDeduct(root, rel, sz) {
				deletedCount++
				freedBytes += sz
			}
		}
	}
	return deletedCount, freedBytes
}

func (s *Storage) ensureCapacity(root *os.Root, newSize, now int64) bool {
	if s.totalCompressedSize+newSize <= s.config.MaxTotalCompressedCacheSize {
		return true
	}

	allFiles := s.evictExpiredAcrossShards(root, now)
	if s.totalCompressedSize+newSize <= s.config.MaxTotalCompressedCacheSize {
		return true
	}

	slices.SortFunc(allFiles, func(a, b cacheFileRecord) int {
		return cmp.Compare(a.createdTS, b.createdTS)
	})
	for _, f := range allFiles {
		s.unlinkAndDeduct(root, f.relPath, f.size)
		if s.totalCompressedSize+newSize <= s.config.MaxTotalCompressedCacheSize {
			break
		}
	}
	return s.totalCompressedSize+newSize <= s.config.MaxTotalCompressedCacheSize
}

func (s *Storage) evictExpiredAcrossShards(root *os.Root, now int64) []cacheFileRecord {
	var valid []cacheFileRecord
	for _, shard := range listShardDirs(s.config.Dir) {
		entries, err := os.ReadDir(filepath.Join(s.config.Dir, shard))
		if err != nil {
			continue
		}
		for _, d := range entries {
			if d.IsDir() || !strings.HasSuffix(d.Name(), ".zst") {
				continue
			}
			info, statErr := d.Info()
			if statErr != nil {
				continue
			}
			rel := filepath.Join(shard, d.Name())
			if strings.HasPrefix(d.Name(), ".tmp_") {
				if now-info.ModTime().Unix() >= staleTmpAgeSeconds {
					_ = root.Remove(rel)
				}
				continue
			}
			rec, ok := parseCacheFileInfo(rel, d.Name(), info.Size())
			if !ok {
				continue
			}
			if rec.ttlTS <= now {
				s.unlinkAndDeduct(root, rel, rec.size)
			} else {
				valid = append(valid, rec)
			}
		}
	}
	return valid
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

	if err := s.lock.Lock(); err != nil {
		return StorageResult{}, err
	}
	defer func() { _ = s.lock.Unlock() }()

	root, err := s.openRoot()
	if err != nil {
		return StorageResult{}, err
	}
	defer func() { _ = root.Close() }()

	now := s.config.nowUnix()
	allExisting := s.findFilesByKey(baseKey)
	newVersion := 1
	for _, f := range allExisting {
		if f.version >= newVersion {
			newVersion = f.version + 1
		}
	}

	s.findAndCleanFiles(root, baseKey, now)

	if !s.ensureCapacity(root, newSize, now) {
		msg := "Storage quota exceeded (max_total_compressed_cache_size)"
		return StorageResult{
			Success: false,
			Key:     &baseKey,
			Version: &newVersion,
			Message: &msg,
		}, nil
	}

	return s.writeEntryFile(baseKey, entry, newVersion, compressed, newSize), nil
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
