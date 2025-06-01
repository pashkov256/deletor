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

type FileScanner struct {
	fileManager  FileManager
	filter       *FileFilter
	ProgressChan chan int64
	haveProgress bool
	mutex        *sync.Mutex
	wg           *sync.WaitGroup
}

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

		if s.filter.MatchesFilters(info, filepath.Join(dir, entry.Name())) {
			s.mutex.Lock()

			toDeleteMap[filepath.Join(dir, entry.Name())] = utils.FormatSize(info.Size())
			totalClearSize += info.Size()

			s.mutex.Unlock()

			if s.haveProgress {
				s.ProgressChan <- info.Size()
			}
		}

	}
	return toDeleteMap, totalClearSize
}

func (s *FileScanner) ScanFilesRecursively(dir string) (toDeleteMap map[string]string, totalClearSize int64) {
	toDeleteMap = make(map[string]string)
	taskCh := make(chan os.FileInfo, runtime.NumCPU())

	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if info == nil {
			fmt.Printf("Warning: Nil FileInfo for path: %s (err: %v)\n", path, err)

			return nil
		}

		if err != nil {
			fmt.Printf("Warning: Error accessing path %s: %v\n", path, err)
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
