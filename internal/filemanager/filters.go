package filemanager

import (
	"os"
	"path/filepath"
	"strings"
)

type FileFilter struct {
	MinSize    int64
	MaxSize    int64
	Extensions map[string]struct{}
	Exclude    []string
}

func (d *defaultFileManager) NewFileFilter(minSize, maxSize int64, extensions map[string]struct{}, exclude []string) *FileFilter {
	return &FileFilter{
		MinSize:    minSize,
		MaxSize:    maxSize,
		Exclude:    exclude,
		Extensions: extensions,
	}
}

func (f *FileFilter) MatchesFilters(info os.FileInfo, path string) bool {
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
