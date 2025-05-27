package filemanager_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/pashkov256/deletor/internal/filemanager"
)

type testFileManager struct{}

func (f *testFileManager) IsEmptyDir(path string) bool {
	entries, err := os.ReadDir(path)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if entry.IsDir() {
			if !f.IsEmptyDir(filepath.Join(path, entry.Name())) {
				return false
			}
		} else {
			return false
		}
	}
	return true
}

func (f *testFileManager) DeleteEmptySubfolders(dir string) {
	emptyDirs := make([]string, 0)

	filepath.WalkDir(dir, func(path string, info os.DirEntry, err error) error {
		if info == nil || !info.IsDir() {
			return nil
		}

		if f.IsEmptyDir(path) {
			emptyDirs = append(emptyDirs, path)
		}

		return nil
	})

	for i := len(emptyDirs) - 1; i >= 0; i-- {
		os.Remove(emptyDirs[i])
	}
}

func createDirStructure(t *testing.T, root string, dirs []string, files map[string]string, modTimes map[string]time.Time) {
	for _, dir := range dirs {
		path := filepath.Join(root, dir)
		err := os.MkdirAll(path, 0755)
		if err != nil {
			t.Fatalf("failed to create dir: %v", err)
		}
	}
	for path, content := range files {
		fullPath := filepath.Join(root, path)
		os.MkdirAll(filepath.Dir(fullPath), 0755)
		err := os.WriteFile(fullPath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("failed to create file: %v", err)
		}
		if modTime, ok := modTimes[path]; ok {
			err = os.Chtimes(fullPath, modTime, modTime)
			if err != nil {
				t.Fatalf("failed to set file modification time: %v", err)
			}
		}
	}
}

func verifyFileContents(t *testing.T, path string, expectedContent string) {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Errorf("failed to read file %s: %v", path, err)
		return
	}
	if string(content) != expectedContent {
		t.Errorf("file %s has unexpected content. got: %s, want: %s", path, string(content), expectedContent)
	}
}

func TestDeleteEmptySubfolders(t *testing.T) {
	tests := []struct {
		name      string
		dirs      []string
		files     map[string]string
		remaining []string
	}{
		{
			name:      "Single empty directory",
			dirs:      []string{"empty"},
			remaining: []string{},
		},
		{
			name:      "Chain of empty directories",
			dirs:      []string{"folder1/folder2/folder3"},
			remaining: []string{},
		},
		{
			name:      "Multiple empty directories at same level",
			dirs:      []string{"root/folder1/empty_subfolder", "root/folder2/empty_subfolder"},
			remaining: []string{},
		},
		{
			name:      "Directory with file remains",
			dirs:      []string{"withfile"},
			files:     map[string]string{"withfile/file.txt": "data"},
			remaining: []string{"withfile"},
		},
		{
			name:      "Directory with non-empty subdir remains",
			dirs:      []string{"parent/child"},
			files:     map[string]string{"parent/child/file.txt": "data"},
			remaining: []string{"parent", "parent/child"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			createDirStructure(t, root, tt.dirs, tt.files, nil)
			fm := &testFileManager{}
			fm.DeleteEmptySubfolders(root)

			for _, dir := range tt.remaining {
				path := filepath.Join(root, dir)
				if _, err := os.Stat(path); os.IsNotExist(err) {
					t.Errorf("expected directory to remain: %s", dir)
				}
			}
		})
	}
}

