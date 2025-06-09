package filemanager_test

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/pashkov256/deletor/internal/filemanager"
	"github.com/pashkov256/deletor/internal/utils"
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

func TestScanFilesCurrentLevel(t *testing.T) {
	tests := []struct {
		name          string
		dirs          []string
		files         map[string]string
		modTimes      map[string]time.Time
		extensions    []string
		exclude       []string
		minSize       int64
		maxSize       int64
		olderThan     time.Time
		newerThan     time.Time
		expectedFiles map[string]int64
		expectedSize  int64
	}{
		{
			name:          "Empty directory",
			dirs:          []string{},
			files:         map[string]string{},
			expectedFiles: map[string]int64{},
			expectedSize:  0,
		},
		{
			name: "Single file",
			files: map[string]string{
				"test.txt": "content",
			},
			modTimes: map[string]time.Time{
				"test.txt": time.Now(),
			},
			expectedFiles: map[string]int64{
				"test.txt": 7,
			},
			expectedSize: 7,
		},
		{
			name: "Multiple files",
			files: map[string]string{
				"file1.txt": "content1",
				"file2.txt": "content2",
				"file3.txt": "content3",
			},
			modTimes: map[string]time.Time{
				"file1.txt": time.Now(),
				"file2.txt": time.Now(),
				"file3.txt": time.Now(),
			},
			expectedFiles: map[string]int64{
				"file1.txt": 8,
				"file2.txt": 8,
				"file3.txt": 8,
			},
			expectedSize: 24,
		},
		{
			name: "Files with size filter",
			files: map[string]string{
				"small.txt": "small",
				"large.txt": "this is a large file with more content",
			},
			modTimes: map[string]time.Time{
				"small.txt": time.Now(),
				"large.txt": time.Now(),
			},
			minSize: 10,
			expectedFiles: map[string]int64{
				"large.txt": 38,
			},
			expectedSize: 38,
		},
		{
			name: "Files with extension filter",
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
			expectedFiles: map[string]int64{
				"file2.pdf": 8,
			},
			expectedSize: 8,
		},
		{
			name: "Files with date filter",
			files: map[string]string{
				"old.txt": "old content",
				"new.txt": "new content",
			},
			modTimes: map[string]time.Time{
				"old.txt": time.Now().Add(-48 * time.Hour),
				"new.txt": time.Now(),
			},
			olderThan: time.Now().Add(-24 * time.Hour),
			expectedFiles: map[string]int64{
				"old.txt": 11,
			},
			expectedSize: 11,
		},
		{
			name: "Files with combined filters",
			files: map[string]string{
				"small_old.txt": "small",
				"large_old.txt": "this is a large file with more content",
				"small_new.txt": "small",
				"large_new.txt": "this is a large file with more content",
			},
			modTimes: map[string]time.Time{
				"small_old.txt": time.Now().Add(-48 * time.Hour),
				"large_old.txt": time.Now().Add(-48 * time.Hour),
				"small_new.txt": time.Now(),
				"large_new.txt": time.Now(),
			},
			minSize:   10,
			olderThan: time.Now().Add(-24 * time.Hour),
			expectedFiles: map[string]int64{
				"large_old.txt": 38,
			},
			expectedSize: 38,
		},
		{
			name: "Files with exclude filter",
			dirs: []string{"backup"},
			files: map[string]string{
				"file1.txt":        "content1",
				"backup/file2.txt": "content2",
			},
			modTimes: map[string]time.Time{
				"file1.txt":        time.Now(),
				"backup/file2.txt": time.Now(),
			},
			exclude: []string{"backup"},
			expectedFiles: map[string]int64{
				"file1.txt": 8,
			},
			expectedSize: 8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			createDirStructure(t, root, tt.dirs, tt.files, tt.modTimes)

			fm := filemanager.NewFileManager()
			filter := fm.NewFileFilter(tt.minSize, tt.maxSize, utils.ParseExtToMap(tt.extensions), tt.exclude, tt.olderThan, tt.newerThan)
			scanner := filemanager.NewFileScanner(fm, filter, false)
			files, totalSize := scanner.ScanFilesCurrentLevel(root)

			if totalSize != tt.expectedSize {
				t.Errorf("expected total size %d, got %d", tt.expectedSize, totalSize)
			}

			if len(files) != len(tt.expectedFiles) {
				t.Errorf("expected %d files, got %d", len(tt.expectedFiles), len(files))
			}

			for path, expectedSize := range tt.expectedFiles {
				fullPath := filepath.Join(root, path)
				if formattedSize, exists := files[fullPath]; !exists {
					t.Errorf("expected file %s to be found", path)
				} else {
					// Remove the space before the unit
					formattedSize = strings.ReplaceAll(formattedSize, " ", "")
					parsedSize, err := utils.ToBytes(formattedSize)
					if err != nil {
						t.Errorf("failed to parse size %s: %v", formattedSize, err)
					} else if parsedSize != expectedSize {
						t.Errorf("expected size %d for file %s, got %d", expectedSize, path, parsedSize)
					}
				}
			}
		})
	}
}

