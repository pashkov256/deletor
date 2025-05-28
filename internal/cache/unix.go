//go:build !windows
// +build !windows

package cache

func deleteFileWithWindowsAPI(path string) error {
	// This is a stub for non-Windows platforms
	return nil
}
