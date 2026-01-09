package codex

import (
	"errors"
	"fmt"
)

// ErrCodexNotFound is returned when the codex binary cannot be found.
var ErrCodexNotFound = errors.New("codex binary not found in PATH or bundled location")

// ErrInvalidInput represents an error caused by invalid user input.
type ErrInvalidInput struct {
	// Field is the name of the field that failed validation.
	Field string
	// Value is the invalid value provided.
	Value string
	// Reason explains why the value is invalid.
	Reason string
}

// Error implements the error interface.
func (e *ErrInvalidInput) Error() string {
	if e.Value == "" {
		return fmt.Sprintf("invalid %s: %s", e.Field, e.Reason)
	}
	return fmt.Sprintf("invalid %s %q: %s", e.Field, e.Value, e.Reason)
}

// ErrExecFailed represents an error from the codex CLI execution.
type ErrExecFailed struct {
	// ExitCode is the process exit code.
	ExitCode int
	// Stderr contains the stderr output from the process.
	Stderr string
	// Err is the underlying error, if any.
	Err error
}

// Error implements the error interface.
func (e *ErrExecFailed) Error() string {
	if e.Stderr != "" {
		return fmt.Sprintf("codex exec exited with code %d: %s", e.ExitCode, e.Stderr)
	}
	return fmt.Sprintf("codex exec exited with code %d", e.ExitCode)
}

// Unwrap returns the underlying error.
func (e *ErrExecFailed) Unwrap() error {
	return e.Err
}