func TestScanFilesRecursively(t *testing.T) {
	tests := []struct {
		name          string
		dirs          []string
		files         map[string]string
		modTimes      map[string]time.Time
		extensions    []string
		exclude       []string
		minSize       int64
		maxSize       int64
		olderThan     time.Time
		newerThan     time.Time
		expectedFiles map[string]int64
		expectedSize  int64
	}{
		{
			name:          "Empty directory",
			dirs:          []string{},
			files:         map[string]string{},
			expectedFiles: map[string]int64{},
			expectedSize:  0,
		},
		{
			name: "Flat directory structure",
			files: map[string]string{
				"file1.txt": "abc",
				"file2.log": "defg",
				"file3.md":  "hijkl",
			},
			modTimes: map[string]time.Time{
				"file1.txt": time.Now(),
				"file2.log": time.Now(),
				"file3.md":  time.Now(),
			},
			expectedFiles: map[string]int64{
				"file1.txt": 3,
				"file2.log": 4,
				"file3.md":  5,
			},
			expectedSize: 12,
		},
		{
			name: "Deep directory structure",
			dirs: []string{"a/b/c", "a/b2", "x/y/z"},
			files: map[string]string{
				"a/b/c/file1.txt": "abc",
				"a/b2/file2.log":  "defg",
				"x/y/z/file3.md":  "hijkl",
			},
			modTimes: map[string]time.Time{
				"a/b/c/file1.txt": time.Now(),
				"a/b2/file2.log":  time.Now(),
				"x/y/z/file3.md":  time.Now(),
			},
			expectedFiles: map[string]int64{
				"a/b/c/file1.txt": 3,
				"a/b2/file2.log":  4,
				"x/y/z/file3.md":  5,
			},
			expectedSize: 12,
		},
		{
			name: "Mixed file types",
			files: map[string]string{
				"a.txt":  "1",
				"b.jpg":  "22",
				"c.pdf":  "333",
				"d.docx": "4444",
				"e.xlsx": "55555",
				"f.mp3":  "666666",
				"g.mp4":  "7777777",
			},
			modTimes: map[string]time.Time{
				"a.txt":  time.Now(),
				"b.jpg":  time.Now(),
				"c.pdf":  time.Now(),
				"d.docx": time.Now(),
				"e.xlsx": time.Now(),
				"f.mp3":  time.Now(),
				"g.mp4":  time.Now(),
			},
			expectedFiles: map[string]int64{
				"a.txt":  1,
				"b.jpg":  2,
				"c.pdf":  3,
				"d.docx": 4,
				"e.xlsx": 5,
				"f.mp3":  6,
				"g.mp4":  7,
			},
			expectedSize: 28,
		},
		{
			name: "Size filter",
			files: map[string]string{
				"small.txt": "12",
				"large.txt": "1234567890",
			},
			modTimes: map[string]time.Time{
				"small.txt": time.Now(),
				"large.txt": time.Now(),
			},
			minSize: 5,
			expectedFiles: map[string]int64{
				"large.txt": 10,
			},
			expectedSize: 10,
		},
		{
			name: "Extension filter",
			files: map[string]string{
				"a.txt": "1",
				"b.jpg": "22",
				"c.pdf": "333",
			},
			modTimes: map[string]time.Time{
				"a.txt": time.Now(),
				"b.jpg": time.Now(),
				"c.pdf": time.Now(),
			},
			extensions: []string{".pdf", ".jpg"},
			expectedFiles: map[string]int64{
				"b.jpg": 2,
				"c.pdf": 3,
			},
			expectedSize: 5,
		},
		{
			name: "Date filter",
			files: map[string]string{
				"old.txt": "old content",
				"new.txt": "new content",
			},
			modTimes: map[string]time.Time{
				"old.txt": time.Now().Add(-48 * time.Hour),
				"new.txt": time.Now(),
			},
			olderThan: time.Now().Add(-24 * time.Hour),
			expectedFiles: map[string]int64{
				"old.txt": 11,
			},
			expectedSize: 11,
		},
		{
			name: "Combined filters",
			files: map[string]string{
				"a.txt":  "1",
				"b.jpg":  "22",
				"c.pdf":  "333",
				"d.docx": "4444",
				"e.xlsx": "55555",
			},
			modTimes: map[string]time.Time{
				"a.txt":  time.Now().Add(-48 * time.Hour),
				"b.jpg":  time.Now(),
				"c.pdf":  time.Now().Add(-48 * time.Hour),
				"d.docx": time.Now().Add(-48 * time.Hour),
				"e.xlsx": time.Now().Add(-48 * time.Hour),
			},
			minSize:    3,
			extensions: []string{".pdf", ".docx", ".xlsx"},
			olderThan:  time.Now().Add(-24 * time.Hour),
			expectedFiles: map[string]int64{
				"c.pdf":  3,
				"d.docx": 4,
				"e.xlsx": 5,
			},
			expectedSize: 12,
		},
		{
			name: "No filters",
			files: map[string]string{
				"a.txt": "1",
				"b.jpg": "22",
				"c.pdf": "333",
			},
			modTimes: map[string]time.Time{
				"a.txt": time.Now(),
				"b.jpg": time.Now(),
				"c.pdf": time.Now(),
			},
			expectedFiles: map[string]int64{
				"a.txt": 1,
				"b.jpg": 2,
				"c.pdf": 3,
			},
			expectedSize: 6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			createDirStructure(t, root, tt.dirs, tt.files, tt.modTimes)

			fm := filemanager.NewFileManager()
			filter := fm.NewFileFilter(tt.minSize, tt.maxSize, utils.ParseExtToMap(tt.extensions), tt.exclude, tt.olderThan, tt.newerThan)
			scanner := filemanager.NewFileScanner(fm, filter, false)
			files, totalSize := scanner.ScanFilesRecursively(root)

			if totalSize != tt.expectedSize {
				t.Errorf("expected total size %d, got %d", tt.expectedSize, totalSize)
			}

			if len(files) != len(tt.expectedFiles) {
				t.Errorf("expected %d files, got %d", len(tt.expectedFiles), len(files))
			}

			for path, expectedSize := range tt.expectedFiles {
				fullPath := filepath.Join(root, path)
				if formattedSize, exists := files[fullPath]; !exists {
					t.Errorf("expected file %s to be found", path)
				} else {
					formattedSize = strings.ReplaceAll(formattedSize, " ", "")
					parsedSize, err := utils.ToBytes(formattedSize)
					if err != nil {
						t.Errorf("failed to parse size %s: %v", formattedSize, err)
					} else if parsedSize != expectedSize {
						t.Errorf("expected size %d for file %s, got %d", expectedSize, path, parsedSize)
					}
				}
			}
		})
	}
}

