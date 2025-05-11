package filemanager

import (
	"os"
	"path/filepath"
	"strings"
)

type FileFilter struct {
	MinSize    int64
	Extensions map[string]bool
	Exclude    []string
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

	return info.Size() > f.MinSize && f.Extensions[filepath.Ext(info.Name())]
}
