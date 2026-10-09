//go:build !windows

package storage

import "os"

func rename(oldPath, newPath string) error {
	return os.Rename(oldPath, newPath)
}
