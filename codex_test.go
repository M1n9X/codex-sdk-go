package codex

import (
	"encoding/json"
	"testing"
)

func TestNormalizeInput_TextOnly(t *testing.T) {
	input := Text("Hello, world!")
	prompt, images, err := normalizeInput(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if prompt != "Hello, world!" {
		t.Errorf("expected prompt %q, got %q", "Hello, world!", prompt)
	}
	if len(images) != 0 {
		t.Errorf("expected 0 images, got %d", len(images))
	}
}

func TestNormalizeInput_Compose(t *testing.T) {
	input := Compose(
		TextPart("First part"),
		TextPart("Second part"),
		ImagePart("/path/to/image.png"),
	)
	prompt, images, err := normalizeInput(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "First part\n\nSecond part"
	if prompt != expected {
		t.Errorf("expected prompt %q, got %q", expected, prompt)
	}
	if len(images) != 1 {
		t.Errorf("expected 1 image, got %d", len(images))
	}
	if images[0] != "/path/to/image.png" {
		t.Errorf("expected image path %q, got %q", "/path/to/image.png", images[0])
	}
}

func TestNormalizeInput_EmptyImagePath(t *testing.T) {
	input := Compose(
		TextPart("Text"),
		ImagePart(""),
	)
	_, _, err := normalizeInput(input)
	if err == nil {
		t.Fatal("expected error for empty image path")
	}
}

func TestNormalizeInput_MissingType(t *testing.T) {
	input := Compose(
		UserInput{}, // No type set
	)
	_, _, err := normalizeInput(input)
	if err == nil {
		t.Fatal("expected error for missing type")
	}
}

func TestValidateOutputSchema(t *testing.T) {
	// Valid schema
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{"type": "string"},
		},
	}
	if err := validateOutputSchema(schema); err != nil {
		t.Errorf("expected no error for valid schema, got: %v", err)
	}

	// Structs and pointers to structs are allowed
	type person struct {
		Name string `json:"name"`
	}
	if err := validateOutputSchema(person{Name: "alice"}); err != nil {
		t.Errorf("expected no error for struct schema, got: %v", err)
	}
	if err := validateOutputSchema(&person{Name: "bob"}); err != nil {
		t.Errorf("expected no error for struct pointer schema, got: %v", err)
	}
	var nilPerson *person
	if err := validateOutputSchema(nilPerson); err != nil {
		t.Errorf("expected no error for nil struct pointer schema, got: %v", err)
	}

	// Nil schema is valid
	if err := validateOutputSchema(nil); err != nil {
		t.Errorf("expected no error for nil schema, got: %v", err)
	}

	invalidSchemas := []struct {
		name   string
		schema any
	}{
		{name: "string", schema: "string"},
		{name: "int", schema: 123},
		{name: "array_any", schema: []any{}},
		{name: "array_string", schema: []string{"a"}},
		{name: "map_non_string_key", schema: map[int]string{1: "a"}},
	}

	for _, tc := range invalidSchemas {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if err := validateOutputSchema(tc.schema); err == nil {
				t.Fatalf("expected error for %s schema", tc.name)
			}
		})
	}
}

func TestTextPartAndImagePart(t *testing.T) {
	textPart := TextPart("hello")
	if textPart.Type != InputText {
		t.Errorf("expected type %q, got %q", InputText, textPart.Type)
	}
	if textPart.Text != "hello" {
		t.Errorf("expected text %q, got %q", "hello", textPart.Text)
	}

	imagePart := ImagePart("/path/to/img.jpg")
	if imagePart.Type != InputLocalImage {
		t.Errorf("expected type %q, got %q", InputLocalImage, imagePart.Type)
	}
	if imagePart.Path != "/path/to/img.jpg" {
		t.Errorf("expected path %q, got %q", "/path/to/img.jpg", imagePart.Path)
	}
}

func TestThreadEventUnmarshal(t *testing.T) {
	// Test thread.started event
	data := `{"type":"thread.started","thread_id":"test-123"}`
	var event ThreadEvent
	if err := json.Unmarshal([]byte(data), &event); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if event.Type != EventThreadStarted {
		t.Errorf("expected type %q, got %q", EventThreadStarted, event.Type)
	}
	if event.ThreadID != "test-123" {
		t.Errorf("expected thread_id %q, got %q", "test-123", event.ThreadID)
	}

	// Test turn.completed event with usage
	data = `{"type":"turn.completed","usage":{"input_tokens":100,"cached_input_tokens":20,"output_tokens":50}}`
	if err := json.Unmarshal([]byte(data), &event); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if event.Type != EventTurnCompleted {
		t.Errorf("expected type %q, got %q", EventTurnCompleted, event.Type)
	}
	if event.Usage == nil {
		t.Fatal("expected usage to be set")
	}
	if event.Usage.InputTokens != 100 {
		t.Errorf("expected input_tokens 100, got %d", event.Usage.InputTokens)
	}

	// Test item.completed with agent_message
	data = `{"type":"item.completed","item":{"id":"item-1","type":"agent_message","text":"Hello!"}}`
	if err := json.Unmarshal([]byte(data), &event); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if event.Type != EventItemCompleted {
		t.Errorf("expected type %q, got %q", EventItemCompleted, event.Type)
	}
	if event.Item == nil {
		t.Fatal("expected item to be set")
	}
	msg, ok := event.Item.(*AgentMessageItem)
	if !ok {
		t.Fatalf("expected *AgentMessageItem, got %T", event.Item)
	}
	if msg.Text != "Hello!" {
		t.Errorf("expected text %q, got %q", "Hello!", msg.Text)
	}
}

