package filemanager

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// recursively traverse deletion
func (f *defaultFileManager) DeleteFiles(dir string, extensions []string, exclude []string, minSize int64) {
	taskCh := make(chan FileTask, runtime.NumCPU())
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if info == nil {
			return nil
		}

		if err != nil {
			return nil
		}

		go func(path string, info os.FileInfo) {
			// Acquire token from channel first
			taskCh <- FileTask{info: info}
			defer func() { <-taskCh }() // Release token when done

			if len(exclude) != 0 {
				for _, excludePattern := range exclude {
					if strings.Contains(filepath.ToSlash(path), excludePattern+"/") ||
						strings.HasPrefix(info.Name(), excludePattern) {
						return
					}
				}
			}

			if len(extensions) > 0 {
				ext := strings.ToLower(filepath.Ext(path))
				matched := false
				for _, allowedExt := range extensions {
					if ext == allowedExt {
						matched = true
						break
					}
				}
				if !matched {
					return
				}
			}

			if info.Size() > minSize {
				os.Remove(path)
			}
		}(path, info)
		return nil
	})
}

func (f *defaultFileManager) DeleteEmptySubfolders(dir string) {
	emptyDirs := make([]string, 0)

	filepath.WalkDir(dir, func(path string, info os.DirEntry, err error) error {
		if info == nil && !info.IsDir() {
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
