package path

import (
	"path/filepath"
	"strings"
	"testing"
)

// TestAppDirName verifies the AppDirName constant is properly defined
func TestAppDirName(t *testing.T) {
	tests := []struct {
		name        string
		checkFunc   func(t *testing.T)
		description string
	}{
		{
			name: "not_empty",
			checkFunc: func(t *testing.T) {
				if AppDirName == "" {
					t.Error("AppDirName should not be empty")
				}
			},
			description: "AppDirName constant must not be empty",
		},
		{
			name: "expected_value",
			checkFunc: func(t *testing.T) {
				expected := "deletor"
				if AppDirName != expected {
					t.Errorf("AppDirName = %q, expected %q", AppDirName, expected)
				}
			},
			description: "AppDirName should equal 'deletor'",
		},
		{
			name: "no_whitespace",
			checkFunc: func(t *testing.T) {
				if strings.TrimSpace(AppDirName) != AppDirName {
					t.Error("AppDirName should not contain leading or trailing whitespace")
				}
			},
			description: "AppDirName should not have leading/trailing whitespace",
		},
		{
			name: "no_path_separators",
			checkFunc: func(t *testing.T) {
				if strings.Contains(AppDirName, string(filepath.Separator)) {
					t.Error("AppDirName should not contain path separators")
				}
			},
			description: "AppDirName should be a simple directory name without path separators",
		},
		{
			name: "valid_directory_name",
			checkFunc: func(t *testing.T) {
				// Check for invalid characters in directory names (common across OS)
				invalidChars := []string{"<", ">", ":", "\"", "|", "?", "*"}
				for _, char := range invalidChars {
					if strings.Contains(AppDirName, char) {
						t.Errorf("AppDirName should not contain invalid character: %q", char)
					}
				}
			},
			description: "AppDirName should not contain characters invalid for directory names",
		},
		{
			name: "lowercase",
			checkFunc: func(t *testing.T) {
				if AppDirName != strings.ToLower(AppDirName) {
					t.Error("AppDirName should be lowercase for cross-platform consistency")
				}
			},
			description: "AppDirName should be lowercase",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.checkFunc(t)
		})
	}
}

// TestRuleFileName verifies the RuleFileName constant is properly defined
func TestRuleFileName(t *testing.T) {
	tests := []struct {
		name        string
		checkFunc   func(t *testing.T)
		description string
	}{
		{
			name: "not_empty",
			checkFunc: func(t *testing.T) {
				if RuleFileName == "" {
					t.Error("RuleFileName should not be empty")
				}
			},
			description: "RuleFileName constant must not be empty",
		},
		{
			name: "expected_value",
			checkFunc: func(t *testing.T) {
				expected := "rule.json"
				if RuleFileName != expected {
					t.Errorf("RuleFileName = %q, expected %q", RuleFileName, expected)
				}
			},
			description: "RuleFileName should equal 'rule.json'",
		},
		{
			name: "has_json_extension",
			checkFunc: func(t *testing.T) {
				if !strings.HasSuffix(RuleFileName, ".json") {
					t.Error("RuleFileName should have .json extension")
				}
			},
			description: "RuleFileName should have JSON extension",
		},
		{
			name: "no_whitespace",
			checkFunc: func(t *testing.T) {
				if strings.TrimSpace(RuleFileName) != RuleFileName {
					t.Error("RuleFileName should not contain leading or trailing whitespace")
				}
			},
			description: "RuleFileName should not have leading/trailing whitespace",
		},
		{
			name: "no_path_separators",
			checkFunc: func(t *testing.T) {
				if strings.Contains(RuleFileName, string(filepath.Separator)) {
					t.Error("RuleFileName should not contain path separators")
				}
			},
			description: "RuleFileName should be a simple filename without path",
		},
		{
			name: "valid_filename",
			checkFunc: func(t *testing.T) {
				// Check for invalid characters in filenames
				invalidChars := []string{"<", ">", ":", "\"", "|", "?", "*"}
				for _, char := range invalidChars {
					if strings.Contains(RuleFileName, char) {
						t.Errorf("RuleFileName should not contain invalid character: %q", char)
					}
				}
			},
			description: "RuleFileName should not contain characters invalid for filenames",
		},
		{
			name: "valid_json_structure",
			checkFunc: func(t *testing.T) {
				ext := filepath.Ext(RuleFileName)
				if ext != ".json" {
					t.Errorf("RuleFileName extension = %q, expected '.json'", ext)
				}
			},
			description: "RuleFileName extension should be exactly '.json'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.checkFunc(t)
		})
	}
}

