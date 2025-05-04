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
	if path == "" && !optional {
		return errors.New("An empty path")
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return errors.New("Path does not exist")
	}

	return nil
}

func (v *Validator) ValidateExtension(ext string) error {
	re := regexp.MustCompile(`^\d+(\.\d+)?(mb|kb|b|gb)$`)
	if !re.MatchString(ext) {
		return errors.New("invalid extension format")
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
