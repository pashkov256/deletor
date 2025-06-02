package filemanager

import (
	"os"
	"time"
)

// FileManager defines the interface for file system operations
type FileManager interface {
	NewFileFilter(minSize, maxSize int64, extensions map[string]struct{}, exclude []string, olderThan, newerThan time.Time) *FileFilter
	WalkFilesWithFilter(callback func(fi os.FileInfo, path string), dir string, filter *FileFilter)
	MoveFilesToTrash(dir string, extensions []string, exclude []string, minSize, maxSize int64, olderThan, newerThan time.Time)
	DeleteFiles(dir string, extensions []string, exclude []string, minSize, maxSize int64, olderThan, newerThan time.Time)
	DeleteEmptySubfolders(dir string)
	IsEmptyDir(dir string) bool
	ExpandTilde(path string) string
	CalculateDirSize(path string) int64
	MoveFileToTrash(filePath string)
}

// FileTask represents a file operation task
type FileTask struct {
	info os.FileInfo
}

// defaultFileManager implements the FileManager interface
type defaultFileManager struct {
}

// NewFileManager creates a new instance of the default file manager
func NewFileManager() FileManager {
	return &defaultFileManager{}
}
