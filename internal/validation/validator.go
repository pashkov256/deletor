package validation

import (
	"errors"
	"os"
	"regexp"
)

// Validator provides methods for validating various input parameters
type Validator struct{}

// NewValidator creates a new instance of the Validator
func (v *Validator) NewValidator() *Validator {
	return &Validator{}
}

// ValidatePath checks if a path exists and is valid
// If optional is true, empty paths are allowed
func (v *Validator) ValidatePath(path string, optional bool) error {
	if path == "" {
		if optional {
			return nil
		}
		return errors.New("path cannot be empty")
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return errors.New("path does not exist")
	}

	return nil
}

// ValidateExtension checks if a file extension is valid
// Extensions must contain only alphanumeric characters
func (v *Validator) ValidateExtension(ext string) error {
	if ext == "" {
		return errors.New("extension cannot be empty")
	}

	// Check for invalid characters
	re := regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	if !re.MatchString(ext) {
		return errors.New("extension contains invalid characters")
	}

	return nil
}

// ValidateSize checks if a size string is in a valid format
// Valid format: number followed by unit (e.g., "1.5MB", "2GB")
func (v *Validator) ValidateSize(size string) error {
	re := regexp.MustCompile(`^\d+(\.\d+)?\s*(mb|kb|b|gb)$`)
	if !re.MatchString(size) {
		return errors.New("invalid size format")
	}
	return nil
}
