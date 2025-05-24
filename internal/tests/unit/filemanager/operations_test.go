package filemanager

import (
	"os"
	"path/filepath"
	"testing"
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

func createDirStructure(t *testing.T, root string, dirs []string, files map[string]string) {
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
			createDirStructure(t, root, tt.dirs, tt.files)
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
