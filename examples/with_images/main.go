// Package main demonstrates attaching images to input.
//
// This example shows how to include local images alongside text,
// using the Compose function with TextPart and ImagePart.
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

	// Check for image arguments
	if len(os.Args) < 2 {
		fmt.Println("Usage: with_images <image1.png> [image2.jpg] ...")
		fmt.Println()
		fmt.Println("This example demonstrates sending images to the Codex agent.")
		fmt.Println("Provide one or more local image paths as arguments.")
		return nil
	}

	imagePaths := os.Args[1:]

	// Create Codex client
	client, err := codex.New(exampleutil.ClientOptions()...)
	if err != nil {
		return fmt.Errorf("create codex client: %w", err)
	}

	// Start a new thread
	thread := client.StartThread()

	// Build the input with text and images
	parts := make([]codex.UserInput, 0, len(imagePaths)+1)
	parts = append(parts, codex.TextPart("Describe these images in detail:"))

	for _, path := range imagePaths {
		// Verify the file exists
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return fmt.Errorf("image not found: %s", path)
		}
		parts = append(parts, codex.ImagePart(path))
		fmt.Printf("Adding image: %s\n", path)
	}

	fmt.Println()
	fmt.Println("Sending images to Codex...")
	fmt.Println()

	// Run with composed input
	turn, err := thread.Run(ctx, codex.Compose(parts...))
	if err != nil {
		return fmt.Errorf("run: %w", err)
	}

	fmt.Println("Response:")
	fmt.Println(turn.FinalResponse)

	if turn.Usage != nil {
		fmt.Printf("\n[Usage: %d input tokens, %d cached, %d output tokens]\n",
			turn.Usage.InputTokens,
			turn.Usage.CachedInputTokens,
			turn.Usage.OutputTokens)
	}

	return nil
}
