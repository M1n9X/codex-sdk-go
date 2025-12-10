// Package main demonstrates basic streaming usage of the Codex SDK.
//
// This example corresponds to the TypeScript SDK's samples/basic_streaming.ts.
// It creates a Codex client, starts a thread, and processes streaming events
// as they are produced by the agent.
package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
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

	// Create a scanner for reading user input
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("Codex SDK - Basic Streaming Example")
	fmt.Println("Type your messages and press Enter. Use Ctrl+C to exit.")
	fmt.Println()

	for {
		fmt.Print("> ")

		// Check for context cancellation before blocking on input
		select {
		case <-ctx.Done():
			fmt.Println("\nGoodbye!")
			return nil
		default:
		}

		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		// Run with streaming
		streamed, err := thread.RunStreamed(ctx, codex.Text(input))
		if err != nil {
			return fmt.Errorf("run streamed: %w", err)
		}

		// Process events as they arrive
		for event := range streamed.Events {
			handleEvent(event)
		}

		// Check for any errors after the stream completes
		if err := streamed.Wait(); err != nil {
			fmt.Fprintf(os.Stderr, "Stream error: %v\n", err)
		}

		fmt.Println()
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read input: %w", err)
	}

	return nil
}

func handleEvent(event codex.ThreadEvent) {
	switch event.Type {
	case codex.EventItemCompleted:
		handleItemCompleted(event.Item)
	case codex.EventItemUpdated, codex.EventItemStarted:
		handleItemUpdated(event.Item)
	case codex.EventTurnCompleted:
		if event.Usage != nil {
			fmt.Printf("\n[Usage: %d input tokens, %d cached, %d output tokens]\n",
				event.Usage.InputTokens,
				event.Usage.CachedInputTokens,
				event.Usage.OutputTokens)
		}
	case codex.EventTurnFailed:
		if event.Error != nil {
			fmt.Fprintf(os.Stderr, "\n[Turn failed: %s]\n", event.Error.Message)
		}
	}
}

func handleItemCompleted(item codex.ThreadItem) {
	if item == nil {
		return
	}

	switch v := item.(type) {
	case *codex.AgentMessageItem:
		fmt.Printf("\nAssistant: %s\n", v.Text)
	case *codex.ReasoningItem:
		fmt.Printf("\n[Reasoning: %s]\n", v.Text)
	case *codex.CommandExecutionItem:
		exitText := ""
		if v.ExitCode != nil {
			exitText = fmt.Sprintf(" (exit code %d)", *v.ExitCode)
		}
		fmt.Printf("\n[Command: %s - %s%s]\n", v.Command, v.Status, exitText)
	case *codex.FileChangeItem:
		for _, change := range v.Changes {
			fmt.Printf("\n[File %s: %s]\n", change.Kind, change.Path)
		}
	case *codex.McpToolCallItem:
		fmt.Printf("\n[MCP Tool: %s/%s - %s]\n", v.Server, v.Tool, v.Status)
	case *codex.WebSearchItem:
		fmt.Printf("\n[Web Search: %s]\n", v.Query)
	case *codex.ErrorItem:
		fmt.Fprintf(os.Stderr, "\n[Error: %s]\n", v.Message)
	}
}

func handleItemUpdated(item codex.ThreadItem) {
	if item == nil {
		return
	}

	switch v := item.(type) {
	case *codex.TodoListItem:
		fmt.Println("\n[Todo List:]")
		for _, todo := range v.Items {
			marker := " "
			if todo.Completed {
				marker = "x"
			}
			fmt.Printf("  [%s] %s\n", marker, todo.Text)
		}
	}
}
