package filemanager

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Bios-Marcel/wastebasket/v2"
	"github.com/pashkov256/deletor/internal/utils"
)

// WalkFilesWithFilter traverses files in a directory with concurrent processing
// and applies the given filter to each file
func (f *defaultFileManager) WalkFilesWithFilter(callback func(fi os.FileInfo, path string), dir string, filter *FileFilter) {
	taskCh := make(chan struct{}, runtime.NumCPU())
	var wg sync.WaitGroup

	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if info == nil {
			return nil
		}

		if err != nil {
			return nil
		}

		wg.Add(1)
		go func(path string, info os.FileInfo) {
			defer wg.Done()
			// Acquire token from channel first
			taskCh <- struct{}{}
			defer func() { <-taskCh }() // Release token when done
			if filter.MatchesFilters(info, path) {
				callback(info, path)
			}
		}(path, info)
		return nil
	})

	wg.Wait()
}

// DeleteFiles removes files matching the specified criteria from the given directory
func (f *defaultFileManager) DeleteFiles(dir string, extensions []string, exclude []string, minSize, maxSize int64, olderThan, newerThan time.Time) {
	callback := func(fi os.FileInfo, path string) {
		os.Remove(path)
	}
	fileFilter := f.NewFileFilter(minSize, maxSize, utils.ParseExtToMap(extensions), exclude, olderThan, newerThan)
	f.WalkFilesWithFilter(callback, dir, fileFilter)
}

// DeleteEmptySubfolders removes all empty directories in the given path
func (f *defaultFileManager) DeleteEmptySubfolders(dir string) {
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

// CalculateDirSize computes the total size of all files in a directory
// Uses concurrent processing with limits to handle large directories efficiently
func (f *defaultFileManager) CalculateDirSize(path string) int64 {
	// For very large directories, return a placeholder value immediately
	// to avoid blocking the UI
	_, err := os.Stat(path)
	if err != nil {
		return 0
	}

	// If it's a very large directory (like C: or Program Files)
	// just return 0 immediately to prevent lag
	if strings.HasSuffix(path, ":\\") || strings.Contains(path, "Program Files") {
		return 0
	}

	var totalSize int64 = 0

	// Use a channel to limit concurrency
	semaphore := make(chan struct{}, 10)
	var wg sync.WaitGroup

	// Create a function to process a directory
	var processDir func(string) int64
	processDir = func(dirPath string) int64 {
		var size int64 = 0
		entries, err := os.ReadDir(dirPath)
		if err != nil {
			return 0
		}

		for _, entry := range entries {
			// Skip hidden files and directories unless enabled
			if strings.HasPrefix(entry.Name(), ".") {
				continue
			}

			fullPath := filepath.Join(dirPath, entry.Name())
			if entry.IsDir() {
				// Process directories with concurrency limits
				wg.Add(1)
				go func(p string) {
					semaphore <- struct{}{}
					defer func() {
						<-semaphore
						wg.Done()
					}()
					dirSize := processDir(p)
					atomic.AddInt64(&totalSize, dirSize)
				}(fullPath)
			} else {
				// Process files directly
				info, err := entry.Info()
				if err == nil {
					fileSize := info.Size()
					atomic.AddInt64(&totalSize, fileSize)
					size += fileSize
				}
			}
		}
		return size
	}

	// Start processing
	processDir(path)

	wg.Wait()

	return totalSize
}

// MoveFilesToTrash moves files matching the criteria to the system's recycle bin
func (f *defaultFileManager) MoveFilesToTrash(dir string, extensions []string, exclude []string, minSize, maxSize int64, olderThan, newerThan time.Time) {
	callback := func(fi os.FileInfo, path string) {
		f.MoveFileToTrash(path)
	}

	fileFilter := f.NewFileFilter(minSize, maxSize, utils.ParseExtToMap(extensions), exclude, olderThan, newerThan)
	f.WalkFilesWithFilter(callback, dir, fileFilter)
}

// MoveFileToTrash moves a single file to the system's recycle bin
func (f *defaultFileManager) MoveFileToTrash(filePath string) {
	wastebasket.Trash(filePath)
}

// DeleteFile deletes a single file
func (f *defaultFileManager) DeleteFile(filePath string) {
	os.Remove(filePath)
}
