//go:build unix

package snippetmgr

import (
	"io/fs"
	"syscall"
)

func fileInode(info fs.FileInfo) (uint64, bool) {
	if info == nil {
		return 0, false
	}
	st, ok := info.Sys().(*syscall.Stat_t)
	if !ok || st == nil {
		return 0, false
	}
	return st.Ino, true
}
