//go:build linux || darwin
// +build linux darwin

package cache

import (
	"golang.org/x/sys/unix"
)

// deleteFileWithUnixAPI deletes a file using Unix system calls
func deleteFileWithUnixAPI(path string) error {
	var stat unix.Stat_t
	err := unix.Stat(path, &stat)
	if err != nil {
		return err
	}

	if stat.Mode&0744 != 0 {
		err := unix.Chmod(path, 0744)
		if err != nil {
			return err
		}
	}

	return unix.Unlink(path)
}

// deleteFileWithWindowsAPI is a stub for Unix platforms
func deleteFileWithWindowsAPI(path string) error {
	return nil
}
