//go:build !windows && !linux && !darwin
// +build !windows,!linux,!darwin

package cache

// deleteFileWithWindowsAPI is a stub implementation for non-Windows platforms.
// Returns nil as this function is not implemented on non-Windows systems.
func deleteFileWithWindowsAPI(path string) error {
	return nil
}

// deleteFileWithUnixAPI is a stub implementation for non-Unix platforms.
// Returns nil as this function is not implemented on non-Unix systems.
func deleteFileWithUnixAPI(path string) error {
	return nil
}
