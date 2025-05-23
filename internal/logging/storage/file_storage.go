package storage

import (
	"encoding/json"
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
	// Реализация чтения статистики
	return nil, nil
}

func (fs *FileStorage) GetOperations(scanID string) ([]logging.FileOperation, error) {
	// Реализация чтения операций
	return nil, nil
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
	// Реализация добавления в файл
	return nil
}
