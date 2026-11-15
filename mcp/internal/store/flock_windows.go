//go:build windows

package store

func flockExclusive(fd int) error {
	return nil
}

func flockShared(fd int) error {
	return nil
}

func flockUnlock(fd int) error {
	return nil
}

func isWouldBlock(err error) bool {
	return false
}
