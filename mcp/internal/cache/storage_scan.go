package cache

import (
	"cmp"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

type cacheFileRecord struct {
	relPath   string
	key       string
	createdTS int64
	ttlTS     int64
	version   int
	size      int64
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

func listShardDirs(root *os.Root) []string {
	entries, err := fs.ReadDir(root.FS(), ".")
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

func forEachShardZstEntry(root *os.Root, shard string, visit func(relPath, fileName string, info os.FileInfo)) {
	entries, err := fs.ReadDir(root.FS(), shard)
	if err != nil {
		return
	}
	for _, d := range entries {
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".zst") {
			continue
		}
		info, statErr := d.Info()
		if statErr != nil {
			continue
		}
		visit(filepath.Join(shard, d.Name()), d.Name(), info)
	}
}

func isTmpFile(name string) bool {
	return strings.HasPrefix(name, ".tmp_")
}

func isStaleTmpFile(info os.FileInfo, now int64, force bool) bool {
	return force || (now-info.ModTime().Unix() >= staleTmpAgeSeconds)
}

func (s *Storage) initScanTotalSize() {
	now := s.config.nowUnix()
	_ = s.withLockedRoot(func(root *os.Root) {
		var total int64
		for _, shard := range listShardDirs(root) {
			forEachShardZstEntry(root, shard, func(rel, name string, info os.FileInfo) {
				if isTmpFile(name) {
					if isStaleTmpFile(info, now, false) {
						_ = root.Remove(rel)
					}
					return
				}
				total += info.Size()
			})
		}
		s.totalCompressedSize = total
	})
}

func parseCacheFileInfo(relPath, fileName string, size int64) (cacheFileRecord, bool) {
	if !strings.HasSuffix(fileName, ".zst") || isTmpFile(fileName) {
		return cacheFileRecord{}, false
	}
	stem := strings.TrimSuffix(fileName, ".zst")
	k, c, t, v, ok := parseStandardStem(stem)
	if !ok {
		return cacheFileRecord{}, false
	}
	return cacheFileRecord{
		relPath:   relPath,
		key:       k,
		createdTS: c,
		ttlTS:     t,
		version:   v,
		size:      size,
	}, true
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

func (s *Storage) findFilesByKey(root *os.Root, baseKey string) []cacheFileRecord {
	shard := baseKey[:2]
	prefix := baseKey + "_"
	var records []cacheFileRecord
	forEachShardZstEntry(root, shard, func(rel, name string, info os.FileInfo) {
		if !strings.HasPrefix(name, prefix) {
			return
		}
		if rec, ok := parseCacheFileInfo(rel, name, info.Size()); ok && rec.key == baseKey {
			records = append(records, rec)
		}
	})
	return records
}

func (s *Storage) findAndCleanFiles(root *os.Root, baseKey string, now int64) ([]cacheFileRecord, []cacheFileRecord) {
	files := s.findFilesByKey(root, baseKey)
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
