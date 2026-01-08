# Codex SDK for Go

Embed the Codex agent in your Go workflows and applications.

This SDK mirrors the official [OpenAI TypeScript Codex SDK](https://github.com/openai/codex/tree/main/sdk/typescript), allowing developers to integrate Codex capabilities into Go applications. The SDK wraps the bundled `codex` binary and exchanges JSONL events over stdin/stdout.

## Installation

```bash
go get github.com/M1n9X/codex-sdk-go
```

### Prerequisites

- Go 1.22 or later
- Codex CLI binary: the SDK first looks for a bundled binary under `vendor/<triple>/codex/` (for example `vendor/aarch64-apple-darwin/codex/codex`). If none is present, it falls back to `codex` on `PATH`.

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/M1n9X/codex-sdk-go"
)

func main() {
    client, err := codex.New()
    if err != nil {
        log.Fatal(err)
    }

    thread := client.StartThread()
    turn, err := thread.Run(context.Background(), codex.Text("Diagnose the test failure and propose a fix"))
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(turn.FinalResponse)
    fmt.Println(turn.Items)
}
```

Call `Run()` repeatedly on the same `Thread` instance to continue that conversation:

```go
nextTurn, err := thread.Run(ctx, codex.Text("Implement the fix"))
```

## Streaming Responses

`Run()` buffers events until the turn finishes. To react to intermediate progress—tool calls, streaming responses, and file change notifications—use `RunStreamed()` instead:

```go
streamed, err := thread.RunStreamed(ctx, codex.Text("Diagnose the test failure and propose a fix"))
if err != nil {
    log.Fatal(err)
}

for event := range streamed.Events {
    switch event.Type {
    case codex.EventItemCompleted:
        fmt.Println("item:", event.Item)
    case codex.EventTurnCompleted:
        fmt.Println("usage:", event.Usage)
    }
}

if err := streamed.Wait(); err != nil {
    log.Fatal(err)
}
```

## Structured Output

The Codex agent can produce a JSON response that conforms to a specified schema:

```go
schema := map[string]any{
    "type": "object",
    "properties": map[string]any{
        "summary": map[string]any{"type": "string"},
        "status":  map[string]any{"type": "string", "enum": []string{"ok", "action_required"}},
    },
    "required":             []string{"summary", "status"},
    "additionalProperties": false,
}

turn, err := thread.Run(ctx, codex.Text("Summarize repository status"),
    codex.WithOutputSchema(schema))
if err != nil {
    log.Fatal(err)
}

fmt.Println(turn.FinalResponse) // JSON conforming to schema
```

To mirror the TypeScript example that derives a schema from Zod, you can generate a
JSON Schema from Go structs using [`github.com/invopop/jsonschema`](https://github.com/invopop/jsonschema):

```go
type RepoStatus struct {
    Summary string `json:"summary"`
    Status  string `json:"status" jsonschema:"enum=ok,enum=action_required"`
}

schema := (&jsonschema.Reflector{
    RequiredFromJSONSchemaTags: true,
}).Reflect(&RepoStatus{})

turn, err := thread.Run(ctx, codex.Text("Summarize repository status"),
    codex.WithOutputSchema(schema))
