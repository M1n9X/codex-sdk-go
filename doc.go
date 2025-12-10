// Package codex provides a Go SDK for interacting with OpenAI's Codex CLI agent.
//
// This SDK mirrors the official OpenAI TypeScript Codex SDK, allowing developers
// to integrate Codex capabilities into Go applications. The SDK wraps the bundled
// codex binary and exchanges JSONL events over stdin/stdout.
//
// # Quick Start
//
// Create a Codex client and start a conversation:
//
//	client, err := codex.New()
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	thread := client.StartThread()
//	turn, err := thread.Run(ctx, codex.Text("Diagnose the test failure and propose a fix"))
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	fmt.Println(turn.FinalResponse)
//
// # Streaming Responses
//
// Use RunStreamed to receive events as they are produced:
//
//	streamed, err := thread.RunStreamed(ctx, codex.Text("Implement the fix"))
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	for event := range streamed.Events {
//		switch event.Type {
//		case codex.EventItemCompleted:
//			fmt.Println("item:", event.Item)
//		case codex.EventTurnCompleted:
//			fmt.Println("usage:", event.Usage)
//		}
//	}
//
//	if err := streamed.Wait(); err != nil {
//		log.Fatal(err)
//	}
//
// # Structured Output
//
// Request JSON output conforming to a schema:
//
//	schema := map[string]any{
//		"type": "object",
//		"properties": map[string]any{
//			"summary": map[string]any{"type": "string"},
//			"status":  map[string]any{"type": "string", "enum": []string{"ok", "action_required"}},
//		},
//		"required":             []string{"summary", "status"},
//		"additionalProperties": false,
//	}
//
//	turn, err := thread.Run(ctx, codex.Text("Summarize repository status"),
//		codex.WithOutputSchema(schema))
//
// # Bundled codex binary
//
// When available, the SDK prefers a bundled codex binary at vendor/<triple>/codex/codex
// (for example, vendor/aarch64-apple-darwin/codex/codex). If no bundled binary is found,
// it falls back to resolving "codex" from PATH.
//
// # Attaching Images
//
// Include images alongside text using Compose:
//
//	turn, err := thread.Run(ctx, codex.Compose(
//		codex.TextPart("Describe these screenshots"),
//		codex.ImagePart("./ui.png"),
//		codex.ImagePart("./diagram.jpg"),
//	))
//
// # Resuming Threads
//
// Threads are persisted in ~/.codex/sessions. Resume a thread by ID:
//
//	thread := client.ResumeThread(savedThreadID)
//	turn, err := thread.Run(ctx, codex.Text("Continue the conversation"))
//
// # Configuration
//
// Configure the client with functional options:
//
//	client, err := codex.New(
//		codex.WithAPIKey("sk-..."),
//		codex.WithBaseURL("https://api.example.com"),
//	)
//
// Configure threads with thread options:
//
//	thread := client.StartThread(
//		codex.WithModel("gpt-4"),
//		codex.WithSandboxMode(codex.SandboxWorkspaceWrite),
//		codex.WithWorkingDirectory("/path/to/project"),
//	)
package codex
