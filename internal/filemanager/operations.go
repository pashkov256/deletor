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

func (f *defaultFileManager) WalkFilesWithFilter(callback func(fi os.FileInfo, path string), dir string, extensions []string, exclude []string, minSize, maxSize int64, olderThan, newerThan time.Time) {
	filter := f.NewFileFilter(minSize, maxSize, utils.ParseExtToMap(extensions), exclude, olderThan, newerThan)
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

			if filter.MatchesFilters(info, path) {
				callback(info, path)
			}

		}(path, info)
		return nil
	})
}

// recursively traverse deletion
func (f *defaultFileManager) DeleteFiles(dir string, extensions []string, exclude []string, minSize, maxSize int64, olderThan, newerThan time.Time) {
	callback := func(fi os.FileInfo, path string) {
		os.Remove(path)
	}

	f.WalkFilesWithFilter(callback, dir, extensions, exclude, minSize, maxSize, olderThan, newerThan)
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

// Function to calculate directory size recursively with option to cancel
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

// Recursively move file to recycle bin
func (f *defaultFileManager) MoveFilesToTrash(dir string, extensions []string, exclude []string, minSize, maxSize int64, olderThan, newerThan time.Time) {
	callback := func(fi os.FileInfo, path string) {
		f.MoveFileToTrash(path)
	}
	f.WalkFilesWithFilter(callback, dir, extensions, exclude, minSize, maxSize, olderThan, newerThan)
}

// Recursively move files to recycle bin by path
func (f *defaultFileManager) MoveFileToTrash(filePath string) {
	wastebasket.Trash(filePath)
}
