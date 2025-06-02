package filemanager

import (
	"os"
	"path/filepath"
	"strings"
)

// IsEmptyDir checks if a directory is empty, including recursive check of subdirectories.
// Returns true if the directory is empty or contains only empty subdirectories.
func (f *defaultFileManager) IsEmptyDir(dirPath string) bool {
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
			if !f.IsEmptyDir(filepath.Join(dirPath, entry.Name())) {
				return false
			}
		} else {
			return false
		}
	}
	return true
}

// ExpandTilde expands the tilde (~) in a path to the user's home directory.
// Returns the original path if it doesn't start with tilde or if home directory cannot be determined.
func (f *defaultFileManager) ExpandTilde(path string) string {
	if !strings.HasPrefix(path, "~") {
		return path
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}

	return filepath.Join(home, path[1:])
}
