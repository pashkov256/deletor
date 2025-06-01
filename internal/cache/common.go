//go:build !windows && !linux && !darwin
// +build !windows,!linux,!darwin

package cache

// deleteFileWithWindowsAPI is a stub for non-Windows platforms
func deleteFileWithWindowsAPI(path string) error {
	return nil
}

// deleteFileWithUnixAPI is a stub for non-Unix platforms
func deleteFileWithUnixAPI(path string) error {
	return nil
}
