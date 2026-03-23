package logging

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/pashkov256/deletor/internal/path"
)

func TestGetLogFilePath_NotEmpty(t *testing.T) {
	result := GetLogFilePath()

	if result == "" {
		t.Fatal("expected non-empty path")
	}
}

func TestGetLogFilePath_ContainsAppDir(t *testing.T) {
	result := GetLogFilePath()

	if !strings.Contains(result, path.AppDirName) {
		t.Errorf("expected path to contain app dir %q, got %q", path.AppDirName, result)
	}
}

func TestGetLogFilePath_HasCorrectFileName(t *testing.T) {
	result := GetLogFilePath()

	if filepath.Base(result) != path.LogFileName {
		t.Errorf("expected file name %q, got %q", path.LogFileName, filepath.Base(result))
	}
}

func TestGetLogFilePath_HasCorrectSuffix(t *testing.T) {
	result := GetLogFilePath()

	expected := filepath.Join(path.AppDirName, path.LogFileName)

	if !strings.HasSuffix(result, expected) {
		t.Errorf("expected path to end with %q, got %q", expected, result)
	}
}
