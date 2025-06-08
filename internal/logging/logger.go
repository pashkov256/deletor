package logging

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// LogLevel represents the severity level of a log message
type LogLevel string

const (
	INFO  LogLevel = "INFO"  // Informational messages
	DEBUG LogLevel = "DEBUG" // Debug messages
	ERROR LogLevel = "ERROR" // Error messages
)

// ScanStatistics tracks metrics for file scanning operations
type ScanStatistics struct {
	TotalFiles    int64     // Total number of files processed
	TotalSize     int64     // Total size of all files
	DeletedFiles  int64     // Number of files deleted
	DeletedSize   int64     // Size of deleted files
	TrashedFiles  int64     // Number of files moved to trash
	TrashedSize   int64     // Size of trashed files
	IgnoredFiles  int64     // Number of ignored files
	IgnoredSize   int64     // Size of ignored files
	StartTime     time.Time // Operation start time
	EndTime       time.Time // Operation end time
	Directory     string    // Target directory
	OperationType string    // Type of operation performed
}

// LogEntry represents a single log entry with metadata
type LogEntry struct {
	Timestamp time.Time       `json:"timestamp"`       // When the entry was created
	Level     LogLevel        `json:"level"`           // Log level
	Message   string          `json:"message"`         // Log message
	Stats     *ScanStatistics `json:"stats,omitempty"` // Optional scan statistics
}

// Logger handles writing log entries to a file with thread safety
type Logger struct {
	mu            sync.Mutex
	logFile       *os.File
	ConfigPath    string
	currentScan   *ScanStatistics
	StatsCallback func(*ScanStatistics) // Callback for stats updates
}

// NewLogger creates a new logger instance with the specified configuration
func NewLogger(ConfigPath string, statsCallback func(*ScanStatistics)) (*Logger, error) {
	// Ensure config directory exists
	if err := os.MkdirAll(filepath.Dir(ConfigPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	// Open or create log file
	logFile, err := os.OpenFile(ConfigPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	return &Logger{
		logFile:       logFile,
		ConfigPath:    ConfigPath,
		StatsCallback: statsCallback,
	}, nil
}

// Log writes a log entry with the specified level and message
func (l *Logger) Log(level LogLevel, message string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
		Stats:     l.currentScan,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal log entry: %w", err)
	}

	if _, err := l.logFile.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("failed to write to log file: %w", err)
	}

	return nil
}

// UpdateStats updates the current scan statistics and triggers callback if set
func (l *Logger) UpdateStats(stats *ScanStatistics) {
	l.mu.Lock()
	l.currentScan = stats
	l.mu.Unlock()

	if l.StatsCallback != nil {
		l.StatsCallback(stats)
	}
}

// Close closes the log file
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.logFile.Close()
}
