//go:build windows
// +build windows

package cache

import (
	"golang.org/x/sys/windows"
)

func deleteFileWithWindowsAPI(path string) error {
	// Convert path to Windows path format
	pathPtr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return err
	}

	// Try to get file attributes
	attrs, err := windows.GetFileAttributes(pathPtr)
	if err != nil {
		return err
	}

	// Remove read-only attribute if present
	if attrs&windows.FILE_ATTRIBUTE_READONLY != 0 {
		windows.SetFileAttributes(pathPtr, attrs&^windows.FILE_ATTRIBUTE_READONLY)
	}

	// Try to delete with Windows API
	return windows.DeleteFile(pathPtr)
}
