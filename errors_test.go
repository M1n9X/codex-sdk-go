package codex

import (
	"errors"
	"strings"
	"testing"
)

func TestErrCodexNotFound(t *testing.T) {
	err := ErrCodexNotFound
	if err == nil {
		t.Fatal("ErrCodexNotFound should not be nil")
	}

	expected := "codex binary not found"
	if !strings.Contains(err.Error(), expected) {
		t.Errorf("expected error message to contain %q, got %q", expected, err.Error())
	}

	// Test that it can be compared with errors.Is
	if !errors.Is(err, ErrCodexNotFound) {
		t.Error("errors.Is should return true for ErrCodexNotFound")
	}
}

func TestErrInvalidInput(t *testing.T) {
	tests := []struct {
		name     string
		err      *ErrInvalidInput
		expected string
	}{
		{
			name: "with_value",
			err: &ErrInvalidInput{
				Field:  "model",
				Value:  "invalid-model",
				Reason: "unsupported model",
			},
			expected: `invalid model "invalid-model": unsupported model`,
		},
		{
			name: "without_value",
			err: &ErrInvalidInput{
				Field:  "path",
				Value:  "",
				Reason: "must not be empty",
			},
			expected: "invalid path: must not be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.expected {
				t.Errorf("expected error message %q, got %q", tt.expected, tt.err.Error())
			}

			// Test that it implements error interface
			var _ error = tt.err
		})
	}
}

func TestErrExecFailed(t *testing.T) {
	tests := []struct {
		name     string
		err      *ErrExecFailed
		expected string
	}{
		{
			name: "with_stderr",
			err: &ErrExecFailed{
				ExitCode: 1,
				Stderr:   "command not found",
			},
			expected: "codex exec exited with code 1: command not found",
		},
		{
			name: "without_stderr",
			err: &ErrExecFailed{
				ExitCode: 2,
				Stderr:   "",
			},
			expected: "codex exec exited with code 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.expected {
				t.Errorf("expected error message %q, got %q", tt.expected, tt.err.Error())
			}

			// Test that it implements error interface
			var _ error = tt.err
		})
	}
}

func TestErrExecFailed_Unwrap(t *testing.T) {
	underlyingErr := errors.New("underlying error")
	execErr := &ErrExecFailed{
		ExitCode: 1,
		Stderr:   "test error",
		Err:      underlyingErr,
	}

	// Test Unwrap
	unwrapped := execErr.Unwrap()
	if unwrapped != underlyingErr {
		t.Errorf("expected unwrapped error to be %v, got %v", underlyingErr, unwrapped)
	}

	// Test errors.Is
	if !errors.Is(execErr, underlyingErr) {
		t.Error("errors.Is should return true for underlying error")
	}
}

func TestErrorChaining(t *testing.T) {
	// Test that custom errors can be used with error wrapping
	baseErr := errors.New("base error")
	invalidInput := &ErrInvalidInput{
		Field:  "test",
		Reason: "test reason",
	}

	// Wrap with custom error
	wrapped := errors.Join(baseErr, invalidInput)
	if wrapped == nil {
		t.Fatal("wrapped error should not be nil")
	}

	// Check that both errors are in the chain
	if !errors.Is(wrapped, baseErr) {
		t.Error("errors.Is should find base error in chain")
	}
	if !errors.Is(wrapped, invalidInput) {
		t.Error("errors.Is should find custom error in chain")
	}
}
