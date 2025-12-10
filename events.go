package codex

import (
	"encoding/json"
	"fmt"
)

// EventType enumerates the JSON events emitted by codex exec.
type EventType string

const (
	// EventThreadStarted is emitted when a new thread is started.
	EventThreadStarted EventType = "thread.started"
	// EventTurnStarted is emitted when a turn begins processing.
	EventTurnStarted EventType = "turn.started"
	// EventTurnCompleted is emitted when a turn finishes successfully.
	EventTurnCompleted EventType = "turn.completed"
	// EventTurnFailed is emitted when a turn fails with an error.
	EventTurnFailed EventType = "turn.failed"
	// EventItemStarted is emitted when a new item is added to the thread.
	EventItemStarted EventType = "item.started"
	// EventItemUpdated is emitted when an item is updated.
	EventItemUpdated EventType = "item.updated"
	// EventItemCompleted is emitted when an item reaches a terminal state.
	EventItemCompleted EventType = "item.completed"
	// EventError is emitted for fatal stream errors.
	EventError EventType = "error"
)

// Usage reports token usage for a turn.
type Usage struct {
	// InputTokens is the number of input tokens used.
	InputTokens int `json:"input_tokens"`
	// CachedInputTokens is the number of cached input tokens used.
	CachedInputTokens int `json:"cached_input_tokens"`
	// OutputTokens is the number of output tokens generated.
	OutputTokens int `json:"output_tokens"`
}

// ThreadError describes a fatal error emitted by a turn.
type ThreadError struct {
	// Message contains the error description.
	Message string `json:"message"`
}

// ThreadEvent represents a single line event emitted by codex exec.
type ThreadEvent struct {
	// Type identifies the event kind.
	Type EventType `json:"type"`
	// ThreadID is populated on thread.started events.
	ThreadID string `json:"thread_id,omitempty"`
	// Usage is populated on turn.completed events.
	Usage *Usage `json:"usage,omitempty"`
	// Error is populated on turn.failed events.
	Error *ThreadError `json:"error,omitempty"`
	// Item contains the thread item for item.* events.
	Item ThreadItem `json:"-"`
	// Message is populated on top-level error events.
	Message string `json:"message,omitempty"`

	// rawItem holds the raw JSON for deferred item parsing.
	rawItem json.RawMessage
}

// UnmarshalJSON customizes decoding to handle the polymorphic item payload.
func (e *ThreadEvent) UnmarshalJSON(data []byte) error {
	type eventAlias ThreadEvent
	var aux struct {
		eventAlias
		Item json.RawMessage `json:"item,omitempty"`
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	*e = ThreadEvent(aux.eventAlias)
	e.rawItem = aux.Item

	if len(aux.Item) > 0 {
		item, err := unmarshalThreadItem(aux.Item)
		if err != nil {
			return fmt.Errorf("decode thread item: %w", err)
		}
		e.Item = item
	}

	return nil
}

// String returns a human-readable representation of the event.
func (e ThreadEvent) String() string {
	switch e.Type {
	case EventThreadStarted:
		if e.ThreadID != "" {
			return fmt.Sprintf("thread.started id=%s", e.ThreadID)
		}
		return "thread.started"
	case EventTurnStarted:
		return "turn.started"
	case EventTurnCompleted:
		if e.Usage != nil {
			return fmt.Sprintf("turn.completed usage={input=%d cached=%d output=%d}",
				e.Usage.InputTokens, e.Usage.CachedInputTokens, e.Usage.OutputTokens)
		}
		return "turn.completed"
	case EventTurnFailed:
		if e.Error != nil {
			return fmt.Sprintf("turn.failed error=%s", e.Error.Message)
		}
		return "turn.failed"
	case EventItemStarted, EventItemUpdated, EventItemCompleted:
		if e.Item != nil {
			return fmt.Sprintf("%s item=%s", e.Type, itemSummary(e.Item))
		}
		return string(e.Type)
	case EventError:
		if e.Message != "" {
			return fmt.Sprintf("error message=%s", e.Message)
		}
		return "error"
	default:
		return string(e.Type)
	}
}

func itemSummary(item ThreadItem) string {
	switch v := item.(type) {
	case *AgentMessageItem:
		text := v.Text
		if len(text) > 50 {
			text = text[:50] + "..."
		}
		return fmt.Sprintf("agent_message text=%q", text)
	case *ReasoningItem:
		text := v.Text
		if len(text) > 50 {
			text = text[:50] + "..."
		}
		return fmt.Sprintf("reasoning text=%q", text)
	case *CommandExecutionItem:
		return fmt.Sprintf("command_execution command=%q status=%s", v.Command, v.Status)
	case *FileChangeItem:
		return fmt.Sprintf("file_change changes=%d status=%s", len(v.Changes), v.Status)
	case *McpToolCallItem:
		return fmt.Sprintf("mcp_tool_call server=%q tool=%q status=%s", v.Server, v.Tool, v.Status)
	case *WebSearchItem:
		return fmt.Sprintf("web_search query=%q", v.Query)
	case *TodoListItem:
		return fmt.Sprintf("todo_list items=%d", len(v.Items))
	case *ErrorItem:
		return fmt.Sprintf("error message=%q", v.Message)
	case *UnknownItem:
		return fmt.Sprintf("unknown type=%s", v.ItemType)
	default:
		return fmt.Sprintf("%T", item)
	}
}
