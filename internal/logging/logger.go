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
	INFO  LogLevel = "INFO"
	DEBUG LogLevel = "DEBUG"
	ERROR LogLevel = "ERROR"
)

// ScanStatistics represents statistics for a file scanning operation
type ScanStatistics struct {
	TotalFiles    int64
	TotalSize     int64
	DeletedFiles  int64
	DeletedSize   int64
	TrashedFiles  int64
	TrashedSize   int64
	IgnoredFiles  int64
	IgnoredSize   int64
	StartTime     time.Time
	EndTime       time.Time
	Directory     string
	OperationType string
}

type LogEntry struct {
	Timestamp time.Time       `json:"timestamp"`
	Level     LogLevel        `json:"level"`
	Message   string          `json:"message"`
	Stats     *ScanStatistics `json:"stats,omitempty"`
}

type Logger struct {
	mu            sync.Mutex
	logFile       *os.File
	configPath    string
	currentScan   *ScanStatistics
	StatsCallback func(*ScanStatistics)
}

func NewLogger(configPath string, statsCallback func(*ScanStatistics)) (*Logger, error) {
	// Ensure config directory exists
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	// Open or create log file
	logFile, err := os.OpenFile(configPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	return &Logger{
		logFile:       logFile,
		configPath:    configPath,
		StatsCallback: statsCallback,
	}, nil
}

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

func (l *Logger) UpdateStats(stats *ScanStatistics) {
	l.mu.Lock()
	l.currentScan = stats
	l.mu.Unlock()

	if l.StatsCallback != nil {
		l.StatsCallback(stats)
	}
}

func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.logFile.Close()
}
