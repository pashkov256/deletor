package runner_test

import (
	"os"
	"time"

	"testing"

	"github.com/pashkov256/deletor/internal/cli/config"
	"github.com/pashkov256/deletor/internal/filemanager"
	"github.com/pashkov256/deletor/internal/rules"
	"github.com/pashkov256/deletor/internal/runner"
	"github.com/stretchr/testify/assert"
)

// mockFileManager implements filemanager.FileManager interface
type mockFileManager struct {
	deletedFiles []string
	trashedFiles []string
}

func (m *mockFileManager) DeleteFile(path string) {
	m.deletedFiles = append(m.deletedFiles, path)
}

func (m *mockFileManager) MoveFileToTrash(filePath string) {
	m.trashedFiles = append(m.trashedFiles, filePath)
}

func (m *mockFileManager) NewFileFilter(minSize, maxSize int64, extensions map[string]struct{}, exclude []string, olderThan, newerThan time.Time) *filemanager.FileFilter {
	return &filemanager.FileFilter{
		MinSize:    minSize,
		MaxSize:    maxSize,
		Exclude:    exclude,
		Extensions: extensions,
		OlderThan:  olderThan,
		NewerThan:  newerThan,
	}
}

func (m *mockFileManager) WalkFilesWithFilter(callback func(fi os.FileInfo, path string), dir string, filter *filemanager.FileFilter) {
	// No operation for mock
}

func (m *mockFileManager) MoveFilesToTrash(dir string, extensions []string, exclude []string, minSize, maxSize int64, olderThan, newerThan time.Time) {
	// No operation for mock
}

func (m *mockFileManager) DeleteFiles(dir string, extensions []string, exclude []string, minSize, maxSize int64, olderThan, newerThan time.Time) {
	// No operation for mock
}

func (m *mockFileManager) DeleteEmptySubfolders(dir string) {
	// No operation for mock
}

func (m *mockFileManager) IsEmptyDir(dir string) bool {
	return true
}

func (m *mockFileManager) ExpandTilde(path string) string {
	return path
}

func (m *mockFileManager) CalculateDirSize(path string) int64 {
	return 0
}

func TestRunCLI_TrashFlag(t *testing.T) {

	tests := []struct {
		name         string
		config       *config.Config
		expectDelete bool
		expectTrash  bool
	}{
		{
			name: "Should permanently delete files when trash flag is false",
			config: &config.Config{
				Extensions:      []string{".txt"},
				SkipConfirm:     true,
				IncludeSubdirs:  true,
				MoveFileToTrash: false,
			},
			expectDelete: true,
			expectTrash:  false,
		},
		{
			name: "Should move files to trash when trash flag is true - without confirmation",
			config: &config.Config{
				Extensions:      []string{".txt"},
				SkipConfirm:     true,
				IncludeSubdirs:  true,
				MoveFileToTrash: true,
			},
			expectDelete: false,
			expectTrash:  true,
		},
		{
			name: "Should move files to trash when trash flag is true - with confirmation",
			config: &config.Config{
				Extensions:      []string{".txt"},
				SkipConfirm:     false,
				IncludeSubdirs:  true,
				MoveFileToTrash: true,
			},
			expectDelete: false,
			expectTrash:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDir, cleanup := setupTestDir(t)
			defer cleanup()

			tt.config.Directory = testDir

			if !tt.config.SkipConfirm {
				tmpFile, err := os.CreateTemp("", "input_*")
				assert.NoError(t, err)
				defer os.Remove(tmpFile.Name())

				_, err = tmpFile.WriteString("y\n")
				assert.NoError(t, err)
				tmpFile.Close()

				oldStdin := os.Stdin
				defer func() { os.Stdin = oldStdin }()

				// Opening a temporary file as stdin
				file, err := os.Open(tmpFile.Name())
				assert.NoError(t, err)
				os.Stdin = file
				defer file.Close()
			}

			// Create mock file manager
			mockFm := &mockFileManager{
				deletedFiles: make([]string, 0),
				trashedFiles: make([]string, 0),
			}

			r := rules.NewRules()

			// Run the CLI
			runner.RunCLI(mockFm, r, tt.config)

			if tt.expectDelete {
				assert.Equal(t, len(mockFm.trashedFiles), 0, "No files should be moved to trash")
				assert.Equal(t, len(mockFm.deletedFiles), 4, "Expected files to be deleted")
			}

			if tt.expectTrash {
				assert.Equal(t, len(mockFm.trashedFiles), 4, "Expected files to be moved to trash")
				assert.Equal(t, len(mockFm.deletedFiles), 0, "No files should be deleted")
			}
		})
	}
}
