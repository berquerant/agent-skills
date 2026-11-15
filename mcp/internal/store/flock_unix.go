//go:build unix

package store

import (
	"errors"
	"syscall"
)

func flockExclusive(fd int) error {
	return syscall.Flock(fd, syscall.LOCK_EX|syscall.LOCK_NB)
}

func flockShared(fd int) error {
	return syscall.Flock(fd, syscall.LOCK_SH|syscall.LOCK_NB)
}

func flockUnlock(fd int) error {
	return syscall.Flock(fd, syscall.LOCK_UN)
}

func isWouldBlock(err error) bool {
	return errors.Is(err, syscall.EWOULDBLOCK) || errors.Is(err, syscall.EAGAIN)
}
