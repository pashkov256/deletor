package validation

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pashkov256/deletor/internal/validation"
)

func setupTestDir(t *testing.T) string {
	dir := filepath.Join(os.TempDir(), "deletor_test")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	return dir
}

func cleanupTestDir(t *testing.T, dir string) {
	if err := os.RemoveAll(dir); err != nil {
		t.Errorf("Failed to clean up test directory: %v", err)
	}
}

func TestValidator_ValidateSize(t *testing.T) {
	validator := validation.NewValidator()

	tests := []struct {
		name    string
		size    string
		wantErr bool
	}{
		// Valid cases
		{"valid size with space", "10 mb", false},
		{"valid size without space", "10mb", false},
		{"valid size with decimal", "10.5 mb", false},
		{"valid size with GB", "1 gb", false},
		{"valid size with KB", "100 kb", false},
		{"valid size with B", "1024 b", false},

		// Invalid cases
		{"invalid format", "10m", true},
		{"invalid unit", "10 tb", true},
		{"empty size", "", true},
		{"negative size", "-10 mb", true},
		{"invalid decimal", "10.5.5 mb", true},
		{"no number", "mb", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateSize(tt.size)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSize(%q) error = %v, wantErr %v", tt.size, err, tt.wantErr)
			}
		})
	}
}

func TestValidator_ValidatePath(t *testing.T) {
	validator := validation.NewValidator()
	testDir := setupTestDir(t)
	defer cleanupTestDir(t, testDir)

	tests := []struct {
		name     string
		path     string
		optional bool
		wantErr  bool
	}{
		// Valid cases
		{"valid path", testDir, false, false},
		{"empty path optional", "", true, false},

		// Invalid cases
		{"empty path not optional", "", false, true},
		{"non-existent path", "/nonexistent/path", false, true},
		{"invalid path format", "invalid/path/format", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidatePath(tt.path, tt.optional)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePath(%q, %v) error = %v, wantErr %v",
					tt.path, tt.optional, err, tt.wantErr)
			}
		})
	}
}

func TestValidator_ValidateExtension(t *testing.T) {
	validator := validation.NewValidator()

	tests := []struct {
		name    string
		ext     string
		wantErr bool
	}{
		// Valid cases
		{"valid extension", "png", false},
		{"valid extension uppercase", "PNG", false},
		{"valid extension mixed case", "Png", false},
		{"valid extension with numbers", "mp4", false},

		// Invalid cases
		{"invalid extension with dot", ".png", true},
		{"invalid extension with space", "p n g", true},
		{"empty extension", "", true},
		{"invalid characters", "png!", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateExtension(tt.ext)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateExtension(%q) error = %v, wantErr %v",
					tt.ext, err, tt.wantErr)
			}
		})
	}
}

func TestValidateTimeDuration(t *testing.T) {
	errString := "expected format: number followed by time unit (sec, min, hour, day, week, month, year)"
	tests := []struct {
		name     string
		input    string
		expected error
	}{
		// Valid cases
		{"Valid seconds singular", "1 sec", nil},
		{"Valid seconds plural", "10 secs", nil},
		{"Valid minutes singular", "1 min", nil},
		{"Valid minutes plural", "30 mins", nil},
		{"Valid hours singular", "1 hour", nil},
		{"Valid hours plural", "24 hours", nil},
		{"Valid days singular", "1 day", nil},
		{"Valid days plural", "7 days", nil},
		{"Valid weeks singular", "1 week", nil},
		{"Valid weeks plural", "4 weeks", nil},
		{"Valid months singular", "1 month", nil},
		{"Valid months plural", "12 months", nil},
		{"Valid years singular", "1 year", nil},
		{"Valid years plural", "5 years", nil},
		{"Valid with space", "5  sec", nil},
		{"Valid with multiple spaces", "10   secs", nil},
		{"Valid uppercase units", "1 WEEK", nil},
		{"Valid mixed case units", "1 mOnTh", nil},
		{"Valid large number", "9999999999999 years", nil},
		{"Valid no space", "5sec", nil},
		{"Valid minimal space", "5 sec", nil},
		{"Valid extra space", "5  sec", nil},
		{"Newline character", "10\nsec", nil},
		{"Tab character", "10\tsec", nil},
		{"Unit without s", "2 year", nil},
		{"Unit with extra s", "1 years", nil},

		// Invalid cases
		{"Empty string", "", errors.New(errString)},
		{"Only spaces", "   ", errors.New(errString)},
		{"Missing number", "sec", errors.New(errString)},
		{"Missing unit", "10", errors.New(errString)},
		{"Invalid unit", "10 apples", errors.New(errString)},
		{"Negative number", "-5 sec", errors.New(errString)},
		{"Decimal number", "5.5 sec", errors.New(errString)},
		{"Trailing characters", "10 sec!", errors.New(errString)},
		{"Leading characters", "about 10 sec", errors.New(errString)},
		{"Multiple numbers", "10 20 sec", errors.New(errString)},
		{"Invalid plural form", "1 secss", errors.New(errString)},
		{"Invalid time unit", "10 lightyears", errors.New(errString)},
		{"Special characters", "#$% sec", errors.New(errString)},
		{"Scientific notation", "1e5 sec", errors.New(errString)},
		{"Comma separated", "1,000 sec", errors.New(errString)},
	}

	v := validation.NewValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateTimeDuration(tt.input)
			if tt.expected == nil {
				if err != nil {
					t.Errorf("Input: %s", tt.input)
				}
			} else {
				if err != nil {
					if !strings.Contains(err.Error(), tt.expected.Error()) {
						t.Errorf("ValidateTimeDuration(%q) error = %v, want containing %v", tt.input, err, tt.expected)
					}
				} else {
					t.Errorf("ValidateTimeDuration(%q) expected error but got nil", tt.input)
				}
			}
		})
	}
}
