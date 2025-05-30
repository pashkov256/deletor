//go:build linux || darwin
// +build linux darwin

package cache

import (
	"golang.org/x/sys/unix"
)

func deleteFileWithLinuxAPI(path string) error {
	// This is a stub for non-Windows platforms

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

	return unix.Rmdir(path)

}
