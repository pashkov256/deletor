package cache

import (
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"sort"
	"testing"

	"github.com/pashkov256/deletor/internal/cache"
	"github.com/pashkov256/deletor/internal/filemanager"
)

func TestScanAllLocations(t *testing.T) {
	fm := filemanager.NewFileManager()
	Os := cache.OS(runtime.GOOS)
	t.Run("successful scanning of empty directory", func(t *testing.T) {
		tempDir1 := t.TempDir()
		tempDir2 := t.TempDir()
		locs := []cache.CacheLocation{
			{Path: tempDir1, Type: "system"},
			{Path: tempDir2, Type: "system"},
		}
		testManager := &cache.Manager{
			Os:          Os,
			Locations:   locs,
			Filemanager: fm,
		}
		results := testManager.ScanAllLocations()
		infoTemp1, statErr1 := os.Stat(tempDir1)
		if statErr1 != nil {
			t.Fatal(statErr1)
		}
		infoTemp2, statErr2 := os.Stat(tempDir2)
		if statErr2 != nil {
			t.Fatal(statErr2)
		}
		expected := []cache.ScanResult{
			{FileCount: 1, Path: tempDir1, Size: infoTemp1.Size(), Error: nil}, //directories are being counted in FileCount as 1
			{FileCount: 1, Path: tempDir2, Size: infoTemp2.Size(), Error: nil}, //directories are being counted in FileCount as 1
		}
		sort.Slice(results, func(i, j int) bool {
			return results[i].Path < results[j].Path
		})
		sort.Slice(expected, func(i, j int) bool {
			return expected[i].Path < expected[j].Path
		})
		if len(results) != len(locs) {
			t.Fatalf("Expected %v scan results, got %v", len(locs), len(results))
		}
		for i, res := range results {
			if res.FileCount != expected[i].FileCount {
				t.Errorf("Expected %v files in scan result %v, got %v", expected[i].FileCount, i, res.FileCount)
			}
			if res.Path != expected[i].Path {
				t.Errorf("Expected file path %v in scan result %v, got %v", expected[i].Path, i, res.Path)
			}
			if res.Size != expected[i].Size {
				t.Errorf("Expected %v size in scan result %v, got %v", expected[i].Size, i, res.Size)
			}
			if res.Error != nil {
				t.Errorf("Error in scan result %v: %v", i, res.Error)
			}
		}
	})
	t.Run("successful scanning of nonempty directories", func(t *testing.T) {
		tempDirA := t.TempDir()
		tempDirB := t.TempDir()
		file1 := filepath.Join(tempDirA, "testfile1.txt")
		file2 := filepath.Join(tempDirB, "testfile2.txt")
		file3 := filepath.Join(tempDirB, "testfile3.txt")
		content1 := []byte("testfile1")
		content2 := []byte("testfile2")
		content3 := []byte("testfile3")
		os.WriteFile(file1, []byte(content1), 0644)
		os.WriteFile(file2, []byte(content2), 0644)
		os.WriteFile(file3, []byte(content3), 0644)
		locs := []cache.CacheLocation{
			{Path: tempDirA, Type: "system"},
			{Path: tempDirB, Type: "system"},
		}
		testManager := &cache.Manager{
			Os:          Os,
			Locations:   locs,
			Filemanager: fm,
		}
		results := testManager.ScanAllLocations()
		infoTempA, statErr1 := os.Stat(tempDirA)
		if statErr1 != nil {
			t.Fatal(statErr1)
		}
		infoTempB, statErr2 := os.Stat(tempDirB)
		if statErr2 != nil {
			t.Fatal(statErr2)
		}
		expected := []cache.ScanResult{
			{FileCount: 2, Path: tempDirA, Size: (infoTempA.Size() + int64(len(content1))), Error: nil},               //directories are being counted in FileCount as 1
			{FileCount: 3, Path: tempDirB, Size: (infoTempB.Size() + int64(len(content2)+len(content3))), Error: nil}, //directories are being counted in FileCount as 1
		}
		sort.Slice(results, func(i, j int) bool {
			return results[i].Path < results[j].Path
		})
		sort.Slice(expected, func(i, j int) bool {
			return expected[i].Path < expected[j].Path
		})
		if len(results) != len(locs) {
			t.Fatalf("Expected %v scan results, got %v", len(locs), len(results))
		}
		for i, res := range results {
			if res.FileCount != expected[i].FileCount {
				t.Errorf("Expected %v files in scan result %v, got %v", expected[i].FileCount, i, res.FileCount)
			}
			if res.Path != expected[i].Path {
				t.Errorf("Expected %v filepath in scan result %v, got %v", expected[i].Path, i, res.Path)
			}
			if res.Size != expected[i].Size {
				t.Errorf("Expected %v size in scan result %v, got %v", expected[i].Size, i, res.Size)
			}
			if res.Error != nil {
				t.Errorf("Error in result %v: %v", i, res.Error)
			}
		}
	})
	t.Run("successful cross-platform scanning", func(t *testing.T) {
		m := cache.NewCacheManager(fm)
		results := m.ScanAllLocations()
		var expectedPaths = []string{}
		switch runtime.GOOS {
		case "windows":
			expectedPaths = []string{
				filepath.Join(os.Getenv("LOCALAPPDATA"), "Temp"),
				filepath.Join(os.Getenv("LOCALAPPDATA"), "Microsoft", "Windows", "Explorer"),
			}
		case "linux":
			home, _ := os.UserHomeDir()
			expectedPaths = []string{
				"/tmp",
				"/var/tmp",
				filepath.Join(home, ".cache"),
			}
		default:
			t.Skip("Unsupported OS")
		}
		sort.Slice(results, func(i, j int) bool {
			return results[i].Path < results[j].Path
		})
		slices.Sort(expectedPaths)
		if len(results) != len(expectedPaths) {
			t.Fatalf("Expected %d results, got %d", len(expectedPaths), len(results))
		}
		for i, result := range results {
			if result.Path != expectedPaths[i] {
				t.Errorf("Expected path %q but got %q", expectedPaths[i], result.Path)
			}
		}
	})
}
