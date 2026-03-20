package logging_test

import (
	"path/filepath"
	"testing"

	"github.com/pashkov256/deletor/internal/logging"
	"github.com/pashkov256/deletor/internal/path"
)

func TestGetLogFilePath(t *testing.T) {
	got := logging.GetLogFilePath()

	if filepath.Base(got) != path.LogFileName {
		t.Errorf("expected log file name %q, got %q",
			path.LogFileName,
			filepath.Base(got),
		)
	}

	if filepath.Base(filepath.Dir(got)) != path.AppDirName {
		t.Errorf("expected app dir %q, got %q",
			path.AppDirName,
			filepath.Base(filepath.Dir(got)),
		)
	}
}
