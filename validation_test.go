package codex

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestValidateNonEmpty(t *testing.T) {
	tests := []struct {
		name      string
		field     string
		value     string
		wantError bool
	}{
		{
			name:      "valid_non_empty",
			field:     "model",
			value:     "gpt-4",
			wantError: false,
		},
		{
			name:      "empty_string",
			field:     "model",
			value:     "",
			wantError: true,
		},
		{
			name:      "whitespace_only",
			field:     "model",
			value:     "   ",
			wantError: true,
		},
		{
			name:      "tabs_and_spaces",
			field:     "model",
			value:     "\t\n  ",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateNonEmpty(tt.field, tt.value)
			if tt.wantError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("expected no error, got: %v", err)
			}

			// Check error type
			if err != nil {
				var invalidInput *ErrInvalidInput
				if !errors.As(err, &invalidInput) {
					t.Errorf("expected ErrInvalidInput, got %T", err)
				}
				if invalidInput.Field != tt.field {
					t.Errorf("expected field %q, got %q", tt.field, invalidInput.Field)
				}
			}
		})
	}
}

func TestValidatePath(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(tmpFile, []byte("test"), 0o644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tests := []struct {
		name      string
		field     string
		path      string
		wantError bool
	}{
		{
			name:      "valid_directory",
			field:     "working_directory",
			path:      tmpDir,
			wantError: false,
		},
		{
			name:      "valid_file",
			field:     "config_file",
			path:      tmpFile,
			wantError: false,
		},
		{
			name:      "empty_path",
			field:     "path",
			path:      "",
			wantError: true,
		},
		{
			name:      "non_existent_path",
			field:     "path",
			path:      "/non/existent/path",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePath(tt.field, tt.path)
			if tt.wantError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("expected no error, got: %v", err)
			}

			// Check error type
			if err != nil {
				var invalidInput *ErrInvalidInput
				if !errors.As(err, &invalidInput) {
					t.Errorf("expected ErrInvalidInput, got %T", err)
				}
			}
		})
	}
}

func TestValidateExecutablePath(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tests := []struct {
		name      string
		field     string
		path      string
		wantError bool
	}{
		{
			name:      "valid_file",
			field:     "codex_path",
			path:      tmpFile,
			wantError: false,
		},
		{
			name:      "directory_not_file",
			field:     "codex_path",
			path:      tmpDir,
			wantError: true,
		},
		{
			name:      "empty_path",
			field:     "codex_path",
			path:      "",
			wantError: true,
		},
		{
			name:      "non_existent_path",
			field:     "codex_path",
			path:      "/non/existent/file",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateExecutablePath(tt.field, tt.path)
			if tt.wantError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("expected no error, got: %v", err)
			}

			// Check error type
			if err != nil {
				var invalidInput *ErrInvalidInput
				if !errors.As(err, &invalidInput) {
					t.Errorf("expected ErrInvalidInput, got %T", err)
				}
			}
		})
	}
}

func TestValidationErrorMessages(t *testing.T) {
	// Test that error messages are descriptive
	err := validateNonEmpty("model", "")
	if err == nil {
		t.Fatal("expected error")
	}
	if !contains(err.Error(), "model") {
		t.Errorf("error message should mention field name: %v", err)
	}
	if !contains(err.Error(), "empty") {
		t.Errorf("error message should mention reason: %v", err)
	}

	err = validatePath("path", "/non/existent")
	if err == nil {
		t.Fatal("expected error")
	}
	if !contains(err.Error(), "path") {
		t.Errorf("error message should mention field name: %v", err)
	}
	if !contains(err.Error(), "exist") {
		t.Errorf("error message should mention reason: %v", err)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
