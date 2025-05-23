package logging

import (
	"os"
	"path/filepath"

	"github.com/pashkov256/deletor/internal/path"
)

func GetLogFilePath() string {
	userConfigDir, _ := os.UserConfigDir()
	fileLogPath := filepath.Join(userConfigDir, path.AppDirName, path.LogFileName)

	return fileLogPath
}