func TestScanEmptySubFolders(t *testing.T) {
	tests := []struct {
		name             string
		dirs             []string
		files            map[string]string
		expectedEmpty    []string
		expectedNonEmpty []string
	}{
		{
			name: "No empty folders",
			dirs: []string{"dir1", "dir2"},
			files: map[string]string{
				"dir1/file1.txt": "content1",
				"dir2/file2.txt": "content2",
			},
			expectedEmpty:    []string{},
			expectedNonEmpty: []string{"dir1", "dir2"},
		},
		{
			name:             "Single empty folder",
			dirs:             []string{"empty"},
			files:            map[string]string{},
			expectedEmpty:    []string{".", "empty"},
			expectedNonEmpty: []string{},
		},
		{
			name:             "Multiple empty folders",
			dirs:             []string{"empty1", "empty2", "empty3"},
			files:            map[string]string{},
			expectedEmpty:    []string{".", "empty1", "empty2", "empty3"},
			expectedNonEmpty: []string{},
		},
		{
			name:             "Nested empty folders",
			dirs:             []string{"parent/child1", "parent/child2/grandchild"},
			files:            map[string]string{},
			expectedEmpty:    []string{".", "parent", "parent/child1", "parent/child2", "parent/child2/grandchild"},
			expectedNonEmpty: []string{},
		},
		{
			name: "Mixed empty and non-empty folders",
			dirs: []string{"empty1", "empty2", "nonempty1", "nonempty2"},
			files: map[string]string{
				"nonempty1/file1.txt": "content1",
				"nonempty2/file2.txt": "content2",
			},
			expectedEmpty:    []string{"empty1", "empty2"},
			expectedNonEmpty: []string{"nonempty1", "nonempty2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			createDirStructure(t, root, tt.dirs, tt.files, nil)

			fm := filemanager.NewFileManager()
			filter := fm.NewFileFilter(0, 0, nil, nil, time.Time{}, time.Time{})
			scanner := filemanager.NewFileScanner(fm, filter, false)
			emptyDirs := scanner.ScanEmptySubFolders(root)

			// Convert expected paths to full paths
			expectedEmptyFull := make([]string, len(tt.expectedEmpty))
			for i, path := range tt.expectedEmpty {
				if path == "." {
					expectedEmptyFull[i] = root
				} else {
					expectedEmptyFull[i] = filepath.Join(root, path)
				}
			}

			// Sort both slices to ensure consistent comparison
			sort.Strings(emptyDirs)
			sort.Strings(expectedEmptyFull)

			if !reflect.DeepEqual(emptyDirs, expectedEmptyFull) {
				t.Errorf("expected empty directories %v, got %v", expectedEmptyFull, emptyDirs)
			}

			// Verify non-empty directories still exist
			for _, path := range tt.expectedNonEmpty {
				fullPath := filepath.Join(root, path)
				if _, err := os.Stat(fullPath); os.IsNotExist(err) {
					t.Errorf("expected non-empty directory to exist: %s", path)
				}
			}
		})
	}
}
