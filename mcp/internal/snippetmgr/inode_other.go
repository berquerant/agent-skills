//go:build !unix

package snippetmgr

import "io/fs"

func fileInode(_ fs.FileInfo) (uint64, bool) {
	return 0, false
}
