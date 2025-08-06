package cache

import (
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/pashkov256/deletor/internal/filemanager"
)

// Manager handles cache operations for different operating systems
type Manager struct {
	Os          OS                      //made exportable for testing
	Locations   []CacheLocation         //made exportable for testing
	Filemanager filemanager.FileManager //made exportable for testing
}

// NewCacheManager creates a new cache manager instance for the current OS
func NewCacheManager(fm filemanager.FileManager) *Manager {
	return &Manager{
		Os:          OS(runtime.GOOS),
		Locations:   getLocationsForOS(OS(runtime.GOOS)),
		Filemanager: fm,
	}
}

// ScanAllLocations concurrently scans all cache locations and returns their statistics
func (m *Manager) ScanAllLocations() []ScanResult {
	var resultsScan []ScanResult

	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, location := range m.Locations {
		wg.Add(1)
		go func() {
			defer wg.Done()

			result := m.scan(location.Path)

			mu.Lock()
			resultsScan = append(resultsScan, result)
			mu.Unlock()
		}()
	}

	wg.Wait()

	return resultsScan
}

// scan analyzes a single cache location and returns its statistics
func (m *Manager) scan(path string) ScanResult {
	result := ScanResult{Path: path, FileCount: 0, Size: 0}
	filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if info == nil {
			return nil
		}

		if err != nil {
			return nil
		}

		result.Size += info.Size()
		result.FileCount++

		return nil
	})
	return result
}

// ClearCache removes all files from cache locations using OS-specific deletion methods
func (m *Manager) ClearCache() (deleteError error) {
	for _, location := range m.Locations {
		filepath.Walk(location.Path, func(path string, info os.FileInfo, err error) error {
			if info == nil {
				return nil
			}

			if err != nil {
				return nil
			}

			if !info.IsDir() {
				// Try normal deletion first
				err := os.Remove(path)
				if err != nil {
					deleteError = err

					if runtime.GOOS == "windows" {
						err := DeleteFileWithWindowsAPI(path)
						if err != nil {
							return err
						}
					}

					if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
						err := DeleteFileWithUnixAPI(path)
						if err != nil {
							return err
						}
					}
				}
				return nil
			}
			return nil
		})
	}

	return deleteError
}
func (m *Manager) GetOS() OS {
	return m.Os
}
