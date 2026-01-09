package codex

import (
	"fmt"
	"reflect"
	"strings"
)

// Input represents the user-provided content for a single agent turn.
type Input struct {
	prompt string
	parts  []UserInput
}

// Text creates an Input containing a single text prompt.
func Text(prompt string) Input {
	return Input{prompt: prompt}
}

// Compose creates an Input from a set of user input parts.
// Use this when mixing text and images.
func Compose(parts ...UserInput) Input {
	cp := make([]UserInput, len(parts))
	copy(cp, parts)
	return Input{parts: cp}
}

// InputType enumerates the supported user input kinds.
type InputType string

const (
	// InputText represents a text input segment.
	InputText InputType = "text"
	// InputLocalImage represents a local filesystem image.
	InputLocalImage InputType = "local_image"
)

// UserInput captures an individual segment of user-supplied input.
type UserInput struct {
	// Type differentiates the payload stored in other fields.
	Type InputType
	// Text contains the textual prompt for text entries.
	Text string
	// Path contains the local filesystem path for image entries.
	Path string
}

// TextPart creates a text input segment.
func TextPart(text string) UserInput {
	return UserInput{Type: InputText, Text: text}
}

// ImagePart creates a local image input segment.
func ImagePart(path string) UserInput {
	return UserInput{Type: InputLocalImage, Path: path}
}

// normalizeInput converts an Input to prompt string and image paths.
func normalizeInput(input Input) (prompt string, images []string, err error) {
	if len(input.parts) == 0 {
		return input.prompt, nil, nil
	}

	var promptParts []string
	if input.prompt != "" {
		promptParts = append(promptParts, input.prompt)
	}

	for idx, part := range input.parts {
		switch part.Type {
		case InputText:
			promptParts = append(promptParts, part.Text)
		case InputLocalImage:
			if part.Path == "" {
				return "", nil, &ErrInvalidInput{
					Field:  "image path",
					Value:  "",
					Reason: fmt.Sprintf("input part %d: local image path must be set", idx),
				}
			}
			images = append(images, part.Path)
		case "":
			return "", nil, &ErrInvalidInput{
				Field:  "input type",
				Value:  "",
				Reason: fmt.Sprintf("input part %d: type must be set", idx),
			}
		default:
			return "", nil, &ErrInvalidInput{
				Field:  "input type",
				Value:  string(part.Type),
				Reason: fmt.Sprintf("input part %d: unsupported type", idx),
			}
		}
	}

	prompt = strings.Join(promptParts, "\n\n")
	return prompt, images, nil
}

// validateOutputSchema ensures the schema marshals to a JSON object.
func validateOutputSchema(schema any) error {
	if schema == nil {
		return nil
	}

	v := reflect.ValueOf(schema)
	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			// Treat a nil pointer as no schema provided.
			return nil
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Map:
		if v.IsNil() {
			return nil
		}
		if v.Type().Key().Kind() != reflect.String {
			return &ErrInvalidInput{
				Field:  "output schema",
				Reason: "must be a JSON object with string keys",
			}
		}
		return nil
	case reflect.Struct:
		return nil
	case reflect.Interface:
		if v.IsNil() {
			return nil
		}
		return validateOutputSchema(v.Interface())
	default:
		return &ErrInvalidInput{
			Field:  "output schema",
			Reason: "must be a JSON object, not a primitive or array",
		}
	}
}
