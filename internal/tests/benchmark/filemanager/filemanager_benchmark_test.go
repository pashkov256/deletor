package filemanager_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/pashkov256/deletor/internal/filemanager"
	"github.com/pashkov256/deletor/internal/utils"
)

// setupBenchmarkDir creates a temporary directory structure for benchmarking
// Returns the root directory path and a cleanup function
func setupBenchmarkDir(b *testing.B, fileCount, subDirCount int, fileSize int) string {
	b.Helper()

	root := b.TempDir()
	content := make([]byte, fileSize)
	for i := range content {
		content[i] = byte('a' + (i % 26))
	}

	// Create files in root
	for i := 0; i < fileCount; i++ {
		ext := []string{".txt", ".log", ".pdf", ".doc", ".jpg"}[i%5]
		filename := filepath.Join(root, "file"+string(rune('0'+i%10))+ext)
		if err := os.WriteFile(filename, content, 0644); err != nil {
			b.Fatalf("failed to create file: %v", err)
		}
	}

	// Create subdirectories with files
	for d := 0; d < subDirCount; d++ {
		subDir := filepath.Join(root, "subdir"+string(rune('0'+d%10)))
		if err := os.MkdirAll(subDir, 0755); err != nil {
			b.Fatalf("failed to create subdir: %v", err)
		}
		for i := 0; i < fileCount/2; i++ {
			ext := []string{".txt", ".log", ".pdf", ".doc", ".jpg"}[i%5]
			filename := filepath.Join(subDir, "file"+string(rune('0'+i%10))+ext)
			if err := os.WriteFile(filename, content, 0644); err != nil {
				b.Fatalf("failed to create file: %v", err)
			}
		}
	}

	return root
}

// setupDeepBenchmarkDir creates a deeply nested directory structure for benchmarking
func setupDeepBenchmarkDir(b *testing.B, depth, filesPerLevel int, fileSize int) string {
	b.Helper()

	root := b.TempDir()
	content := make([]byte, fileSize)
	for i := range content {
		content[i] = byte('a' + (i % 26))
	}

	current := root
	for d := 0; d < depth; d++ {
		// Create files at current level
		for i := 0; i < filesPerLevel; i++ {
			ext := []string{".txt", ".log", ".pdf"}[i%3]
			filename := filepath.Join(current, "file"+string(rune('0'+i%10))+ext)
			if err := os.WriteFile(filename, content, 0644); err != nil {
				b.Fatalf("failed to create file: %v", err)
			}
		}
		// Create next level directory
		current = filepath.Join(current, "level"+string(rune('0'+d%10)))
		if err := os.MkdirAll(current, 0755); err != nil {
			b.Fatalf("failed to create subdir: %v", err)
		}
	}

	return root
}

// =============================================================================
// File Scanning Benchmarks
// =============================================================================

// BenchmarkFileScanner_ScanFilesRecursively measures performance of recursive file scanning
func BenchmarkFileScanner_ScanFilesRecursively(b *testing.B) {
	benchmarks := []struct {
		name        string
		fileCount   int
		subDirCount int
		fileSize    int
	}{
		{"Small_10files_2dirs", 10, 2, 100},
		{"Medium_100files_10dirs", 100, 10, 1024},
		{"Large_500files_20dirs", 500, 20, 1024},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			root := setupBenchmarkDir(b, bm.fileCount, bm.subDirCount, bm.fileSize)

			fm := filemanager.NewFileManager()
			filter := fm.NewFileFilter(0, 0, nil, nil, time.Time{}, time.Time{})

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				scanner := filemanager.NewFileScanner(fm, filter, false)
				scanner.ScanFilesRecursively(root)
			}
		})
	}
}

// BenchmarkFileScanner_ScanFilesRecursively_Deep measures performance with deep directory nesting
func BenchmarkFileScanner_ScanFilesRecursively_Deep(b *testing.B) {
	benchmarks := []struct {
		name          string
		depth         int
		filesPerLevel int
		fileSize      int
	}{
		{"Depth5_5files", 5, 5, 100},
		{"Depth10_3files", 10, 3, 100},
		{"Depth20_2files", 20, 2, 100},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			root := setupDeepBenchmarkDir(b, bm.depth, bm.filesPerLevel, bm.fileSize)

			fm := filemanager.NewFileManager()
			filter := fm.NewFileFilter(0, 0, nil, nil, time.Time{}, time.Time{})

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				scanner := filemanager.NewFileScanner(fm, filter, false)
				scanner.ScanFilesRecursively(root)
			}
		})
	}
}

// BenchmarkFileScanner_ScanFilesCurrentLevel measures performance of non-recursive file scanning
func BenchmarkFileScanner_ScanFilesCurrentLevel(b *testing.B) {
	benchmarks := []struct {
		name      string
		fileCount int
		fileSize  int
	}{
		{"Small_10files", 10, 100},
		{"Medium_100files", 100, 1024},
		{"Large_500files", 500, 1024},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			root := setupBenchmarkDir(b, bm.fileCount, 0, bm.fileSize)

			fm := filemanager.NewFileManager()
			filter := fm.NewFileFilter(0, 0, nil, nil, time.Time{}, time.Time{})

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				scanner := filemanager.NewFileScanner(fm, filter, false)
				scanner.ScanFilesCurrentLevel(root)
			}
		})
	}
}

