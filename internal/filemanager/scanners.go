package filemanager

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/pashkov256/deletor/internal/utils"
	"github.com/schollz/progressbar/v3"
)

// FileScanner handles file system scanning operations with progress tracking
type FileScanner struct {
	fileManager  FileManager // File manager instance for operations
	filter       *FileFilter // Filter criteria for files
	ProgressChan chan int64  // Channel for progress updates
	haveProgress bool        // Whether progress tracking is enabled
	mutex        *sync.Mutex
	wg           *sync.WaitGroup
}

// NewFileScanner creates a new file scanner with the specified configuration
func NewFileScanner(fileManager FileManager, filter *FileFilter, haveProgress bool) *FileScanner {
	return &FileScanner{
		fileManager:  fileManager,
		filter:       filter,
		ProgressChan: make(chan int64),
		haveProgress: haveProgress,
		mutex:        &sync.Mutex{},
		wg:           &sync.WaitGroup{},
	}
}

// ProgressBarScanner initializes and displays a progress bar for file scanning
func (s *FileScanner) ProgressBarScanner(dir string) {
	var totalScanSize int64
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if info == nil {
			return nil
		}

		if s.filter.MatchesFilters(info, path) {
			totalScanSize += info.Size()
		}

		return nil
	})

	bar := progressbar.NewOptions64(
		totalScanSize,
		progressbar.OptionSetDescription("Scanning files..."),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(10),
		progressbar.OptionThrottle(65*time.Millisecond),
		progressbar.OptionShowCount(),
		progressbar.OptionOnCompletion(func() {
			fmt.Fprint(os.Stderr, "\n")
		}),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetRenderBlankState(true))

	go func() {
		for incr := range s.ProgressChan {
			bar.Add64(incr)
		}
	}()
}

// ScanFilesCurrentLevel scans files in the current directory level only
func (s *FileScanner) ScanFilesCurrentLevel(dir string) (toDeleteMap map[string]string, totalClearSize int64) {
	toDeleteMap = make(map[string]string)
	entries, err := os.ReadDir(dir)
	if err != nil {
		panic(err)
	}

	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			panic(err)
		}

		if info.IsDir() {
			continue
		}

		if s.filter.MatchesFilters(info, filepath.Join(dir, entry.Name())) {
			toDeleteMap[filepath.Join(dir, entry.Name())] = utils.FormatSize(info.Size())
			totalClearSize += info.Size()

			if s.haveProgress {
				s.ProgressChan <- info.Size()
			}
		}

	}
	return toDeleteMap, totalClearSize
}

// ScanFilesRecursively scans files in the directory and all subdirectories
func (s *FileScanner) ScanFilesRecursively(dir string) (toDeleteMap map[string]string, totalClearSize int64) {
	toDeleteMap = make(map[string]string)
	taskCh := make(chan os.FileInfo, runtime.NumCPU())

	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if info == nil || info.IsDir() {
			return nil
		}

		if err != nil {
			return nil
		}

		s.wg.Add(1)
		go func(path string, info os.FileInfo) {
			// Acquire token from channel first
			taskCh <- info
			defer func() { <-taskCh }() // Release token when done
			defer s.wg.Done()

			if s.filter.MatchesFilters(info, path) {
				s.mutex.Lock()

				toDeleteMap[path] = utils.FormatSize(info.Size())
				totalClearSize += info.Size()

				s.mutex.Unlock()

				if s.haveProgress {
					s.ProgressChan <- info.Size()
				}
			}

		}(path, info)

		return nil
	})

	s.wg.Wait()

	return toDeleteMap, totalClearSize
}

// ScanEmptySubFolders finds all empty subdirectories in the given path
func (s *FileScanner) ScanEmptySubFolders(dir string) []string {
	emptyDirs := make([]string, 0)

	filepath.WalkDir(dir, func(path string, info os.DirEntry, err error) error {
		if info == nil && !info.IsDir() {
			return nil
		}
		if s.fileManager.IsEmptyDir(path) {
			emptyDirs = append(emptyDirs, path)
		}

		return nil
	})

	return emptyDirs
}
