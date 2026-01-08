// Package main demonstrates generating a JSON schema from a Go struct and
// requesting structured output from the Codex agent.
//
// This mirrors the TypeScript example `structured_output_zod.ts`, replacing
// Zod with Go's JSON Schema generation.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/M1n9X/codex-sdk-go"
	"github.com/M1n9X/codex-sdk-go/examples/internal/exampleutil"
	"github.com/invopop/jsonschema"
)

// RepoStatus is the structured shape we want back from Codex.
type RepoStatus struct {
	Summary string `json:"summary"`
	Status  string `json:"status" jsonschema:"enum=ok,enum=action_required"`
}

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

	// Reflect a JSON schema from the Go struct (similar to Zod->JSON Schema).
	// Configure the reflector to inline the struct instead of using $ref/$defs
	// because the Codex CLI expects the root schema object directly.
	reflector := &jsonschema.Reflector{
		RequiredFromJSONSchemaTags: true,
		DoNotReference:             true,
		ExpandedStruct:             true,
	}
	rawSchema := reflector.Reflect(&RepoStatus{})

	// The Codex CLI expects a plain JSON object (no $schema/$ref). Normalize
	// the generated schema to match the TypeScript example shape.
	var schemaMap map[string]any
	b, err := json.Marshal(rawSchema)
	if err != nil {
		return fmt.Errorf("marshal schema: %w", err)
	}
	if err := json.Unmarshal(b, &schemaMap); err != nil {
		return fmt.Errorf("unmarshal schema: %w", err)
	}
	delete(schemaMap, "$schema")
	if _, ok := schemaMap["required"]; !ok {
		if props, ok := schemaMap["properties"].(map[string]any); ok {
			req := make([]string, 0, len(props))
			for k := range props {
				req = append(req, k)
			}
			schemaMap["required"] = req
		}
	}

	fmt.Println("Requesting structured output using a schema derived from a Go struct...")
	fmt.Println()

	turn, err := thread.Run(ctx, codex.Text("Summarize repository status"),
		codex.WithOutputSchema(schemaMap))
	if err != nil {
		return fmt.Errorf("run: %w", err)
	}

	var result RepoStatus
	if err := json.Unmarshal([]byte(turn.FinalResponse), &result); err != nil {
		fmt.Println("Raw response (could not parse as RepoStatus):")
		fmt.Println(turn.FinalResponse)
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
