package filemanager_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/pashkov256/deletor/internal/filemanager"
)

func createTestFilesWithTimes(t *testing.T, root string, files map[string]struct {
	size    int64
	modTime time.Time
}) {
	for name, data := range files {
		fullPath := filepath.Join(root, name)
		os.MkdirAll(filepath.Dir(fullPath), 0755)
		f, err := os.Create(fullPath)
		if err != nil {
			t.Fatalf("failed to create file %s: %v", name, err)
		}
		if data.size > 0 {
			if err := f.Truncate(data.size); err != nil {
				t.Fatalf("failed to set size for %s: %v", name, err)
			}
		}
		f.Close()
		os.Chtimes(fullPath, data.modTime, data.modTime)
	}
}

func TestFileFilter_FullRequirements(t *testing.T) {
	now := time.Now()
	dayAgo := now.Add(-24 * time.Hour)
	weekAgo := now.Add(-7 * 24 * time.Hour)
	monthAgo := now.Add(-30 * 24 * time.Hour)

	tests := []struct {
		name  string
		files map[string]struct {
			size    int64
			modTime time.Time
		}
		exclude     []string
		extensions  map[string]struct{}
		minSize     int64
		maxSize     int64
		olderThan   time.Time
		newerThan   time.Time
		expectMatch map[string]bool
	}{
		{
			name: "SizeFilters",
			files: map[string]struct {
				size    int64
				modTime time.Time
			}{
				"small.txt": {100, now},
				"large.txt": {1024 * 1024 * 5, now},
			},
			minSize: 1000,
			maxSize: 1024 * 1024 * 10,
			expectMatch: map[string]bool{
				"small.txt": false,
				"large.txt": true,
			},
		},
		{
			name: "ExtensionFilters",
			files: map[string]struct {
				size    int64
				modTime time.Time
			}{
				"doc.txt":    {500, now},
				"report.pdf": {600, now},
				"image.JPG":  {700, now},
			},
			extensions: map[string]struct{}{".txt": {}, ".pdf": {}, ".jpg": {}},
			expectMatch: map[string]bool{
				"doc.txt":    true,
				"report.pdf": true,
				"image.JPG":  false,
			},
		},
		{
			name: "DateFilters",
			files: map[string]struct {
				size    int64
				modTime time.Time
			}{
				"new.log": {200, dayAgo},
				"old.log": {200, monthAgo},
				"now.log": {200, now},
			},
			olderThan: now.Add(-5 * 24 * time.Hour),
			newerThan: now.Add(-31 * 24 * time.Hour),
			expectMatch: map[string]bool{
				"new.log": false,
				"old.log": true,
				"now.log": false,
			},
		},
		{
			name: "CombinedFilters",
			files: map[string]struct {
				size    int64
				modTime time.Time
			}{
				"target.txt": {2048, weekAgo},
				"skip.txt":   {50, now},
			},
			extensions: map[string]struct{}{".txt": {}},
			minSize:    1000,
			olderThan:  now.Add(-2 * 24 * time.Hour),
			expectMatch: map[string]bool{
				"target.txt": true,
				"skip.txt":   false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			createTestFilesWithTimes(t, root, tt.files)

			filter := &filemanager.FileFilter{
				MinSize:    tt.minSize,
				MaxSize:    tt.maxSize,
				Extensions: tt.extensions,
				Exclude:    tt.exclude,
				OlderThan:  tt.olderThan,
				NewerThan:  tt.newerThan,
			}

			for name := range tt.files {
				path := filepath.Join(root, name)
				info, err := os.Stat(path)
				if err != nil {
					t.Fatalf("failed to stat file: %v", err)
				}
				matched := filter.MatchesFilters(info, path)
				expected := tt.expectMatch[name]
				if matched != expected {
					t.Errorf("file %s: expected match = %v, got %v", name, expected, matched)
				}
			}
		})
	}
}
