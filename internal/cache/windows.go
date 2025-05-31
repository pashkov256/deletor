//go:build windows
// +build windows

package cache

import (
	"golang.org/x/sys/windows"
)

// deleteFileWithWindowsAPI deletes a file using Windows API calls
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
		err := windows.SetFileAttributes(pathPtr, attrs&^windows.FILE_ATTRIBUTE_READONLY)
		if err != nil {
			return err
		}
	}

	// Try to delete with Windows API
	return windows.DeleteFile(pathPtr)
}

// deleteFileWithUnixAPI is a stub for Windows platforms
func deleteFileWithUnixAPI(path string) error {
	return nil
}
