package fs

import (
	"os"
	"path/filepath"
	"strings"
)

// Check if a directory is empty,true if directory have subfolders
func IsEmptyDir(dirPath string) bool {
	dir, err := os.Open(dirPath)
	if err != nil {
		return false
	}
	defer dir.Close()

	entries, err := dir.Readdir(0)

	if err != nil {
		return false
	}
	if len(entries) == 0 {
		return true
	}

	for _, entry := range entries {
		if entry.IsDir() {
			// If this is a directory, we check recursively
			if !IsEmptyDir(filepath.Join(dirPath, entry.Name())) {
				return false
			}
		} else {
			return false
		}
	}
	return true
}

// Helper function to expand tilde in path
func ExpandTilde(path string) string {
	if !strings.HasPrefix(path, "~") {
		return path
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}

	return filepath.Join(home, path[1:])
}
