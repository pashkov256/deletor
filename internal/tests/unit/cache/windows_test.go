//go:build windows
// +build windows

package cache

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/pashkov256/deletor/internal/cache"
)

func TestDeleteFileWithWindowsAPI(t *testing.T) {
	tempDir := t.TempDir()
	testFileContent := []byte("lorem ipsum")

	t.Run("successful file deletion", func(t *testing.T) {
		testFilePath := filepath.Join(tempDir, "test")
		err := os.WriteFile(testFilePath, testFileContent, os.ModePerm)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		err = cache.DeleteFileWithWindowsAPI(testFilePath)
		if err != nil {
			t.Errorf("DeleteFileWithWindowsAPI failed: %v", err)
		}

		if _, err := os.Stat(testFilePath); !os.IsNotExist(err) {
			t.Errorf("File still exists after deletion")
		}
	})

	t.Run("deletion of read-only file", func(t *testing.T) {
		testFilePath := filepath.Join(tempDir, "test-read-only")
		err := os.WriteFile(testFilePath, testFileContent, os.FileMode(os.O_RDONLY))
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		filenameW, err := syscall.UTF16PtrFromString(testFilePath)
		if err != nil {
			t.Fatalf("Failed to convert test file pathname to UTF16 ptr: %v", err)
		}

		if err := syscall.SetFileAttributes(filenameW, syscall.FILE_ATTRIBUTE_READONLY); err != nil {
			t.Fatalf("Failed to set readonly attribute to test file: %v", err)
		}

		err = cache.DeleteFileWithWindowsAPI(testFilePath)
		if err != nil {
			t.Errorf("DeleteFileWithWindowsAPI failed: %v", err)
		}

		if _, err := os.Stat(testFilePath); !os.IsNotExist(err) {
			t.Errorf("File still exists after deletion")
		}
	})

	t.Run("deletion of hidden file", func(t *testing.T) {
		testFilePath := filepath.Join(tempDir, "test-hidden-file")
		err := os.WriteFile(testFilePath, testFileContent, os.ModePerm)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		filenameW, err := syscall.UTF16PtrFromString(testFilePath)
		if err != nil {
			t.Fatalf("Failed to convert test file pathname to UTF16 ptr: %v", err)
		}

		if err := syscall.SetFileAttributes(filenameW, syscall.FILE_ATTRIBUTE_HIDDEN); err != nil {
			t.Fatalf("Failed to set hidden attribute to test file: %v", err)
		}

		err = cache.DeleteFileWithWindowsAPI(testFilePath)
		if err != nil {
			t.Errorf("DeleteFileWithWindowsAPI failed: %v", err)
		}

		if _, err := os.Stat(testFilePath); !os.IsNotExist(err) {
			t.Errorf("File still exists after deletion")
		}
	})
}
