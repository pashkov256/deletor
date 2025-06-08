package logging

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/pashkov256/deletor/internal/logging"
)

// setupTestLogger creates a temporary logger for testing
func setupTestLogger(t *testing.T) (*logging.Logger, string) {
	t.Helper()

	tempDir, err := os.MkdirTemp("", "logger_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	logPath := filepath.Join(tempDir, "test.log")

	logger, err := logging.NewLogger(logPath, nil)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	return logger, tempDir
}

// cleanupTestLogger removes temporary test files
func cleanupTestLogger(t *testing.T, logger *logging.Logger, tempDir string) {
	t.Helper()

	if err := logger.Close(); err != nil {
		t.Errorf("Failed to close logger: %v", err)
	}

	if err := os.RemoveAll(tempDir); err != nil {
		t.Errorf("Failed to remove temp dir: %v", err)
	}
}

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name        string
		ConfigPath  string
		setup       func() error
		cleanup     func() error
		wantErr     bool
		description string
	}{
		{
			name:       "Valid config path",
			ConfigPath: filepath.Join(os.TempDir(), "valid_test.log"),
			setup: func() error {
				return os.MkdirAll(filepath.Dir(filepath.Join(os.TempDir(), "valid_test.log")), 0755)
			},
			cleanup: func() error {
				return os.RemoveAll(filepath.Join(os.TempDir(), "valid_test.log"))
			},
			wantErr:     false,
			description: "Should create logger with valid config path",
		},
		{
			name:        "Invalid config path",
			ConfigPath:  "",
			setup:       func() error { return nil },
			cleanup:     func() error { return nil },
			wantErr:     true,
			description: "Should fail with empty config path",
		},
		{
			name:        "Non-existent directory",
			ConfigPath:  filepath.Join(os.TempDir(), "nonexistent", "test.log"),
			setup:       func() error { return nil },
			cleanup:     func() error { return nil },
			wantErr:     false,
			description: "Should create directory and logger",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.setup(); err != nil {
				t.Fatalf("Setup failed: %v", err)
			}
			defer tt.cleanup()

			logger, err := logging.NewLogger(tt.ConfigPath, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewLogger() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil {
				logger.Close()
			}
		})
	}
}

func TestLogLevels(t *testing.T) {
	logger, tempDir := setupTestLogger(t)
	defer cleanupTestLogger(t, logger, tempDir)

	tests := []struct {
		name    string
		level   logging.LogLevel
		message string
	}{
		{"INFO level", logging.INFO, "Test info message"},
		{"DEBUG level", logging.DEBUG, "Test debug message"},
		{"ERROR level", logging.ERROR, "Test error message"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := logger.Log(tt.level, tt.message); err != nil {
				t.Errorf("Log() error = %v", err)
			}
		})
	}

	// Read and verify all log entries
	data, err := os.ReadFile(logger.ConfigPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	// Split the file content into individual log entries
	entries := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(entries) != len(tests) {
		t.Fatalf("Expected %d log entries, got %d", len(tests), len(entries))
	}

	// Verify each log entry
	for i, entryStr := range entries {
		var entry logging.LogEntry
		if err := json.Unmarshal([]byte(entryStr), &entry); err != nil {
			t.Fatalf("Failed to unmarshal log entry %d: %v", i, err)
		}

		expected := tests[i]
		if entry.Level != expected.level {
			t.Errorf("Entry %d: Log level = %v, want %v", i, entry.Level, expected.level)
		}
		if entry.Message != expected.message {
			t.Errorf("Entry %d: Message = %v, want %v", i, entry.Message, expected.message)
		}
	}
}

func TestScanStatistics(t *testing.T) {

	logger, tempDir := setupTestLogger(t)
	defer cleanupTestLogger(t, logger, tempDir)

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

	// Test stats update
	logger.UpdateStats(stats)

	// Log an entry with stats
	if err := logger.Log(logging.INFO, "Test stats message"); err != nil {
		t.Errorf("Log() error = %v", err)
	}

	// Read and verify the log entry
	data, err := os.ReadFile(logger.ConfigPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	var entry logging.LogEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		t.Fatalf("Failed to unmarshal log entry: %v", err)
	}

	if entry.Stats == nil {
		t.Fatal("Stats should not be nil")
	}

	if entry.Stats.TotalFiles != stats.TotalFiles {
		t.Errorf("TotalFiles = %v, want %v", entry.Stats.TotalFiles, stats.TotalFiles)
	}
	if entry.Stats.TotalSize != stats.TotalSize {
		t.Errorf("TotalSize = %v, want %v", entry.Stats.TotalSize, stats.TotalSize)
	}
}

func TestFileOperations(t *testing.T) {
	tests := []struct {
		name         string
		filePath     string
		size         int64
		opType       logging.OperationType
		reason       string
		rule         string
		expectedType logging.OperationType
	}{
		{
			name:         "Delete operation",
			filePath:     "/test/file1.txt",
			size:         100,
			opType:       logging.OperationDeleted,
			reason:       "Test deletion",
			rule:         "size_rule",
			expectedType: logging.OperationDeleted,
		},
		{
			name:         "Trash operation",
			filePath:     "/test/file2.txt",
			size:         200,
			opType:       logging.OperationTrashed,
			reason:       "Test trash",
			rule:         "date_rule",
			expectedType: logging.OperationTrashed,
		},
		{
			name:         "Ignore operation",
			filePath:     "/test/file3.txt",
			size:         300,
			opType:       logging.OperationIgnored,
			reason:       "Test ignore",
			rule:         "pattern_rule",
			expectedType: logging.OperationIgnored,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			op := logging.NewFileOperation(tt.filePath, tt.size, tt.opType, tt.reason, tt.rule)

			if op.FilePath != tt.filePath {
				t.Errorf("FilePath = %v, want %v", op.FilePath, tt.filePath)
			}
			if op.FileSize != tt.size {
				t.Errorf("FileSize = %v, want %v", op.FileSize, tt.size)
			}
			if op.OperationType != tt.expectedType {
				t.Errorf("OperationType = %v, want %v", op.OperationType, tt.expectedType)
			}
			if op.Reason != tt.reason {
				t.Errorf("Reason = %v, want %v", op.Reason, tt.reason)
			}
			if op.RuleApplied != tt.rule {
				t.Errorf("RuleApplied = %v, want %v", op.RuleApplied, tt.rule)
			}
		})
	}
}

func TestConcurrentLogging(t *testing.T) {
	logger, tempDir := setupTestLogger(t)
	defer cleanupTestLogger(t, logger, tempDir)

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 10; j++ {
				msg := fmt.Sprintf("Concurrent log message %d-%d", id, j)
				if err := logger.Log(logging.INFO, msg); err != nil {
					t.Errorf("Log() error in goroutine %d: %v", id, err)
				}
			}
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	data, err := os.ReadFile(logger.ConfigPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	lines := strings.Split(string(data), "\n")
	// Subtract 1 for the empty line at the end
	if len(lines)-1 != 100 {
		t.Errorf("Expected 100 log entries, got %d", len(lines)-1)
	}
}
