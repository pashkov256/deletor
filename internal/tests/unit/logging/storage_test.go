package logging

import (
	"os"
	"testing"
	"time"

	"github.com/pashkov256/deletor/internal/logging"
	"github.com/pashkov256/deletor/internal/logging/storage"
)

func setupTestStorage(t *testing.T) (*storage.FileStorage, string) {
	t.Helper()
	tempDir, err := os.MkdirTemp("", "storage_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	return storage.NewFileStorage(tempDir), tempDir
}

func cleanupTestStorage(t *testing.T, tempDir string) {
	t.Helper()

	if err := os.RemoveAll(tempDir); err != nil {
		t.Errorf("Failed to remove temp dir: %v", err)
	}
}

func TestSaveAndReadStatistics(t *testing.T) {
	fs, tempDir := setupTestStorage(t)
	defer cleanupTestStorage(t, tempDir)

	stats := &logging.ScanStatistics{
		TotalFiles:    100,
		TotalSize:     1000,
		DeletedFiles:  50,
		DeletedSize:   500,
		TrashedFiles:  30,
		TrashedSize:   300,
		IgnoredFiles:  20,
		IgnoredSize:   200,
		StartTime:     time.Now(),
		EndTime:       time.Now().Add(time.Hour),
		Directory:     "/test/dir",
		OperationType: "scan",
	}

	if err := fs.SaveStatistics(stats); err != nil {
		t.Fatalf("Failed to save statistics: %v", err)
	}

	readStats, err := fs.GetStatistics("test")
	if err != nil {
		t.Fatalf("Failed to read statistics: %v", err)
	}
	if readStats.TotalFiles != stats.TotalFiles {
		t.Errorf("TotalFiles = %v, want %v", readStats.TotalFiles, stats.TotalFiles)
	}
	if readStats.TotalSize != stats.TotalSize {
		t.Errorf("TotalSize = %v, want %v", readStats.TotalSize, stats.TotalSize)
	}
	if readStats.DeletedFiles != stats.DeletedFiles {
		t.Errorf("DeletedFiles = %v, want %v", readStats.DeletedFiles, stats.DeletedFiles)
	}
	if readStats.TrashedFiles != stats.TrashedFiles {
		t.Errorf("TrashedFiles = %v, want %v", readStats.TrashedFiles, stats.TrashedFiles)
	}
	if readStats.IgnoredFiles != stats.IgnoredFiles {
		t.Errorf("IgnoredFiles = %v, want %v", readStats.IgnoredFiles, stats.IgnoredFiles)
	}
	if readStats.Directory != stats.Directory {
		t.Errorf("Directory = %v, want %v", readStats.Directory, stats.Directory)
	}
	if readStats.OperationType != stats.OperationType {
		t.Errorf("OperationType = %v, want %v", readStats.OperationType, stats.OperationType)
	}
}

func TestSaveAndReadOperations(t *testing.T) {
	fs, tempDir := setupTestStorage(t)
	defer cleanupTestStorage(t, tempDir)

	// Create test operations
	operations := []*logging.FileOperation{
		{
			FilePath:      "/test/file1.txt",
			FileSize:      100,
			OperationType: logging.OperationDeleted,
			Reason:        "Test deletion",
			RuleApplied:   "size_rule",
		},
		{
			FilePath:      "/test/file2.txt",
			FileSize:      200,
			OperationType: logging.OperationTrashed,
			Reason:        "Test trash",
			RuleApplied:   "date_rule",
		},
		{
			FilePath:      "/test/file3.txt",
			FileSize:      300,
			OperationType: logging.OperationIgnored,
			Reason:        "Test ignore",
			RuleApplied:   "pattern_rule",
		},
	}

	for _, op := range operations {
		if err := fs.SaveOperation(op); err != nil {
			t.Fatalf("Failed to save operation: %v", err)
		}
	}

	readOps, err := fs.GetOperations("test")
	if err != nil {
		t.Fatalf("Failed to read operations: %v", err)
	}

	if len(readOps) != len(operations) {
		t.Fatalf("Expected %d operations, got %d", len(operations), len(readOps))
	}

	for i, op := range operations {
		readOp := readOps[i]
		if readOp.FilePath != op.FilePath {
			t.Errorf("Operation %d: FilePath = %v, want %v", i, readOp.FilePath, op.FilePath)
		}
		if readOp.FileSize != op.FileSize {
			t.Errorf("Operation %d: FileSize = %v, want %v", i, readOp.FileSize, op.FileSize)
		}
		if readOp.OperationType != op.OperationType {
			t.Errorf("Operation %d: OperationType = %v, want %v", i, readOp.OperationType, op.OperationType)
		}
		if readOp.Reason != op.Reason {
			t.Errorf("Operation %d: Reason = %v, want %v", i, readOp.Reason, op.Reason)
		}
		if readOp.RuleApplied != op.RuleApplied {
			t.Errorf("Operation %d: RuleApplied = %v, want %v", i, readOp.RuleApplied, op.RuleApplied)
		}
	}
}

func TestConcurrentStorageOperations(t *testing.T) {
	fs, tempDir := setupTestStorage(t)
	defer cleanupTestStorage(t, tempDir)

	stats := &logging.ScanStatistics{
		TotalFiles:    100,
		TotalSize:     1000,
		DeletedFiles:  50,
		DeletedSize:   500,
		TrashedFiles:  30,
		TrashedSize:   300,
		IgnoredFiles:  20,
		IgnoredSize:   200,
		StartTime:     time.Now(),
		EndTime:       time.Now().Add(time.Hour),
		Directory:     "/test/dir",
		OperationType: "scan",
	}

	op := &logging.FileOperation{
		FilePath:      "/test/file1.txt",
		FileSize:      100,
		OperationType: logging.OperationDeleted,
		Reason:        "Test deletion",
		RuleApplied:   "size_rule",
	}

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			if err := fs.SaveStatistics(stats); err != nil {
				t.Errorf("Failed to save statistics in goroutine %d: %v", id, err)
			}

			if err := fs.SaveOperation(op); err != nil {
				t.Errorf("Failed to save operation in goroutine %d: %v", id, err)
			}

			if _, err := fs.GetStatistics("test"); err != nil {
				t.Errorf("Failed to read statistics in goroutine %d: %v", id, err)
			}

			if _, err := fs.GetOperations("test"); err != nil {
				t.Errorf("Failed to read operations in goroutine %d: %v", id, err)
			}

			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	readStats, err := fs.GetStatistics("test")
	if err != nil {
		t.Fatalf("Failed to read final statistics: %v", err)
	}

	if readStats.TotalFiles != stats.TotalFiles {
		t.Errorf("Final TotalFiles = %v, want %v", readStats.TotalFiles, stats.TotalFiles)
	}

	readOps, err := fs.GetOperations("test")
	if err != nil {
		t.Fatalf("Failed to read final operations: %v", err)
	}

	if len(readOps) == 0 {
		t.Error("Expected at least one operation after concurrent writes")
	}
}
