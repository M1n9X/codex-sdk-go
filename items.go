package codex

import (
	"encoding/json"
	"fmt"
)

// ItemType identifies the kind of thread item.
type ItemType string

const (
	// ItemAgentMessage is a text response from the agent.
	ItemAgentMessage ItemType = "agent_message"
	// ItemReasoning is the agent's reasoning summary.
	ItemReasoning ItemType = "reasoning"
	// ItemCommandExecution is a shell command executed by the agent.
	ItemCommandExecution ItemType = "command_execution"
	// ItemFileChange is a set of file modifications.
	ItemFileChange ItemType = "file_change"
	// ItemMcpToolCall is an MCP tool invocation.
	ItemMcpToolCall ItemType = "mcp_tool_call"
	// ItemWebSearch is a web search request.
	ItemWebSearch ItemType = "web_search"
	// ItemTodoList is the agent's running to-do list.
	ItemTodoList ItemType = "todo_list"
	// ItemError is a non-fatal error surfaced as an item.
	ItemError ItemType = "error"
)

// ThreadItem is the interface implemented by all thread item types.
type ThreadItem interface {
	// itemType returns the type identifier for this item.
	itemType() ItemType
	// GetID returns the unique identifier for this item.
	GetID() string
}

// CommandExecutionStatus represents a command execution state.
type CommandExecutionStatus string

const (
	CommandStatusInProgress CommandExecutionStatus = "in_progress"
	CommandStatusCompleted  CommandExecutionStatus = "completed"
	CommandStatusFailed     CommandExecutionStatus = "failed"
)

// PatchChangeKind indicates the type of file change.
type PatchChangeKind string

const (
	PatchAdd    PatchChangeKind = "add"
	PatchDelete PatchChangeKind = "delete"
	PatchUpdate PatchChangeKind = "update"
)

// PatchApplyStatus indicates the result of applying a patch.
type PatchApplyStatus string

const (
	PatchCompleted PatchApplyStatus = "completed"
	PatchFailed    PatchApplyStatus = "failed"
)

// McpToolCallStatus reflects the state of an MCP tool invocation.
type McpToolCallStatus string

const (
	McpStatusInProgress McpToolCallStatus = "in_progress"
	McpStatusCompleted  McpToolCallStatus = "completed"
	McpStatusFailed     McpToolCallStatus = "failed"
)

// AgentMessageItem contains the assistant's text response.
type AgentMessageItem struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	// Text contains either natural-language text or JSON when structured output is requested.
	Text string `json:"text"`
}

func (i *AgentMessageItem) itemType() ItemType { return ItemAgentMessage }
func (i *AgentMessageItem) GetID() string      { return i.ID }

// ReasoningItem captures the agent's reasoning summary.
type ReasoningItem struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Text string `json:"text"`
}

func (i *ReasoningItem) itemType() ItemType { return ItemReasoning }
func (i *ReasoningItem) GetID() string      { return i.ID }

// CommandExecutionItem records a shell command executed by the agent.
type CommandExecutionItem struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	// Command is the command line executed.
	Command string `json:"command"`
	// AggregatedOutput is the captured stdout and stderr.
	AggregatedOutput string `json:"aggregated_output"`
	// ExitCode is set when the command exits.
	ExitCode *int `json:"exit_code,omitempty"`
	// Status is the current execution status.
	Status CommandExecutionStatus `json:"status"`
}

func (i *CommandExecutionItem) itemType() ItemType { return ItemCommandExecution }
func (i *CommandExecutionItem) GetID() string      { return i.ID }

// FileUpdateChange describes an individual file operation.
type FileUpdateChange struct {
	Path string          `json:"path"`
	Kind PatchChangeKind `json:"kind"`
}

// FileChangeItem aggregates a set of file modifications.
type FileChangeItem struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	// Changes lists individual file operations.
	Changes []FileUpdateChange `json:"changes"`
	// Status indicates whether the patch succeeded or failed.
	Status PatchApplyStatus `json:"status"`
}

func (i *FileChangeItem) itemType() ItemType { return ItemFileChange }
func (i *FileChangeItem) GetID() string      { return i.ID }

// McpContentBlock represents content returned by an MCP tool.
type McpContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
	// Additional fields may be present depending on the content type.
	Data json.RawMessage `json:"data,omitempty"`
}

