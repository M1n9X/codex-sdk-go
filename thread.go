package codex

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"
)

// Thread represents a conversation with the Codex agent.
// One thread can have multiple consecutive turns.
type Thread struct {
	exec          *Exec
	codexOptions  CodexOptions
	threadOptions ThreadOptions
	id            string
	mu            sync.RWMutex
}

// ID returns the identifier of the thread.
// The ID is populated after the first turn starts.
func (t *Thread) ID() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.id
}

func (t *Thread) setID(id string) {
	if id == "" {
		return
	}
	t.mu.Lock()
	t.id = id
	t.mu.Unlock()
}

func (t *Thread) currentID() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.id
}

// Turn contains the result of a completed agent turn.
type Turn struct {
	// Items are the completed thread items emitted during the turn.
	Items []ThreadItem
	// FinalResponse is the assistant's last agent_message text.
	FinalResponse string
	// Usage reports token consumption for the turn.
	Usage *Usage
}

// RunResult is an alias for Turn, matching the TypeScript SDK API.
type RunResult = Turn

// StreamedTurn streams thread events as they are produced during a run.
type StreamedTurn struct {
	// Events yields parsed events in the order emitted by the CLI.
	Events   <-chan ThreadEvent
	waitFn   func() error
	waitOnce sync.Once
	waitErr  error
}

// RunStreamedResult is an alias for StreamedTurn, matching the TypeScript SDK API.
type RunStreamedResult = StreamedTurn

// Wait blocks until the underlying run completes and returns any terminal error.
func (s *StreamedTurn) Wait() error {
	s.waitOnce.Do(func() {
		if s.waitFn != nil {
			s.waitErr = s.waitFn()
		}
	})
	return s.waitErr
}

// Run executes a complete agent turn with the provided input and returns its result.
// The call blocks until the CLI exits or the context is cancelled.
func (t *Thread) Run(ctx context.Context, input Input, opts ...TurnOption) (*Turn, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	streamed, err := t.runStreamedInternal(ctx, input, opts)
	if err != nil {
		return nil, err
	}

	var (
		items         []ThreadItem
		finalResponse string
		usage         *Usage
		turnFailure   *ThreadError
	)

loop:
	for event := range streamed.Events {
		switch event.Type {
		case EventItemCompleted:
			if event.Item != nil {
				if msg, ok := event.Item.(*AgentMessageItem); ok {
					finalResponse = msg.Text
				}
				items = append(items, event.Item)
			}
		case EventTurnCompleted:
			usage = event.Usage
		case EventTurnFailed:
			if event.Error != nil {
				turnFailure = event.Error
			} else {
				turnFailure = &ThreadError{Message: "turn failed"}
			}
			cancel()
			break loop
		}
	}

	waitErr := streamed.Wait()

	if turnFailure != nil {
		if waitErr != nil && !errors.Is(waitErr, context.Canceled) {
			return nil, waitErr
		}
		return nil, errors.New(turnFailure.Message)
	}

	if waitErr != nil {
		return nil, waitErr
	}

	return &Turn{Items: items, FinalResponse: finalResponse, Usage: usage}, nil
}

// RunStreamed streams events for a single agent turn.
// Callers should drain Events and then invoke Wait to retrieve any terminal error.
func (t *Thread) RunStreamed(ctx context.Context, input Input, opts ...TurnOption) (*StreamedTurn, error) {
	return t.runStreamedInternal(ctx, input, opts)
}

func (t *Thread) runStreamedInternal(ctx context.Context, input Input, opts []TurnOption) (*StreamedTurn, error) {
	turnOptions := applyTurnOptions(opts)

	schemaFile, err := createOutputSchemaFile(turnOptions.OutputSchema)
	if err != nil {
		return nil, err
	}

	prompt, images, err := normalizeInput(input)
	if err != nil {
		_ = schemaFile.Cleanup()
		return nil, err
	}

	stream, err := t.exec.Run(ctx, ExecArgs{
		Input:                 prompt,
		BaseURL:               t.codexOptions.BaseURL,
		APIKey:                t.codexOptions.APIKey,
		ThreadID:              t.currentID(),
		Images:                images,
		Model:                 t.threadOptions.Model,
		SandboxMode:           t.threadOptions.SandboxMode,
		WorkingDirectory:      t.threadOptions.WorkingDirectory,
		SkipGitRepoCheck:      t.threadOptions.SkipGitRepoCheck,
		OutputSchemaFile:      schemaFile.Path(),
		ModelReasoningEffort:  t.threadOptions.ModelReasoningEffort,
		NetworkAccessEnabled:  t.threadOptions.NetworkAccessEnabled,
		WebSearchEnabled:      t.threadOptions.WebSearchEnabled,
		ApprovalPolicy:        t.threadOptions.ApprovalPolicy,
		AdditionalDirectories: t.threadOptions.AdditionalDirectories,
	})
	if err != nil {
		_ = schemaFile.Cleanup()
		return nil, err
	}

	events := make(chan ThreadEvent)
	errCh := make(chan error, 1)

	go func() {
		defer close(events)
		stdout := stream.Stdout()
		defer stdout.Close()
		defer func() {
			_ = schemaFile.Cleanup()
		}()

		reader := bufio.NewReader(stdout)
		var runErr error

		for {
			if ctxErr := ctx.Err(); ctxErr != nil {
				runErr = ctxErr
				break
			}

			line, readErr := reader.ReadBytes('\n')
			trimmed := bytes.TrimSpace(line)
			if len(trimmed) > 0 {
				var event ThreadEvent
				if err := json.Unmarshal(trimmed, &event); err != nil {
					runErr = fmt.Errorf("parse codex event: %w", err)
					break
				}

				if event.Type == EventThreadStarted && event.ThreadID != "" {
					t.setID(event.ThreadID)
				}

				select {
				case events <- event:
				case <-ctx.Done():
					runErr = ctx.Err()
					break
				}
			}

			if readErr != nil {
				if errors.Is(readErr, io.EOF) {
					break
				}
				if runErr == nil {
					runErr = fmt.Errorf("read codex output: %w", readErr)
				}
				break
			}

			if runErr != nil {
				break
			}
		}

		waitErr := stream.Wait()
		if runErr == nil {
			runErr = waitErr
		} else if waitErr != nil && !errors.Is(runErr, waitErr) {
			runErr = fmt.Errorf("%w; wait error: %v", runErr, waitErr)
		}

		errCh <- runErr
	}()

	return &StreamedTurn{
		Events: events,
		waitFn: func() error {
			return <-errCh
		},
	}, nil
}
