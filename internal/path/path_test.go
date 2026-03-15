package path

import (
	"testing"
)

func TestConstants(t *testing.T) {
	// Test AppDirName
	if AppDirName == "" {
		t.Error("AppDirName should not be empty")
	}
	if AppDirName != "deletor" {
		t.Errorf("AppDirName = %q, want %q", AppDirName, "deletor")
	}

	// Test RuleFileName
	if RuleFileName == "" {
		t.Error("RuleFileName should not be empty")
	}
	if RuleFileName != "rule.json" {
		t.Errorf("RuleFileName = %q, want %q", RuleFileName, "rule.json")
	}

	// Test LogFileName
	if LogFileName == "" {
		t.Error("LogFileName should not be empty")
	}
	if LogFileName != "deletor.log" {
		t.Errorf("LogFileName = %q, want %q", LogFileName, "deletor.log")
	}
}