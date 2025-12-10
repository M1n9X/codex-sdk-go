package codex

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// outputSchemaFile manages a temporary file containing the output schema.
type outputSchemaFile struct {
	path    string
	cleanup func() error
}

// Path returns the path to the schema file, or empty if no schema was set.
func (f *outputSchemaFile) Path() string {
	if f == nil {
		return ""
	}
	return f.path
}

// Cleanup removes the temporary schema file.
func (f *outputSchemaFile) Cleanup() error {
	if f == nil || f.cleanup == nil {
		return nil
	}
	return f.cleanup()
}

// createOutputSchemaFile creates a temporary file containing the JSON schema.
// Returns a no-op cleanup if schema is nil.
func createOutputSchemaFile(schema any) (*outputSchemaFile, error) {
	if schema == nil {
		return &outputSchemaFile{
			cleanup: func() error { return nil },
		}, nil
	}

	if err := validateOutputSchema(schema); err != nil {
		return nil, err
	}

	dir, err := os.MkdirTemp("", "codex-output-schema-")
	if err != nil {
		return nil, err
	}

	cleanup := func() error {
		return os.RemoveAll(dir)
	}

	data, err := json.Marshal(schema)
	if err != nil {
		cleanup()
		return nil, err
	}

	path := filepath.Join(dir, "schema.json")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		cleanup()
		return nil, err
	}

	return &outputSchemaFile{
		path:    path,
		cleanup: cleanup,
	}, nil
}