// BenchmarkFileScanner_ScanEmptySubFolders measures performance of empty folder detection
func BenchmarkFileScanner_ScanEmptySubFolders(b *testing.B) {
	benchmarks := []struct {
		name        string
		emptyDirs   int
		nonEmptyDir int
	}{
		{"Few_5empty", 5, 2},
		{"Medium_20empty", 20, 5},
		{"Many_50empty", 50, 10},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			root := b.TempDir()

			// Create empty directories
			for i := 0; i < bm.emptyDirs; i++ {
				emptyDir := filepath.Join(root, "empty"+string(rune('0'+i%10))+string(rune('a'+i/10)))
				if err := os.MkdirAll(emptyDir, 0755); err != nil {
					b.Fatalf("failed to create empty dir: %v", err)
				}
			}

			// Create non-empty directories
			for i := 0; i < bm.nonEmptyDir; i++ {
				nonEmptyDir := filepath.Join(root, "nonempty"+string(rune('0'+i%10)))
				if err := os.MkdirAll(nonEmptyDir, 0755); err != nil {
					b.Fatalf("failed to create non-empty dir: %v", err)
				}
				filename := filepath.Join(nonEmptyDir, "file.txt")
				if err := os.WriteFile(filename, []byte("content"), 0644); err != nil {
					b.Fatalf("failed to create file: %v", err)
				}
			}

			fm := filemanager.NewFileManager()
			filter := fm.NewFileFilter(0, 0, nil, nil, time.Time{}, time.Time{})

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				scanner := filemanager.NewFileScanner(fm, filter, false)
				scanner.ScanEmptySubFolders(root)
			}
		})
	}
}

// =============================================================================
// File Filtering Benchmarks
// =============================================================================

// BenchmarkFileFilter_MatchesFilters measures performance of filter matching operations
func BenchmarkFileFilter_MatchesFilters(b *testing.B) {
	fm := filemanager.NewFileManager()

	// Create a temporary file to get FileInfo
	tmpFile, err := os.CreateTemp("", "bench_*.txt")
	if err != nil {
		b.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Write([]byte("benchmark content for size testing"))
	tmpFile.Close()

	fileInfo, err := os.Stat(tmpFile.Name())
	if err != nil {
		b.Fatalf("failed to stat temp file: %v", err)
	}

	benchmarks := []struct {
		name       string
		minSize    int64
		maxSize    int64
		extensions []string
		exclude    []string
		olderThan  time.Time
		newerThan  time.Time
	}{
		{"NoFilters", 0, 0, nil, nil, time.Time{}, time.Time{}},
		{"SizeFilter", 10, 1000, nil, nil, time.Time{}, time.Time{}},
		{"ExtensionFilter", 0, 0, []string{".txt", ".log", ".pdf"}, nil, time.Time{}, time.Time{}},
		{"ExcludeFilter", 0, 0, nil, []string{"node_modules", "vendor", ".git"}, time.Time{}, time.Time{}},
		{"DateFilter", 0, 0, nil, nil, time.Now().Add(-24 * time.Hour), time.Time{}},
		{"CombinedFilters", 10, 10000, []string{".txt"}, []string{"backup"}, time.Now().Add(-7 * 24 * time.Hour), time.Now()},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			filter := fm.NewFileFilter(
				bm.minSize,
				bm.maxSize,
				utils.ParseExtToMap(bm.extensions),
				bm.exclude,
				bm.olderThan,
				bm.newerThan,
			)

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				filter.MatchesFilters(fileInfo, tmpFile.Name())
			}
		})
	}
}

// BenchmarkFileFilter_ExcludeFilter measures performance of exclude pattern matching
func BenchmarkFileFilter_ExcludeFilter(b *testing.B) {
	fm := filemanager.NewFileManager()

	tmpFile, err := os.CreateTemp("", "bench_exclude_*.txt")
	if err != nil {
		b.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	fileInfo, err := os.Stat(tmpFile.Name())
	if err != nil {
		b.Fatalf("failed to stat temp file: %v", err)
	}

	benchmarks := []struct {
		name    string
		exclude []string
	}{
		{"NoExclude", nil},
		{"SinglePattern", []string{"node_modules"}},
		{"FewPatterns", []string{"node_modules", "vendor", ".git"}},
		{"ManyPatterns", []string{
			"node_modules", "vendor", ".git", "build", "dist",
			"cache", "tmp", "temp", "logs", "backup",
		}},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			filter := fm.NewFileFilter(0, 0, nil, bm.exclude, time.Time{}, time.Time{})

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				filter.ExcludeFilter(fileInfo, tmpFile.Name())
			}
		})
	}
}