```

## Attaching Images

Provide structured input when you need to include images alongside text:

```go
turn, err := thread.Run(ctx, codex.Compose(
    codex.TextPart("Describe these screenshots"),
    codex.ImagePart("./ui.png"),
    codex.ImagePart("./diagram.jpg"),
))
```

## Resuming an Existing Thread

Threads are persisted in `~/.codex/sessions`. If you lose the in-memory `Thread` object, reconstruct it with `ResumeThread()`:

```go
savedThreadID := os.Getenv("CODEX_THREAD_ID")
thread := client.ResumeThread(savedThreadID)
turn, err := thread.Run(ctx, codex.Text("Implement the fix"))
```

## Working Directory Controls

Codex runs in the current working directory by default. To avoid unrecoverable errors, Codex requires the working directory to be a Git repository. You can skip the Git repository check:

```go
thread := client.StartThread(
    codex.WithWorkingDirectory("/path/to/project"),
    codex.WithSkipGitRepoCheck(),
)
```

## Thread Options

Configure threads with various options:

```go
thread := client.StartThread(
    codex.WithModel("gpt-4"),
    codex.WithSandboxMode(codex.SandboxWorkspaceWrite),
    codex.WithWorkingDirectory("/path/to/project"),
    codex.WithSkipGitRepoCheck(),
    codex.WithModelReasoningEffort(codex.ReasoningXHigh),
    codex.WithNetworkAccess(true),
    codex.WithWebSearch(true),
    codex.WithApprovalPolicy(codex.ApprovalOnRequest),
    codex.WithAdditionalDirectories("../backend", "/tmp/shared"),
)
```

### Available Options

| Option | Description |
|--------|-------------|
| `WithModel(model)` | Select the model identifier |
| `WithSandboxMode(mode)` | Control filesystem access (`SandboxReadOnly`, `SandboxWorkspaceWrite`, `SandboxDangerFullAccess`) |
| `WithWorkingDirectory(dir)` | Set the working directory |
| `WithSkipGitRepoCheck()` | Skip Git repository validation |
| `WithModelReasoningEffort(effort)` | Set reasoning intensity (`ReasoningMinimal`, `ReasoningLow`, `ReasoningMedium`, `ReasoningHigh`, `ReasoningXHigh`) |
| `WithNetworkAccess(enabled)` | Enable/disable network access |
| `WithWebSearch(enabled)` | Enable/disable web search |
| `WithApprovalPolicy(policy)` | Set approval mode (`ApprovalNever`, `ApprovalOnRequest`, `ApprovalOnFailure`, `ApprovalUntrusted`) |
| `WithAdditionalDirectories(dirs...)` | Add accessible directories |

## Client Options

Configure the Codex client:

```go
client, err := codex.New(
    codex.WithAPIKey("sk-..."),
    codex.WithBaseURL("https://api.example.com"),
    codex.WithCodexPath("/custom/path/to/codex"),
    codex.WithEnv(map[string]string{
        "PATH": "/usr/local/bin",
    }),
)
```

## Event Types

The SDK emits the following event types during streaming:

| Event Type | Description |
|------------|-------------|
| `EventThreadStarted` | New thread started |
| `EventTurnStarted` | Turn processing began |
| `EventTurnCompleted` | Turn finished successfully |
| `EventTurnFailed` | Turn failed with error |
| `EventItemStarted` | New item added |
| `EventItemUpdated` | Item updated |
| `EventItemCompleted` | Item reached terminal state |
| `EventError` | Fatal stream error |

## Item Types

Thread items represent different agent actions:

| Item Type | Go Type | Description |
|-----------|---------|-------------|
| `agent_message` | `*AgentMessageItem` | Text response from agent |
| `reasoning` | `*ReasoningItem` | Agent's reasoning summary |
| `command_execution` | `*CommandExecutionItem` | Shell command executed |
| `file_change` | `*FileChangeItem` | File modifications |
| `mcp_tool_call` | `*McpToolCallItem` | MCP tool invocation |
| `web_search` | `*WebSearchItem` | Web search request |
| `todo_list` | `*TodoListItem` | Agent's running plan |
| `error` | `*ErrorItem` | Non-fatal error |

## Examples

See the [examples](./examples) directory for complete working examples:

- [basic_streaming](./examples/basic_streaming) - Interactive streaming chat
- [structured_output](./examples/structured_output) - JSON schema-constrained output
- [structured_output_jsonschema](./examples/structured_output_jsonschema) - Generate schemas from Go structs (parity with TS `structured_output_zod.ts`)
- [resume_thread](./examples/resume_thread) - Thread persistence and resumption
- [with_images](./examples/with_images) - Image input handling

All examples honor the `CODEX_EXECUTABLE` environment variable (or a local `codex-rs/target/debug/codex`
build) via shared helper options, matching the TypeScript samples' `codexPathOverride`. They will also
work with any authentication you have already configured for the `codex` CLI; set `CODEX_API_KEY` only
if you want to override that.

```bash
go run ./examples/basic_streaming
```

## API Reference

See the [Go package documentation](https://pkg.go.dev/github.com/M1n9X/codex-sdk-go) for complete API reference.

## TypeScript SDK Compatibility

This SDK is designed to be API-compatible with the [official TypeScript SDK](https://github.com/openai/codex/tree/main/sdk/typescript). The following table shows the mapping:

| TypeScript | Go |
|------------|-----|
| `new Codex(options)` | `codex.New(options...)` |
| `codex.startThread(options)` | `client.StartThread(options...)` |
| `codex.resumeThread(id, options)` | `client.ResumeThread(id, options...)` |
| `thread.run(input, turnOptions)` | `thread.Run(ctx, input, opts...)` |
| `thread.runStreamed(input, turnOptions)` | `thread.RunStreamed(ctx, input, opts...)` |
| `thread.id` | `thread.ID()` |
| `"string input"` | `codex.Text("string input")` |
| `[{ type: "text", text }]` | `codex.Compose(codex.TextPart(text))` |
| `{ type: "local_image", path }` | `codex.ImagePart(path)` |
| `{ outputSchema: schema }` | `codex.WithOutputSchema(schema)` |

## License

MIT License - see [LICENSE](./LICENSE) for details.
