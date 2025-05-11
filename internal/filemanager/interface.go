package filemanager

import "os"

type FileManager interface {
	DeleteFiles(dir string, extensions []string, exclude []string, minSize int64)
	DeleteEmptySubfolders(dir string)
	IsEmptyDir(dir string) bool
	ExpandTilde(path string) string
}

type FileTask struct {
	info os.FileInfo
}

type defaultFileManager struct {
}

func NewFileManager() FileManager {
	return &defaultFileManager{}
}
