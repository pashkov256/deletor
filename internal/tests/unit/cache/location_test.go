package cache_test

import (
	"os"
	"path/filepath"
	"runtime"
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

	// Verify locations
	if expectedLocations == nil {
		if cm.ScanAllLocations() != nil {
			t.Error("Expected nil locations for unsupported OS")
		}
	} else {
		if len(expectedLocations) != len(cm.ScanAllLocations()) {
			t.Error("Wrong number of locations initialized")
		}
		for i, expected := range expectedLocations {
			if expected.Path != cm.ScanAllLocations()[i].Path {
				t.Errorf("Path mismatch for location %d", i)
			}
		}
	}
}
