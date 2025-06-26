//go:build linux || darwin
// +build linux darwin

package cache

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/pashkov256/deletor/internal/cache"
)

func TestDeleteFileWithUnixAPI(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("successful deletion of regular file", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "test_file.txt")
		err := os.WriteFile(testFile, []byte("test content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		err = cache.DeleteFileWithUnixAPI(testFile)
		if err != nil {
			t.Errorf("DeleteFileWithUnixAPI failed: %v", err)
		}

		if _, err := os.Stat(testFile); !os.IsNotExist(err) {
			t.Errorf("File still exists after deletion")
		}
	})

	t.Run("file with incorrect permissions gets chmod before deletion", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "restricted_file.txt")
		err := os.WriteFile(testFile, []byte("restricted content"), 0600)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		err = cache.DeleteFileWithUnixAPI(testFile)
		if err != nil {
			t.Errorf("DeleteFileWithUnixAPI failed: %v", err)
		}

		if _, err := os.Stat(testFile); !os.IsNotExist(err) {
			t.Errorf("File still exists after deletion")
		}
	})

	t.Run("file with correct permissions 0744", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "correct_perms.txt")
		err := os.WriteFile(testFile, []byte("content"), 0744)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		err = cache.DeleteFileWithUnixAPI(testFile)
		if err != nil {
			t.Errorf("DeleteFileWithUnixAPI failed: %v", err)
		}

		if _, err := os.Stat(testFile); !os.IsNotExist(err) {
			t.Errorf("File still exists after deletion")
		}
	})

	t.Run("nonexistent file returns error", func(t *testing.T) {
		nonexistentFile := filepath.Join(tempDir, "nonexistent.txt")

		err := cache.DeleteFileWithUnixAPI(nonexistentFile)
		if err == nil {
			t.Errorf("Expected error for nonexistent file, got nil")
		}
	})

	t.Run("directory instead of file returns error", func(t *testing.T) {
		testDir := filepath.Join(tempDir, "test_directory")
		err := os.Mkdir(testDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}

		err = cache.DeleteFileWithUnixAPI(testDir)
		if err == nil {
			t.Errorf("Expected error when trying to delete directory, got nil")
		}

		os.RemoveAll(testDir)
	})

	t.Run("read-only file gets proper permissions before deletion", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "readonly.txt")
		err := os.WriteFile(testFile, []byte("readonly content"), 0444)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		err = cache.DeleteFileWithUnixAPI(testFile)
		if err != nil {
			t.Errorf("DeleteFileWithUnixAPI failed: %v", err)
		}

		if _, err := os.Stat(testFile); !os.IsNotExist(err) {
			t.Errorf("File still exists after deletion")
		}
	})
}

func TestDeleteFileWithWindowsAPI(t *testing.T) {
	t.Run("stub function returns nil", func(t *testing.T) {
		err := cache.DeleteFileWithWindowsAPI("/any/path")
		if err != nil {
			t.Errorf("Expected nil from stub function, got: %v", err)
		}
	})
}

func BenchmarkDeleteFileWithUnixAPI(b *testing.B) {
	tempDir := b.TempDir()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testFile := filepath.Join(tempDir, fmt.Sprintf("bench_file_%d.txt", i))
		err := os.WriteFile(testFile, []byte("benchmark content"), 0644)
		if err != nil {
			b.Fatalf("Failed to create test file: %v", err)
		}

		err = cache.DeleteFileWithUnixAPI(testFile)
		if err != nil {
			b.Fatalf("DeleteFileWithUnixAPI failed: %v", err)
		}
	}
}

func TestDeleteFileWithUnixAPI_Symlink(t *testing.T) {
	tempDir := t.TempDir()

	originalFile := filepath.Join(tempDir, "original.txt")
	err := os.WriteFile(originalFile, []byte("original content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create original file: %v", err)
	}

	symlinkFile := filepath.Join(tempDir, "symlink.txt")
	err = os.Symlink(originalFile, symlinkFile)
	if err != nil {
		t.Fatalf("Failed to create symlink: %v", err)
	}

	err = cache.DeleteFileWithUnixAPI(symlinkFile)
	if err != nil {
		t.Errorf("DeleteFileWithUnixAPI failed on symlink: %v", err)
	}

	if _, err := os.Lstat(symlinkFile); !os.IsNotExist(err) {
		t.Errorf("Symlink still exists after deletion")
	}

	if _, err := os.Stat(originalFile); os.IsNotExist(err) {
		t.Errorf("Original file was deleted when removing symlink")
	}
}
