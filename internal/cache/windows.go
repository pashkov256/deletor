//go:build windows
// +build windows

package cache

import (
	"golang.org/x/sys/windows"
)

// DeleteFileWithWindowsAPI deletes a file using Windows API calls.
// Handles read-only files by removing the read-only attribute before deletion.
// Returns error if file cannot be deleted or if path conversion fails.
func DeleteFileWithWindowsAPI(path string) error {
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

// DeleteFileWithUnixAPI is a stub implementation for Windows platforms.
// Returns nil as this function is not implemented on Windows.
func DeleteFileWithUnixAPI(path string) error {
	return nil
}
