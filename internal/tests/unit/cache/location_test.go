package cache_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/pashkov256/deletor/internal/cache"
	"github.com/pashkov256/deletor/internal/filemanager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCacheManager_InitializesCorrectLocations(t *testing.T) {
	fm := filemanager.NewFileManager()
	cm := cache.NewCacheManager(fm)
	expectedOS := cache.OS(runtime.GOOS)
	assert.Equal(t, string(expectedOS), cm.GetOS(), "Cache manager initialized with wrong OS")
	// Get expected locations
	var expectedLocations []cache.CacheLocation
	switch expectedOS {
	case cache.Windows:
		localAppData := os.Getenv("LOCALAPPDATA")
		require.NotEmpty(t, localAppData, "LOCALAPPDATA environment variable should be set on Windows")
		expectedLocations = []cache.CacheLocation{
			{Path: filepath.Join(localAppData, "Temp")},
			{Path: filepath.Join(localAppData, "Microsoft", "Windows", "Explorer")},
		}
	case cache.Linux:
		homeDir, err := os.UserHomeDir()
		require.NoError(t, err)
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
		assert.Nil(t, cm.ScanAllLocations(), "Expected nil locations for unsupported OS")
	} else {
		require.Equal(t, len(expectedLocations), len(cm.ScanAllLocations()), "Wrong number of locations initialized")
		for i, expected := range expectedLocations {
			assert.Equal(t, expected.Path, cm.ScanAllLocations()[i].Path, "Path mismatch for location %d", i)
		}
	}
}
