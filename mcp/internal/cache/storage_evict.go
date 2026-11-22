package cache

import (
	"cmp"
	"os"
	"slices"
)

// Delete removes cache files for key (all versions, "latest", or a specific integer version).
func (s *Storage) Delete(key string, version *int) (DeleteResult, error) {
	baseKey, keyVer, ok := ParseVersionedKey(key)
	if !ok {
		return DeleteResult{Hit: false, Deleted: false, Key: key}, nil
	}
	targetVer, ok := ResolveVersionTarget(version, keyVer)
	if !ok {
		return DeleteResult{Hit: false, Deleted: false, Key: key}, nil
	}

	var out DeleteResult
	err := s.withLockedRoot(func(root *os.Root) {
		files := s.findFilesByKey(root, baseKey)
		if len(files) == 0 {
			var verOut *int
			if targetVer.Kind == VersionSpecific {
				verOut = &targetVer.Number
			}
			out = DeleteResult{
				Hit:          false,
				Deleted:      false,
				Key:          FormatVersionedKey(baseKey, targetVer),
				Version:      verOut,
				DeletedCount: 0,
			}
			return
		}

		switch targetVer.Kind {
		case VersionNone:
			out = s.deleteAllVersions(root, baseKey, files)
		case VersionLatest:
			out = s.deleteLatestVersion(root, baseKey, files)
		case VersionSpecific:
			out = s.deleteSpecificVersion(root, baseKey, files, targetVer.Number)
		default:
			out = DeleteResult{Hit: false, Deleted: false, Key: key}
		}
	})
	return out, err
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
		Version:      &latest.version,
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
			Version:      &targetVer,
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
		Version:      &targetVer,
		DeletedCount: deletedCount,
	}
}

// Purge removes expired (or all) cache files across 2-character shard directories and removes empty shards.
// Non-shard subdirectories such as locks, logs, and checkpoints are never touched.
func (s *Storage) Purge(all bool) (PurgeResult, error) {
	var out PurgeResult
	err := s.withLockedRoot(func(root *os.Root) {
		now := s.config.nowUnix()
		deletedCount := 0
		var freedBytes int64
		var trackedFreedBytes int64

		for _, shard := range listShardDirs(root) {
			c, b, tb := s.purgeShard(root, shard, all, now)
			deletedCount += c
			freedBytes += b
			trackedFreedBytes += tb
			_ = root.Remove(shard) // Removes shard directory only if empty
		}

		if all {
			s.totalCompressedSize = max(0, s.totalCompressedSize-trackedFreedBytes)
		}

		out = PurgeResult{
			Success:      true,
			DeletedCount: deletedCount,
			FreedBytes:   freedBytes,
			All:          all,
		}
	})
	return out, err
}

func (s *Storage) purgeShard(root *os.Root, shard string, all bool, now int64) (int, int64, int64) {
	deletedCount := 0
	var freedBytes int64
	var trackedFreedBytes int64

	forEachShardZstEntry(root, shard, func(rel, name string, info os.FileInfo) {
		sz := info.Size()
		if isTmpFile(name) {
			if isStaleTmpFile(info, now, all) && root.Remove(rel) == nil {
				deletedCount++
				freedBytes += sz
			}
			return
		}
		if all {
			if root.Remove(rel) == nil {
				deletedCount++
				freedBytes += sz
				trackedFreedBytes += sz
			}
			return
		}
		if rec, ok := parseCacheFileInfo(rel, name, sz); ok && rec.ttlTS <= now {
			if s.unlinkAndDeduct(root, rel, sz) {
				deletedCount++
				freedBytes += sz
			}
		}
	})
	return deletedCount, freedBytes, trackedFreedBytes
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
	var currentTotalSize int64
	for _, shard := range listShardDirs(root) {
		forEachShardZstEntry(root, shard, func(rel, name string, info os.FileInfo) {
			if isTmpFile(name) {
				if isStaleTmpFile(info, now, false) {
					_ = root.Remove(rel)
				}
				return
			}
			rec, ok := parseCacheFileInfo(rel, name, info.Size())
			if !ok {
				return
			}
			if rec.ttlTS <= now {
				_ = root.Remove(rel)
			} else {
				valid = append(valid, rec)
				currentTotalSize += rec.size
			}
		})
	}
	s.totalCompressedSize = currentTotalSize
	return valid
}
