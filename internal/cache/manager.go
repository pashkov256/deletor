package cache

import (
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/pashkov256/deletor/internal/filemanager"
	"golang.org/x/sys/windows"
)

type Manager struct {
	os          OS
	locations   []CacheLocation
	filemanager filemanager.FileManager
}

func NewCacheManager(fm filemanager.FileManager) *Manager {
	return &Manager{
		os:          OS(runtime.GOOS),
		locations:   getLocationsForOS(OS(runtime.GOOS)),
		filemanager: fm,
	}
}

func (m *Manager) ScanAllLocations() []ScanResult {
	var resultsScan []ScanResult

	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, location := range m.locations {
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

func (m *Manager) ClearCache() {
	for _, location := range m.locations {
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
					if runtime.GOOS == "windows" {
						// Convert path to Windows path format
						pathPtr, err := windows.UTF16PtrFromString(path)
						if err != nil {
							return nil
						}

						// Try to get file attributes
						attrs, err := windows.GetFileAttributes(pathPtr)
						if err != nil {
							return nil
						}

						// Remove read-only attribute if present
						if attrs&windows.FILE_ATTRIBUTE_READONLY != 0 {
							windows.SetFileAttributes(pathPtr, attrs&^windows.FILE_ATTRIBUTE_READONLY)
						}

						// Try to delete with Windows API
						windows.DeleteFile(pathPtr)
					}
				}
				return nil
			}

			return nil
		})
	}
}