// McpToolResult contains the result of an MCP tool call.
type McpToolResult struct {
	Content           []McpContentBlock `json:"content"`
	StructuredContent json.RawMessage   `json:"structured_content,omitempty"`
}

// McpToolError contains error information from an MCP tool call.
type McpToolError struct {
	Message string `json:"message"`
}

// McpToolCallItem represents an MCP tool invocation.
type McpToolCallItem struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	// Server is the name of the MCP server handling the request.
	Server string `json:"server"`
	// Tool is the tool invoked on the MCP server.
	Tool string `json:"tool"`
	// Arguments are forwarded to the tool invocation.
	Arguments json.RawMessage `json:"arguments,omitempty"`
	// Result is the payload returned for successful calls.
	Result *McpToolResult `json:"result,omitempty"`
	// Error is the message reported for failed calls.
	Error *McpToolError `json:"error,omitempty"`
	// Status is the current invocation status.
	Status McpToolCallStatus `json:"status"`
}

func (i *McpToolCallItem) itemType() ItemType { return ItemMcpToolCall }
func (i *McpToolCallItem) GetID() string      { return i.ID }

// WebSearchItem captures a web search request.
type WebSearchItem struct {
	ID    string `json:"id"`
	Type  string `json:"type"`
	Query string `json:"query"`
}

func (i *WebSearchItem) itemType() ItemType { return ItemWebSearch }
func (i *WebSearchItem) GetID() string      { return i.ID }

// TodoItem describes a single checklist item.
type TodoItem struct {
	Text      string `json:"text"`
	Completed bool   `json:"completed"`
}

// TodoListItem models the agent's running plan.
type TodoListItem struct {
	ID    string     `json:"id"`
	Type  string     `json:"type"`
	Items []TodoItem `json:"items"`
}

func (i *TodoListItem) itemType() ItemType { return ItemTodoList }
func (i *TodoListItem) GetID() string      { return i.ID }

// ErrorItem reflects a non-fatal error surfaced to the user.
type ErrorItem struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Message string `json:"message"`
}

func (i *ErrorItem) itemType() ItemType { return ItemError }
func (i *ErrorItem) GetID() string      { return i.ID }

// UnknownItem preserves unrecognized item payloads.
type UnknownItem struct {
	ItemType string          `json:"type"`
	Raw      json.RawMessage `json:"-"`
}

func (i *UnknownItem) itemType() ItemType { return ItemType(i.ItemType) }
func (i *UnknownItem) GetID() string      { return "" }

// unmarshalThreadItem decodes a thread item into the corresponding Go type.
func unmarshalThreadItem(data []byte) (ThreadItem, error) {
	var discriminator struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(data, &discriminator); err != nil {
		return nil, err
	}

	switch ItemType(discriminator.Type) {
	case ItemAgentMessage:
		var item AgentMessageItem
		if err := json.Unmarshal(data, &item); err != nil {
			return nil, err
		}
		return &item, nil

	case ItemReasoning:
		var item ReasoningItem
		if err := json.Unmarshal(data, &item); err != nil {
			return nil, err
		}
		return &item, nil

	case ItemCommandExecution:
		var item CommandExecutionItem
		if err := json.Unmarshal(data, &item); err != nil {
			return nil, err
		}
		return &item, nil

	case ItemFileChange:
		var item FileChangeItem
		if err := json.Unmarshal(data, &item); err != nil {
			return nil, err
		}
		return &item, nil

	case ItemMcpToolCall:
		var item McpToolCallItem
		if err := json.Unmarshal(data, &item); err != nil {
			return nil, err
		}
		return &item, nil

	case ItemWebSearch:
		var item WebSearchItem
		if err := json.Unmarshal(data, &item); err != nil {
			return nil, err
		}
		return &item, nil

	case ItemTodoList:
		var item TodoListItem
		if err := json.Unmarshal(data, &item); err != nil {
			return nil, err
		}
		return &item, nil

	case ItemError:
		var item ErrorItem
		if err := json.Unmarshal(data, &item); err != nil {
			return nil, err
		}
		return &item, nil

	case "":
		return nil, fmt.Errorf("thread item missing type discriminator")

	default:
		return &UnknownItem{
			ItemType: discriminator.Type,
			Raw:      json.RawMessage(data),
		}, nil
	}
}
