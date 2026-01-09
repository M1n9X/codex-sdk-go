package codex

import (
	"os"
	"strings"
)

// validateNonEmpty checks if a string is non-empty after trimming whitespace.
// Returns an ErrInvalidInput if the string is empty.
func validateNonEmpty(field, value string) error {
	if strings.TrimSpace(value) == "" {
		return &ErrInvalidInput{
			Field:  field,
			Value:  value,
			Reason: "must not be empty",
		}
	}
	return nil
}

// validatePath checks if a path exists and is accessible.
// Returns an ErrInvalidInput if the path does not exist or is not accessible.
func validatePath(field, path string) error {
	if path == "" {
		return &ErrInvalidInput{
			Field:  field,
			Value:  path,
			Reason: "path must not be empty",
		}
	}

	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return &ErrInvalidInput{
				Field:  field,
				Value:  path,
				Reason: "path does not exist",
			}
		}
		return &ErrInvalidInput{
			Field:  field,
			Value:  path,
			Reason: "path is not accessible: " + err.Error(),
		}
	}

	return nil
}

// validateExecutablePath checks if a path exists and is a regular file.
// Returns an ErrInvalidInput if the path is invalid or a directory.
func validateExecutablePath(field, path string) error {
	if err := validatePath(field, path); err != nil {
		return err
	}

	info, err := os.Stat(path)
	if err != nil {
		return &ErrInvalidInput{
			Field:  field,
			Value:  path,
			Reason: "cannot stat file: " + err.Error(),
		}
	}

	if info.IsDir() {
		return &ErrInvalidInput{
			Field:  field,
			Value:  path,
			Reason: "path is a directory, not a file",
		}
	}

	return nil
}