// TestLogFileName verifies the LogFileName constant is properly defined
func TestLogFileName(t *testing.T) {
	tests := []struct {
		name        string
		checkFunc   func(t *testing.T)
		description string
	}{
		{
			name: "not_empty",
			checkFunc: func(t *testing.T) {
				if LogFileName == "" {
					t.Error("LogFileName should not be empty")
				}
			},
			description: "LogFileName constant must not be empty",
		},
		{
			name: "expected_value",
			checkFunc: func(t *testing.T) {
				expected := "deletor.log"
				if LogFileName != expected {
					t.Errorf("LogFileName = %q, expected %q", LogFileName, expected)
				}
			},
			description: "LogFileName should equal 'deletor.log'",
		},
		{
			name: "has_log_extension",
			checkFunc: func(t *testing.T) {
				if !strings.HasSuffix(LogFileName, ".log") {
					t.Error("LogFileName should have .log extension")
				}
			},
			description: "LogFileName should have .log extension",
		},
		{
			name: "no_whitespace",
			checkFunc: func(t *testing.T) {
				if strings.TrimSpace(LogFileName) != LogFileName {
					t.Error("LogFileName should not contain leading or trailing whitespace")
				}
			},
			description: "LogFileName should not have leading/trailing whitespace",
		},
		{
			name: "no_path_separators",
			checkFunc: func(t *testing.T) {
				if strings.Contains(LogFileName, string(filepath.Separator)) {
					t.Error("LogFileName should not contain path separators")
				}
			},
			description: "LogFileName should be a simple filename without path",
		},
		{
			name: "valid_filename",
			checkFunc: func(t *testing.T) {
				// Check for invalid characters in filenames
				invalidChars := []string{"<", ">", ":", "\"", "|", "?", "*"}
				for _, char := range invalidChars {
					if strings.Contains(LogFileName, char) {
						t.Errorf("LogFileName should not contain invalid character: %q", char)
					}
				}
			},
			description: "LogFileName should not contain characters invalid for filenames",
		},
		{
			name: "valid_log_structure",
			checkFunc: func(t *testing.T) {
				ext := filepath.Ext(LogFileName)
				if ext != ".log" {
					t.Errorf("LogFileName extension = %q, expected '.log'", ext)
				}
			},
			description: "LogFileName extension should be exactly '.log'",
		},
		{
			name: "matches_app_name",
			checkFunc: func(t *testing.T) {
				// Log filename should start with app name for consistency
				if !strings.HasPrefix(LogFileName, AppDirName) {
					t.Errorf("LogFileName should start with AppDirName (%q), got %q", AppDirName, LogFileName)
				}
			},
			description: "LogFileName should start with application name for consistency",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.checkFunc(t)
		})
	}
}

// TestPathConstantsIntegration verifies that all constants work together correctly
func TestPathConstantsIntegration(t *testing.T) {
	tests := []struct {
		name        string
		checkFunc   func(t *testing.T)
		description string
	}{
		{
			name: "all_constants_defined",
			checkFunc: func(t *testing.T) {
				if AppDirName == "" || RuleFileName == "" || LogFileName == "" {
					t.Error("All path constants must be defined")
				}
			},
			description: "All constants should be non-empty",
		},
		{
			name: "constants_are_unique",
			checkFunc: func(t *testing.T) {
				if AppDirName == RuleFileName || AppDirName == LogFileName || RuleFileName == LogFileName {
					t.Error("All constants should have unique values")
				}
			},
			description: "All constants should be distinct from each other",
		},
		{
			name: "can_construct_rule_path",
			checkFunc: func(t *testing.T) {
				// Simulate constructing a typical rule file path
				path := filepath.Join("config", AppDirName, RuleFileName)
				if !strings.Contains(path, AppDirName) || !strings.Contains(path, RuleFileName) {
					t.Error("Should be able to construct valid rule file path")
				}
			},
			description: "Constants should work together to construct valid paths",
		},
		{
			name: "can_construct_log_path",
			checkFunc: func(t *testing.T) {
				// Simulate constructing a typical log file path
				path := filepath.Join("config", AppDirName, LogFileName)
				if !strings.Contains(path, AppDirName) || !strings.Contains(path, LogFileName) {
					t.Error("Should be able to construct valid log file path")
				}
			},
			description: "Constants should work together to construct valid log paths",
		},
		{
			name: "filenames_have_different_extensions",
			checkFunc: func(t *testing.T) {
				ruleExt := filepath.Ext(RuleFileName)
				logExt := filepath.Ext(LogFileName)
				if ruleExt == logExt {
					t.Error("RuleFileName and LogFileName should have different extensions")
				}
			},
			description: "Rule and log files should have different extensions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.checkFunc(t)
		})
	}
}

// TestConstantsImmutability ensures constants maintain expected values
// This test acts as a safety net against accidental changes
func TestConstantsImmutability(t *testing.T) {
	// These values are critical for the application's file structure
	// Any change should be intentional and well-documented
	expectedValues := map[string]string{
		"AppDirName":   "deletor",
		"RuleFileName": "rule.json",
		"LogFileName":  "deletor.log",
	}

	actualValues := map[string]string{
		"AppDirName":   AppDirName,
		"RuleFileName": RuleFileName,
		"LogFileName":  LogFileName,
	}

	for name, expected := range expectedValues {
		t.Run(name, func(t *testing.T) {
			actual := actualValues[name]
			if actual != expected {
				t.Errorf("%s changed from %q to %q - this may break existing installations",
					name, expected, actual)
			}
		})
	}
}