// =============================================================================
// Directory Size Calculation Benchmarks
// =============================================================================

// BenchmarkCalculateDirSize measures performance of directory size calculation
func BenchmarkCalculateDirSize(b *testing.B) {
	benchmarks := []struct {
		name        string
		fileCount   int
		subDirCount int
		fileSize    int
	}{
		{"Small_10files_2dirs", 10, 2, 100},
		{"Medium_50files_5dirs", 50, 5, 1024},
		{"Large_100files_10dirs", 100, 10, 4096},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			root := setupBenchmarkDir(b, bm.fileCount, bm.subDirCount, bm.fileSize)
			fm := filemanager.NewFileManager()

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				fm.CalculateDirSize(root)
			}
		})
	}
}

// BenchmarkCalculateDirSize_Deep measures performance with deeply nested directories
func BenchmarkCalculateDirSize_Deep(b *testing.B) {
	benchmarks := []struct {
		name          string
		depth         int
		filesPerLevel int
		fileSize      int
	}{
		{"Depth5_5files", 5, 5, 1024},
		{"Depth10_3files", 10, 3, 1024},
		{"Depth15_2files", 15, 2, 1024},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			root := setupDeepBenchmarkDir(b, bm.depth, bm.filesPerLevel, bm.fileSize)
			fm := filemanager.NewFileManager()

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				fm.CalculateDirSize(root)
			}
		})
	}
}

// =============================================================================
// Concurrent Operations Benchmarks
// =============================================================================

// BenchmarkWalkFilesWithFilter measures performance of concurrent file walking
func BenchmarkWalkFilesWithFilter(b *testing.B) {
	benchmarks := []struct {
		name        string
		fileCount   int
		subDirCount int
		fileSize    int
	}{
		{"Small_10files_2dirs", 10, 2, 100},
		{"Medium_100files_10dirs", 100, 10, 1024},
		{"Large_500files_20dirs", 500, 20, 1024},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			root := setupBenchmarkDir(b, bm.fileCount, bm.subDirCount, bm.fileSize)
			fm := filemanager.NewFileManager()
			filter := fm.NewFileFilter(0, 0, nil, nil, time.Time{}, time.Time{})

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				count := 0
				fm.WalkFilesWithFilter(func(fi os.FileInfo, path string) {
					count++
				}, root, filter)
			}
		})
	}
}

// BenchmarkWalkFilesWithFilter_WithExtensionFilter measures concurrent walking with extension filtering
func BenchmarkWalkFilesWithFilter_WithExtensionFilter(b *testing.B) {
	root := setupBenchmarkDir(b, 100, 10, 1024)
	fm := filemanager.NewFileManager()

	extensions := []string{".txt", ".log"}
	filter := fm.NewFileFilter(0, 0, utils.ParseExtToMap(extensions), nil, time.Time{}, time.Time{})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		count := 0
		fm.WalkFilesWithFilter(func(fi os.FileInfo, path string) {
			count++
		}, root, filter)
	}
}

// =============================================================================
// Utility Function Benchmarks
// =============================================================================

// BenchmarkIsEmptyDir measures performance of empty directory detection
func BenchmarkIsEmptyDir(b *testing.B) {
	fm := filemanager.NewFileManager()

	b.Run("EmptyDir", func(b *testing.B) {
		root := b.TempDir()
		emptyDir := filepath.Join(root, "empty")
		if err := os.MkdirAll(emptyDir, 0755); err != nil {
			b.Fatalf("failed to create empty dir: %v", err)
		}

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			fm.IsEmptyDir(emptyDir)
		}
	})

	b.Run("NonEmptyDir", func(b *testing.B) {
		root := b.TempDir()
		nonEmptyDir := filepath.Join(root, "nonempty")
		if err := os.MkdirAll(nonEmptyDir, 0755); err != nil {
			b.Fatalf("failed to create dir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(nonEmptyDir, "file.txt"), []byte("content"), 0644); err != nil {
			b.Fatalf("failed to create file: %v", err)
		}

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			fm.IsEmptyDir(nonEmptyDir)
		}
	})

	b.Run("NestedEmptyDirs", func(b *testing.B) {
		root := b.TempDir()
		nestedDir := filepath.Join(root, "level1", "level2", "level3")
		if err := os.MkdirAll(nestedDir, 0755); err != nil {
			b.Fatalf("failed to create nested dirs: %v", err)
		}

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			fm.IsEmptyDir(filepath.Join(root, "level1"))
		}
	})
}

// BenchmarkExpandTilde measures performance of tilde expansion
func BenchmarkExpandTilde(b *testing.B) {
	fm := filemanager.NewFileManager()

	benchmarks := []struct {
		name string
		path string
	}{
		{"NoTilde", "/home/user/documents"},
		{"WithTilde", "~/documents/files"},
		{"TildeOnly", "~"},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				fm.ExpandTilde(bm.path)
			}
		})
	}
}
