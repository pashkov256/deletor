package logging

import (
	"strings"
	"testing"
)

func TestGetLogFilePath(t *testing.T) {
	logPath := GetLogFilePath()

	// Test that the path is not empty
	if logPath == "" {
		t.Error("GetLogFilePath() returned an empty string")
	}

	// Test that the path contains the expected app directory name
	if !strings.Contains(logPath, "deletor") {
		t.Errorf("GetLogFilePath() = %q, expected path to contain 'deletor'", logPath)
	}

	// Test that the path ends with the expected log file name
	expectedSuffix := "deletor.log"
	if !strings.HasSuffix(logPath, expectedSuffix) {
		t.Errorf("GetLogFilePath() = %q, expected path to end with %q", logPath, expectedSuffix)
	}

	// Test that the path contains both the app directory and log file
	// e.g., "deletor/deletor.log" should appear in the path
	expectedSegment := "deletor/deletor.log"
	// Handle Windows paths where separator might be backslash
	expectedSegmentWin := "deletor\\deletor.log"
	if !strings.Contains(logPath, expectedSegment) && !strings.Contains(logPath, expectedSegmentWin) {
		t.Errorf("GetLogFilePath() = %q, expected path to contain %q or %q", logPath, expectedSegment, expectedSegmentWin)
	}
}