func TestDeleteFiles(t *testing.T) {
	tests := []struct {
		name           string
		dirs           []string
		files          map[string]string
		modTimes       map[string]time.Time
		extensions     []string
		exclude        []string
		minSize        int64
		maxSize        int64
		olderThan      time.Time
		newerThan      time.Time
		shouldExist    []string
		shouldNotExist []string
	}{
		{
			name: "Delete files in root directory",
			dirs: []string{"subdir1", "subdir2"},
			files: map[string]string{
				"file1.txt":         "content1",
				"file2.txt":         "content2",
				"subdir1/file3.txt": "content3",
				"subdir2/file4.txt": "content4",
			},
			modTimes: map[string]time.Time{
				"file1.txt":         time.Now(),
				"file2.txt":         time.Now(),
				"subdir1/file3.txt": time.Now(),
				"subdir2/file4.txt": time.Now(),
			},
			extensions: []string{".txt"},
			shouldNotExist: []string{
				"file1.txt",
				"file2.txt",
				"subdir1/file3.txt",
				"subdir2/file4.txt",
			},
			shouldExist: []string{
				"subdir1",
				"subdir2",
			},
		},
		{
			name: "Delete files with size filter",
			files: map[string]string{
				"small.txt": "small",
				"large.txt": "this is a large file with more content",
			},
			modTimes: map[string]time.Time{
				"small.txt": time.Now(),
				"large.txt": time.Now(),
			},
			extensions: []string{".txt"},
			minSize:    10,
			shouldNotExist: []string{
				"large.txt",
			},
			shouldExist: []string{
				"small.txt",
			},
		},
		{
			name: "Delete files with extension filter",
			files: map[string]string{
				"file1.txt": "content1",
				"file2.pdf": "content2",
				"file3.txt": "content3",
			},
			modTimes: map[string]time.Time{
				"file1.txt": time.Now(),
				"file2.pdf": time.Now(),
				"file3.txt": time.Now(),
			},
			extensions: []string{".pdf"},
			shouldNotExist: []string{
				"file2.pdf",
			},
			shouldExist: []string{
				"file1.txt",
				"file3.txt",
			},
		},
		{
			name: "Delete files with exclude filter",
			dirs: []string{"backup"},
			files: map[string]string{
				"file1.txt":        "content1",
				"backup/file2.txt": "content2",
			},
			modTimes: map[string]time.Time{
				"file1.txt":        time.Now(),
				"backup/file2.txt": time.Now(),
			},
			extensions: []string{".txt"},
			exclude:    []string{"backup"},
			shouldNotExist: []string{
				"file1.txt",
			},
			shouldExist: []string{
				"backup",
				"backup/file2.txt",
			},
		},
		{
			name: "Delete files with time filter",
			files: map[string]string{
				"old.txt": "old content",
				"new.txt": "new content",
			},
			modTimes: map[string]time.Time{
				"old.txt": time.Now().Add(-48 * time.Hour),
				"new.txt": time.Now().Add(-12 * time.Hour),
			},
			extensions: []string{".txt"},
			olderThan:  time.Now().Add(-24 * time.Hour),
			shouldNotExist: []string{
				"old.txt",
			},
			shouldExist: []string{
				"new.txt",
			},
		},
		{
			name: "Delete files with multiple extensions",
			files: map[string]string{
				"file1.txt": "content1",
				"file2.pdf": "content2",
				"file3.doc": "content3",
			},
			modTimes: map[string]time.Time{
				"file1.txt": time.Now(),
				"file2.pdf": time.Now(),
				"file3.doc": time.Now(),
			},
			extensions: []string{".txt", ".pdf"},
			shouldNotExist: []string{
				"file1.txt",
				"file2.pdf",
			},
			shouldExist: []string{
				"file3.doc",
			},
		},
		{
			name: "Delete files with multiple exclude patterns",
			dirs: []string{"backup", "temp"},
			files: map[string]string{
				"file1.txt":        "content1",
				"backup/file2.txt": "content2",
				"temp/file3.txt":   "content3",
			},
			modTimes: map[string]time.Time{
				"file1.txt":        time.Now(),
				"backup/file2.txt": time.Now(),
				"temp/file3.txt":   time.Now(),
			},
			extensions: []string{".txt"},
			exclude:    []string{"backup", "temp"},
			shouldNotExist: []string{
				"file1.txt",
			},
			shouldExist: []string{
				"backup",
				"backup/file2.txt",
				"temp",
				"temp/file3.txt",
			},
		},
		{
			name: "Delete files with both min and max size",
			files: map[string]string{
				"tiny.txt":  "t",
				"small.txt": "small content",
				"large.txt": "this is a very large file with lots of content that should be deleted",
			},
			modTimes: map[string]time.Time{
				"tiny.txt":  time.Now(),
				"large.txt": time.Now(),
			},
			extensions: []string{".txt"},
			minSize:    10,
			maxSize:    50,
			shouldNotExist: []string{
				"small.txt",
			},
			shouldExist: []string{
				"tiny.txt",
				"large.txt",
			},
		},
		{
			name: "Delete files with both time filters",
			files: map[string]string{
				"old.txt":     "old content",
				"new.txt":     "new content",
				"current.txt": "current content",
			},
			modTimes: map[string]time.Time{
				"old.txt":     time.Now().Add(-72 * time.Hour),
				"new.txt":     time.Now().Add(-12 * time.Hour),
				"current.txt": time.Now().Add(-36 * time.Hour),
			},
			extensions: []string{".txt"},
			olderThan:  time.Now().Add(-48 * time.Hour),
			newerThan:  time.Now().Add(-24 * time.Hour),
			shouldNotExist: []string{
				"current.txt",
			},
			shouldExist: []string{
				"old.txt",
				"new.txt",
			},
		},
		{
			name:       "Delete files with empty directory",
			dirs:       []string{"empty"},
			files:      map[string]string{},
			modTimes:   map[string]time.Time{},
			extensions: []string{".txt"},
			shouldExist: []string{
				"empty",
			},
		},
		{
			name: "Delete files with no matching extensions",
			files: map[string]string{
				"file1.txt": "content1",
				"file2.pdf": "content2",
			},
			extensions: []string{".doc"},
			shouldExist: []string{
				"file1.txt",
				"file2.pdf",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory for test
			root := t.TempDir()

			// Create directory structure and files
			createDirStructure(t, root, tt.dirs, tt.files, tt.modTimes)

			// Verify initial file contents
			for path, content := range tt.files {
				fullPath := filepath.Join(root, path)
				verifyFileContents(t, fullPath, content)
			}

			// Create file manager instance
			fm := filemanager.NewFileManager()

			// Execute DeleteFiles
			fm.DeleteFiles(root, tt.extensions, tt.exclude, tt.minSize, tt.maxSize, tt.olderThan, tt.newerThan)

			// Verify files that should not exist
			for _, path := range tt.shouldNotExist {
				fullPath := filepath.Join(root, path)
				if _, err := os.Stat(fullPath); !os.IsNotExist(err) {
					t.Errorf("file should have been deleted: %s", path)
				}
			}

			// Verify files that should exist and their contents
			for _, path := range tt.shouldExist {
				fullPath := filepath.Join(root, path)
				if _, err := os.Stat(fullPath); os.IsNotExist(err) {
					t.Errorf("file should not have been deleted: %s", path)
				}
				// If the file should exist and we have its expected content, verify it
				if content, ok := tt.files[path]; ok {
					verifyFileContents(t, fullPath, content)
				}
			}
		})
	}
}
