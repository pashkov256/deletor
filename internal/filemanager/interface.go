package filemanager

import (
	"os"
	"time"
)

type FileManager interface {
	NewFileFilter(minSize, maxSize int64, extensions map[string]struct{}, exclude []string, olderThan, newerThan time.Time) *FileFilter
	WalkFilesWithFilter(callback func(fi os.FileInfo), dir string, extensions []string, exclude []string, minSize, maxSize int64, olderThan, newerThan time.Time)
	MoveFilesToTrash(dir string, extensions []string, exclude []string, minSize, maxSize int64, olderThan, newerThan time.Time)
	DeleteFiles(dir string, extensions []string, exclude []string, minSize, maxSize int64, olderThan, newerThan time.Time)
	DeleteEmptySubfolders(dir string)
	IsEmptyDir(dir string) bool
	ExpandTilde(path string) string
	CalculateDirSize(path string) int64
	MoveFileToTrash(filePath string)
}

type FileTask struct {
	info os.FileInfo
}

type defaultFileManager struct {
}

func NewFileManager() FileManager {
	return &defaultFileManager{}
}
