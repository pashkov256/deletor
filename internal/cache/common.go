//go:build !windows && !linux && !darwin
// +build !windows,!linux,!darwin

package cache

// DeleteFileWithWindowsAPI is a stub implementation for non-Windows platforms.
// Returns nil as this function is not implemented on non-Windows systems.
func DeleteFileWithWindowsAPI(path string) error {
	return nil
}

// DeleteFileWithUnixAPI is a stub implementation for non-Unix platforms.
// Returns nil as this function is not implemented on non-Unix systems.
func DeleteFileWithUnixAPI(path string) error {
	return nil
}