func TestUnmarshalThreadItem(t *testing.T) {
	tests := []struct {
		name     string
		data     string
		itemType ItemType
	}{
		{"agent_message", `{"id":"1","type":"agent_message","text":"hi"}`, ItemAgentMessage},
		{"reasoning", `{"id":"2","type":"reasoning","text":"thinking"}`, ItemReasoning},
		{"command_execution", `{"id":"3","type":"command_execution","command":"ls","status":"completed"}`, ItemCommandExecution},
		{"file_change", `{"id":"4","type":"file_change","changes":[],"status":"completed"}`, ItemFileChange},
		{"mcp_tool_call", `{"id":"5","type":"mcp_tool_call","server":"s","tool":"t","status":"completed"}`, ItemMcpToolCall},
		{"web_search", `{"id":"6","type":"web_search","query":"test"}`, ItemWebSearch},
		{"todo_list", `{"id":"7","type":"todo_list","items":[]}`, ItemTodoList},
		{"error", `{"id":"8","type":"error","message":"oops"}`, ItemError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item, err := unmarshalThreadItem([]byte(tt.data))
			if err != nil {
				t.Fatalf("unmarshal failed: %v", err)
			}
			if item.itemType() != tt.itemType {
				t.Errorf("expected item type %q, got %q", tt.itemType, item.itemType())
			}
		})
	}

	// Test unknown type
	t.Run("unknown", func(t *testing.T) {
		item, err := unmarshalThreadItem([]byte(`{"type":"future_type","data":"test"}`))
		if err != nil {
			t.Fatalf("unmarshal failed: %v", err)
		}
		unknown, ok := item.(*UnknownItem)
		if !ok {
			t.Fatalf("expected *UnknownItem, got %T", item)
		}
		if unknown.ItemType != "future_type" {
			t.Errorf("expected item type %q, got %q", "future_type", unknown.ItemType)
		}
	})
}

func TestOptionsApply(t *testing.T) {
	// Test CodexOptions
	opts := applyCodexOptions([]Option{
		WithAPIKey("test-key"),
		WithBaseURL("https://test.com"),
		WithCodexPath("/custom/codex"),
		WithEnv(map[string]string{"FOO": "bar"}),
	})
	if opts.APIKey != "test-key" {
		t.Errorf("expected APIKey %q, got %q", "test-key", opts.APIKey)
	}
	if opts.BaseURL != "https://test.com" {
		t.Errorf("expected BaseURL %q, got %q", "https://test.com", opts.BaseURL)
	}
	if opts.CodexPath != "/custom/codex" {
		t.Errorf("expected CodexPath %q, got %q", "/custom/codex", opts.CodexPath)
	}
	if opts.Env["FOO"] != "bar" {
		t.Errorf("expected Env[FOO] %q, got %q", "bar", opts.Env["FOO"])
	}

	// Test ThreadOptions
	topts := applyThreadOptions([]ThreadOption{
		WithModel("gpt-4"),
		WithSandboxMode(SandboxWorkspaceWrite),
		WithWorkingDirectory("/work"),
		WithSkipGitRepoCheck(),
		WithModelReasoningEffort(ReasoningXHigh),
		WithNetworkAccess(true),
		WithWebSearch(false),
		WithApprovalPolicy(ApprovalOnRequest),
		WithAdditionalDirectories("/dir1", "/dir2"),
	})
	if topts.Model != "gpt-4" {
		t.Errorf("expected Model %q, got %q", "gpt-4", topts.Model)
	}
	if topts.SandboxMode != SandboxWorkspaceWrite {
		t.Errorf("expected SandboxMode %q, got %q", SandboxWorkspaceWrite, topts.SandboxMode)
	}
	if topts.WorkingDirectory != "/work" {
		t.Errorf("expected WorkingDirectory %q, got %q", "/work", topts.WorkingDirectory)
	}
	if !topts.SkipGitRepoCheck {
		t.Error("expected SkipGitRepoCheck to be true")
	}
	if topts.ModelReasoningEffort != ReasoningXHigh {
		t.Errorf("expected ModelReasoningEffort %q, got %q", ReasoningXHigh, topts.ModelReasoningEffort)
	}
	if topts.NetworkAccessEnabled == nil || !*topts.NetworkAccessEnabled {
		t.Error("expected NetworkAccessEnabled to be true")
	}
	if topts.WebSearchEnabled == nil || *topts.WebSearchEnabled {
		t.Error("expected WebSearchEnabled to be false")
	}
	if topts.ApprovalPolicy != ApprovalOnRequest {
		t.Errorf("expected ApprovalPolicy %q, got %q", ApprovalOnRequest, topts.ApprovalPolicy)
	}
	if len(topts.AdditionalDirectories) != 2 {
		t.Errorf("expected 2 additional directories, got %d", len(topts.AdditionalDirectories))
	}

	// Test TurnOptions
	turnOpts := applyTurnOptions([]TurnOption{
		WithOutputSchema(map[string]any{"type": "object"}),
	})
	if turnOpts.OutputSchema == nil {
		t.Error("expected OutputSchema to be set")
	}
}

func TestTypeAliases(t *testing.T) {
	// Verify RunResult is an alias for Turn
	var turn Turn
	var runResult RunResult = turn
	_ = runResult

	// Verify RunStreamedResult is an alias for StreamedTurn
	// Note: We use pointers here to avoid copying the sync.Once lock
	var streamedTurn *StreamedTurn
	var runStreamedResult *RunStreamedResult = streamedTurn
	_ = runStreamedResult
}
