package filemanager

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

// FileFilter defines criteria for filtering files
type FileFilter struct {
	MinSize    int64               // Minimum file size in bytes
	MaxSize    int64               // Maximum file size in bytes
	Extensions map[string]struct{} // Set of allowed file extensions
	Exclude    []string            // Patterns to exclude from results
	OlderThan  time.Time           // Only include files older than this time
	NewerThan  time.Time           // Only include files newer than this time
}

func (d *defaultFileManager) NewFileFilter(minSize, maxSize int64, extensions map[string]struct{}, exclude []string, olderThan, newerThan time.Time) *FileFilter {
	return &FileFilter{
		MinSize:    minSize,
		MaxSize:    maxSize,
		Exclude:    exclude,
		Extensions: extensions,
		OlderThan:  olderThan,
		NewerThan:  newerThan,
	}
}

// MatchesFilters checks if a file matches all filter criteria
func (f *FileFilter) MatchesFilters(info os.FileInfo, path string) bool {
	if !f.ExcludeFilter(info, path) {
		return false
	}

	if len(f.Extensions) > 0 {
		_, existExt := f.Extensions[filepath.Ext(info.Name())]
		if !existExt {
			return false
		}
	}
	if f.MaxSize > 0 {
		if !(info.Size() <= f.MaxSize) {
			return false
		}
	}
	if f.MinSize > 0 {
		if !(info.Size() >= f.MinSize) {
			return false
		}
	}

	modTime := info.ModTime()
	if !f.OlderThan.IsZero() && !f.NewerThan.IsZero() {
		// Support 'between' range regardless of which is earlier
		start := f.OlderThan
		end := f.NewerThan
		if end.Before(start) {
			start, end = end, start
		}
		if !(modTime.After(start) && modTime.Before(end)) {
			return false
		}
	} else {
		if !f.OlderThan.IsZero() {
			if !f.OlderThanFilter(info) {
				return false
			}
		}
		if !f.NewerThan.IsZero() {
			if !f.NewerThanFilter(info) {
				return false
			}
		}
	}

	return true
}

// ExcludeFilter checks if a file should be excluded based on path patterns
func (f *FileFilter) ExcludeFilter(info os.FileInfo, path string) bool {
	if len(f.Exclude) != 0 {
		for _, excludePattern := range f.Exclude {
			if strings.Contains(filepath.ToSlash(path), excludePattern+"/") {
				return false
			}
			if strings.HasPrefix(info.Name(), excludePattern) {
				return false
			}
		}
	}
	return true
}

// OlderThanFilter checks if a file is older than the specified time
func (f *FileFilter) OlderThanFilter(info os.FileInfo) bool {
	return info.ModTime().Before(f.OlderThan)
}

// NewerThanFilter checks if a file is newer than the specified time
func (f *FileFilter) NewerThanFilter(info os.FileInfo) bool {
	return info.ModTime().After(f.NewerThan)
}
