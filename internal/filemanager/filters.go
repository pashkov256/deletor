package filemanager

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

type FileFilter struct {
	MinSize    int64
	MaxSize    int64
	Extensions map[string]struct{}
	Exclude    []string
	OlderThan  time.Time
	NewerThan  time.Time
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

// If we want to find files older than a certain time
func (f *FileFilter) OlderThanFilter(info os.FileInfo) bool {
	// For older than, we want files that were modified before the specified time
	return info.ModTime().Before(f.OlderThan)
}

// If we want to find files newer than a certain time
func (f *FileFilter) NewerThanFilter(info os.FileInfo) bool {
	// For newer than, we want files that were modified after the specified time
	return info.ModTime().After(f.NewerThan)
}
