package validation

import (
	"errors"
	"os"
	"regexp"
	"strings"
)

// Validator provides methods for validating various input parameters
type Validator struct{}

// NewValidator creates a new instance of the Validator
func NewValidator() *Validator {
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

// ValidateTimeDuration checks if a time duration string is in a valid format
// Valid format: number (optional space) followed by time unit (sec, min, hour, day, week, month, year)
// Examples: "7days", "24 hours", "1min", "2 weeks"
func (v *Validator) ValidateTimeDuration(timeStr string) error {
	re := regexp.MustCompile(`^\d+\s*(sec|min|hour|day|week|month|year)s?$`)
	if !re.MatchString(strings.ToLower(timeStr)) {
		return errors.New("expected format: number followed by time unit (sec, min, hour, day, week, month, year)")
	}

	return nil
}
