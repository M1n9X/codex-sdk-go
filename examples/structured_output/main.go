// Package main demonstrates structured output usage of the Codex SDK.
//
// This example corresponds to the TypeScript SDK's samples/structured_output.ts.
// It shows how to request JSON output conforming to a schema.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/M1n9X/codex-sdk-go"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Create Codex client
	client, err := codex.New()
	if err != nil {
		return fmt.Errorf("create codex client: %w", err)
	}

	// Start a new thread
	thread := client.StartThread()

	// Define the output schema
	// This matches the TypeScript example's schema
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"summary": map[string]any{
				"type": "string",
			},
			"status": map[string]any{
				"type": "string",
				"enum": []string{"ok", "action_required"},
			},
		},
		"required":             []string{"summary", "status"},
		"additionalProperties": false,
	}

	fmt.Println("Requesting structured output from Codex...")
	fmt.Println()

	// Run with the output schema
	turn, err := thread.Run(ctx, codex.Text("Summarize repository status"),
		codex.WithOutputSchema(schema))
	if err != nil {
		return fmt.Errorf("run: %w", err)
	}

	// Parse the structured response
	var result struct {
		Summary string `json:"summary"`
		Status  string `json:"status"`
	}

	if err := json.Unmarshal([]byte(turn.FinalResponse), &result); err != nil {
		// If parsing fails, just print the raw response
		fmt.Println("Raw response:", turn.FinalResponse)
		return nil
	}

	fmt.Println("Structured Response:")
	fmt.Printf("  Summary: %s\n", result.Summary)
	fmt.Printf("  Status:  %s\n", result.Status)

	if turn.Usage != nil {
		fmt.Printf("\n[Usage: %d input tokens, %d cached, %d output tokens]\n",
			turn.Usage.InputTokens,
			turn.Usage.CachedInputTokens,
			turn.Usage.OutputTokens)
	}

	return nil
}
