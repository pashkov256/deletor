package service

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/pashkov256/deletor/internal/logging"
	"github.com/pashkov256/deletor/internal/logging/storage"
	"github.com/pashkov256/deletor/internal/utils"
)

type LogLevel string

const (
	INFO  LogLevel = "INFO"
	DEBUG LogLevel = "DEBUG"
	ERROR LogLevel = "ERROR"
)

type LogEntry struct {
	Timestamp time.Time               `json:"timestamp"`
	Level     LogLevel                `json:"level"`
	Message   string                  `json:"message"`
	Stats     *logging.ScanStatistics `json:"stats,omitempty"`
}

type Logger struct {
	mu            sync.Mutex
	logFile       *os.File
	configPath    string
	storage       *storage.FileStorage
	currentScan   *logging.ScanStatistics
	statsCallback func(*logging.ScanStatistics)
}

func NewLogger(configPath string, statsCallback func(*logging.ScanStatistics)) (*Logger, error) {
	// Ensure config directory exists
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	// Open or create log file
	logFile, err := os.OpenFile(configPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	// Initialize storage
	storage := storage.NewFileStorage(filepath.Dir(configPath))

	return &Logger{
		logFile:       logFile,
		configPath:    configPath,
		storage:       storage,
		statsCallback: statsCallback,
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

func (l *Logger) StartScan(path string) (string, error) {
	scanID := utils.GenerateUUID()
	l.currentScan = &logging.ScanStatistics{
		StartTime:     time.Now(),
		Directory:     path,
		OperationType: "scan",
	}
	return scanID, nil
}

func (l *Logger) EndScan(scanID string) error {
	l.currentScan.EndTime = time.Now()

	return l.storage.SaveStatistics(l.currentScan)
}

func (l *Logger) LogOperation(operation *logging.FileOperation) error {
	return l.storage.SaveOperation(operation)
}

func (l *Logger) UpdateStats(stats *logging.ScanStatistics) {
	l.mu.Lock()
	l.currentScan = stats
	l.mu.Unlock()

	if l.statsCallback != nil {
		l.statsCallback(stats)
	}
}

func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.logFile.Close()
}
