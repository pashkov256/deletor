package cache_test

import (
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"testing"

	"github.com/pashkov256/deletor/internal/cache"
	"github.com/pashkov256/deletor/internal/filemanager"
)

func TestNewCacheManager_InitializesCorrectLocations(t *testing.T) {
	fm := filemanager.NewFileManager()
	cm := cache.NewCacheManager(fm)
	expectedOS := cache.OS(runtime.GOOS)
	if expectedOS != cm.GetOS() {
		t.Errorf("expectedOS  %s: functionResult  match = %s", expectedOS, cm.GetOS())
	}
	// Get expected locations
	var expectedLocations []cache.CacheLocation
	switch expectedOS {
	case cache.Windows:
		localAppData := os.Getenv("LOCALAPPDATA")
		if len(localAppData) == 0 {
			t.Error("LOCALAPPDATA environment variable should be set on Windows")
		}
		expectedLocations = []cache.CacheLocation{
			{Path: filepath.Join(localAppData, "Temp")},
			{Path: filepath.Join(localAppData, "Microsoft", "Windows", "Explorer")},
		}
	case cache.Linux:
		homeDir, err := os.UserHomeDir()
		if err != nil {
			t.Error(err)
		}
		expectedLocations = []cache.CacheLocation{
			{Path: "/tmp"},
			{Path: "/var/tmp"},
			{Path: filepath.Join(homeDir, ".cache")},
		}
	default:
		expectedLocations = nil
	}

	actualLocations := cm.ScanAllLocations()
	if len(expectedLocations) != len(actualLocations) {
		t.Error("Wrong number of locations initialized")
	}

	// Sort both slices by path to ensure consistent comparison
	sort.Slice(expectedLocations, func(i, j int) bool {
		return expectedLocations[i].Path < expectedLocations[j].Path
	})
	sort.Slice(actualLocations, func(i, j int) bool {
		return actualLocations[i].Path < actualLocations[j].Path
	})

	for i, expected := range expectedLocations {
		if expected.Path != actualLocations[i].Path {
			t.Errorf("Path mismatch for location %d", i)
		}
	}
}
