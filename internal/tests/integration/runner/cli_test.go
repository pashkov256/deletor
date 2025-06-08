package runner_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/pashkov256/deletor/internal/cli/config"
	"github.com/pashkov256/deletor/internal/filemanager"
	"github.com/pashkov256/deletor/internal/rules"
	"github.com/pashkov256/deletor/internal/runner"
	"github.com/stretchr/testify/assert"
)

func setupTestDir(t *testing.T) (string, func()) {
	tempDir, err := os.MkdirTemp("", "deletor_test_*")
	assert.NoError(t, err)

	files := map[string]string{
		"test1.txt":        "content1",
		"test2.txt":        "content2",
		"test3.doc":        "content3",
		"test4.doc":        "content4",
		"test5.pdf":        "content5",
		"exclude.txt":      "exclude content",
		"subdir/test6.txt": "content6",
		"subdir/empty/":    "",
	}

	for path, content := range files {
		fullPath := filepath.Join(tempDir, path)
		if path == "subdir/empty/" {
			err := os.MkdirAll(fullPath, 0755)
			assert.NoError(t, err)
			continue
		}
		err := os.MkdirAll(filepath.Dir(fullPath), 0755)
		assert.NoError(t, err)
		err = os.WriteFile(fullPath, []byte(content), 0644)
		assert.NoError(t, err)
	}

	// Возвращаем функцию очистки
	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return tempDir, cleanup
}

func countFilesAndDirs(dir string) (int, int) {
	var fileCount, dirCount int
	fmt.Printf("\nScanning directory: %s\n", dir)
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("Error accessing path %s: %v\n", path, err)
			return nil
		}
		if info.IsDir() {
			dirCount++
			fmt.Printf("Found directory: %s\n", path)
		} else {
			fileCount++
			fmt.Printf("Found file: %s\n", path)
		}
		return nil
	})
	fmt.Printf("Total files: %d, Total directories: %d\n", fileCount, dirCount)
	return fileCount, dirCount
}

func TestRunCLI_BasicFileOperations(t *testing.T) {
	testDir, cleanup := setupTestDir(t)
	defer cleanup()

	tests := []struct {
		name          string
		config        *config.Config
		expectedFiles int
		expectedDirs  int
	}{
		{
			name: "Delete by extension",
			config: &config.Config{
				Directory:      testDir,
				Extensions:     []string{".txt"},
				SkipConfirm:    true,
				IncludeSubdirs: true,
			},
			expectedFiles: 3, // remaining .doc and .pdf files
			expectedDirs:  3,
		},
		{
			name: "Delete by size",
			config: &config.Config{
				Directory:      testDir,
				MinSize:        1,
				MaxSize:        5,
				SkipConfirm:    true,
				IncludeSubdirs: true,
			},
			expectedFiles: 3, // remaining files larger than 5 bytes
			expectedDirs:  3,
		},
		{
			name: "Delete by time",
			config: &config.Config{
				Directory:      testDir,
				OlderThan:      time.Now().Add(-time.Hour),
				SkipConfirm:    true,
				IncludeSubdirs: true,
			},
			expectedFiles: 3, // all files are newly created, but some are deleted
			expectedDirs:  3,
		},
		{
			name: "Delete with exclude",
			config: &config.Config{
				Directory:      testDir,
				Extensions:     []string{".txt"},
				Exclude:        []string{"exclude"},
				SkipConfirm:    true,
				IncludeSubdirs: true,
			},
			expectedFiles: 3, // remaining .doc, .pdf files and exclude.txt
			expectedDirs:  3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm := filemanager.NewFileManager()
			r := rules.NewRules()

			runner.RunCLI(fm, r, tt.config)

			fileCount, dirCount := countFilesAndDirs(testDir)
			assert.Equal(t, tt.expectedFiles, fileCount, "File count mismatch")
			assert.Equal(t, tt.expectedDirs, dirCount, "Directory count mismatch")
		})
	}
}

func TestRunCLI_DirectoryOperations(t *testing.T) {
	testDir, cleanup := setupTestDir(t)
	defer cleanup()

	tests := []struct {
		name          string
		config        *config.Config
		expectedFiles int
		expectedDirs  int
	}{
		{
			name: "Single directory scan",
			config: &config.Config{
				Directory:      testDir,
				Extensions:     []string{".txt"},
				SkipConfirm:    true,
				IncludeSubdirs: false,
			},
			expectedFiles: 4, // remaining .doc, .pdf files and files in subdir
			expectedDirs:  3,
		},
		{
			name: "Recursive directory scan",
			config: &config.Config{
				Directory:      testDir,
				Extensions:     []string{".txt"},
				SkipConfirm:    true,
				IncludeSubdirs: true,
			},
			expectedFiles: 3, // remaining .doc and .pdf files
			expectedDirs:  3,
		},
		{
			name: "Delete empty folders",
			config: &config.Config{
				Directory:          testDir,
				SkipConfirm:        true,
				IncludeSubdirs:     true,
				DeleteEmptyFolders: true,
			},
			expectedFiles: 0,
			expectedDirs:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm := filemanager.NewFileManager()
			r := rules.NewRules()

			runner.RunCLI(fm, r, tt.config)

			fileCount, dirCount := countFilesAndDirs(testDir)
			assert.Equal(t, tt.expectedFiles, fileCount, "File count mismatch")
			assert.Equal(t, tt.expectedDirs, dirCount, "Directory count mismatch")
		})
	}
}

func TestRunCLI_UserInteraction(t *testing.T) {
	testDir, cleanup := setupTestDir(t)
	defer cleanup()

	t.Run("Skip confirmation", func(t *testing.T) {
		config := &config.Config{
			Directory:      testDir,
			Extensions:     []string{".txt"},
			SkipConfirm:    true,
			IncludeSubdirs: true,
		}

		fm := filemanager.NewFileManager()
		r := rules.NewRules()

		runner.RunCLI(fm, r, config)

		fileCount, dirCount := countFilesAndDirs(testDir)
		assert.Equal(t, 3, fileCount, "Should have 3 files remaining (.doc and .pdf files)")
		assert.Equal(t, 3, dirCount, "Should have 3 directories remaining")
	})

	// Тест с подтверждением
	t.Run("With confirmation", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "input_*")
		assert.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		_, err = tmpFile.WriteString("y\n")
		assert.NoError(t, err)
		tmpFile.Close()

		oldStdin := os.Stdin
		defer func() { os.Stdin = oldStdin }()

		// Открываем временный файл как stdin
		file, err := os.Open(tmpFile.Name())
		assert.NoError(t, err)
		os.Stdin = file
		defer file.Close()

		config := &config.Config{
			Directory:      testDir,
			Extensions:     []string{".doc"},
			SkipConfirm:    false,
			IncludeSubdirs: true,
		}

		fm := filemanager.NewFileManager()
		r := rules.NewRules()

		runner.RunCLI(fm, r, config)

		fileCount, dirCount := countFilesAndDirs(testDir)
		assert.Equal(t, 1, fileCount, "Should have 1 file remaining (.pdf file)")
		assert.Equal(t, 3, dirCount, "Should have 3 directories remaining")
	})
}
