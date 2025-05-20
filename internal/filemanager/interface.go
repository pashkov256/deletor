package filemanager

import (
	"os"
	"time"
)

type FileManager interface {
	DeleteFiles(dir string, extensions []string, exclude []string, minSize int64)
	DeleteEmptySubfolders(dir string)
	IsEmptyDir(dir string) bool
	ExpandTilde(path string) string
	CalculateDirSize(path string) int64
	NewFileFilter(minSize, maxSize int64, extensions map[string]struct{}, exclude []string, olderThan, newerThan time.Time) *FileFilter
}

type FileTask struct {
	info os.FileInfo
}

type defaultFileManager struct {
}

func NewFileManager() FileManager {
	return &defaultFileManager{}
}
