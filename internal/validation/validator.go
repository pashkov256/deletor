package validation

import (
	"errors"
	"os"
	"regexp"
)

type Validator struct{}

func (v *Validator) NewValidator() *Validator {
	return &Validator{}
}

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

func (v *Validator) ValidateSize(size string) error {
	re := regexp.MustCompile(`^\d+(\.\d+)?\s*(mb|kb|b|gb)$`)
	if !re.MatchString(size) {
		return errors.New("invalid size format")
	}
	return nil
}
