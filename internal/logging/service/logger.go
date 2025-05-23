package service

import (
	"time"

	"github.com/pashkov256/deletor/internal/logging"
	"github.com/pashkov256/deletor/internal/logging/storage"
	"github.com/pashkov256/deletor/internal/utils"
)

type Logger struct {
	storage     storage.FileStorage
	currentScan *logging.ScanStatistics
}

func NewLogger(storage storage.FileStorage) *Logger {
	return &Logger{
		storage: storage,
	}
}

func (l *Logger) StartScan(path string) (string, error) {
	scanID := utils.GenerateUUID() // Нужно реализовать
	l.currentScan = &logging.ScanStatistics{
		ScanID:    scanID,
		StartTime: time.Now(),
		ScanPath:  path,
	}
	return scanID, nil
}

func (l *Logger) EndScan(scanID string) error {
	// if l.currentScan == nil || l.currentScan.ScanID != scanID {
	// 	return ErrInvalidScanID
	// }

	l.currentScan.EndTime = time.Now()
	// l.currentScan.CalculateDuration()

	return l.storage.SaveStatistics(l.currentScan)
}

func (l *Logger) LogOperation(operation *logging.FileOperation) error {
	return l.storage.SaveOperation(operation)
}

func (l *Logger) UpdateStatistics(stats *logging.ScanStatistics) error {
	return l.storage.SaveStatistics(stats)
}
