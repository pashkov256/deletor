package validation

import (
	"os"
	"path/filepath"
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
	validator := new(validation.Validator).NewValidator()

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
	validator := new(validation.Validator).NewValidator()
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
	validator := new(validation.Validator).NewValidator()

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
