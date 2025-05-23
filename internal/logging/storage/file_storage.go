package storage

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/pashkov256/deletor/internal/logging"
)

type FileStorage struct {
	basePath string
	mu       sync.RWMutex
}

func NewFileStorage(basePath string) *FileStorage {
	return &FileStorage{
		basePath: basePath,
	}
}

func (fs *FileStorage) SaveStatistics(stats *logging.ScanStatistics) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// Create statistics directory if it doesn't exist
	if err := os.MkdirAll(fs.basePath, 0755); err != nil {
		return fmt.Errorf("failed to create statistics directory: %w", err)
	}

	path := filepath.Join(fs.basePath, "statistics.json")
	return fs.saveToFile(path, stats)
}

func (fs *FileStorage) SaveOperation(operation *logging.FileOperation) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	path := filepath.Join(fs.basePath, "operations.json")
	return fs.appendToFile(path, operation)
}

func (fs *FileStorage) GetStatistics(scanID string) (*logging.ScanStatistics, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	path := filepath.Join(fs.basePath, "statistics.json")
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var stats logging.ScanStatistics
	if err := json.NewDecoder(file).Decode(&stats); err != nil {
		return nil, err
	}

	return &stats, nil
}

func (fs *FileStorage) GetOperations(scanID string) ([]logging.FileOperation, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	path := filepath.Join(fs.basePath, "operations.json")
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []logging.FileOperation{}, nil
		}
		return nil, err
	}
	defer file.Close()

	var operations []logging.FileOperation
	if err := json.NewDecoder(file).Decode(&operations); err != nil {
		return nil, err
	}

	return operations, nil
}

// Вспомогательные методы
func (fs *FileStorage) saveToFile(path string, data interface{}) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(data)
}

func (fs *FileStorage) appendToFile(path string, data interface{}) error {
	// Read existing operations
	var operations []interface{}
	if file, err := os.Open(path); err == nil {
		defer file.Close()
		if err := json.NewDecoder(file).Decode(&operations); err != nil && err != io.EOF {
			return err
		}
	}

	// Append new operation
	operations = append(operations, data)

	// Write back to file
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(operations)
}
