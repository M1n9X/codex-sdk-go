// Package main demonstrates resuming a thread by ID.
//
// This example shows how to save a thread ID and resume it later,
// matching the TypeScript SDK's resumeThread functionality.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/M1n9X/codex-sdk-go"
	"github.com/M1n9X/codex-sdk-go/examples/internal/exampleutil"
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
	client, err := codex.New(exampleutil.ClientOptions()...)
	if err != nil {
		return fmt.Errorf("create codex client: %w", err)
	}

	// Start a new thread
	thread := client.StartThread()

	fmt.Println("Starting first turn...")
	turn1, err := thread.Run(ctx, codex.Text("What is 2 + 2?"))
	if err != nil {
		return fmt.Errorf("first run: %w", err)
	}

	fmt.Printf("First response: %s\n", turn1.FinalResponse)
	fmt.Printf("Thread ID: %s\n", thread.ID())
	fmt.Println()

	// Save the thread ID (in a real app, you might persist this)
	savedThreadID := thread.ID()

	// Simulate "losing" the thread by creating a new client
	// In a real application, this could be a different process or session
	newClient, err := codex.New()
	if err != nil {
		return fmt.Errorf("create new client: %w", err)
	}

	// Resume the thread using the saved ID
	fmt.Println("Resuming thread with saved ID...")
	resumedThread := newClient.ResumeThread(savedThreadID)

	// Continue the conversation
	turn2, err := resumedThread.Run(ctx, codex.Text("And what is that multiplied by 3?"))
	if err != nil {
		return fmt.Errorf("resumed run: %w", err)
	}

	fmt.Printf("Second response: %s\n", turn2.FinalResponse)
	fmt.Printf("Thread ID still matches: %v\n", resumedThread.ID() == savedThreadID)

	if turn2.Usage != nil {
		fmt.Printf("\n[Usage: %d input tokens, %d cached, %d output tokens]\n",
			turn2.Usage.InputTokens,
			turn2.Usage.CachedInputTokens,
			turn2.Usage.OutputTokens)
	}

	return nil
}
