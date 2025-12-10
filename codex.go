package codex

// Codex is the main entry point for interacting with the Codex agent.
//
// Use New() to create a client, then StartThread() to begin a new conversation
// or ResumeThread() to continue an existing one.
type Codex struct {
	exec    *Exec
	options CodexOptions
}

// New creates a new Codex client with the given options.
//
// Example:
//
//	client, err := codex.New()
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// With options:
//	client, err := codex.New(
//		codex.WithAPIKey("sk-..."),
//		codex.WithBaseURL("https://api.example.com"),
//	)
func New(opts ...Option) (*Codex, error) {
	options := applyCodexOptions(opts)

	exec, err := newExec(options.CodexPath, options.Env)
	if err != nil {
		return nil, err
	}

	return &Codex{
		exec:    exec,
		options: options,
	}, nil
}

// StartThread starts a new conversation with the agent.
//
// Example:
//
//	thread := client.StartThread()
//	turn, err := thread.Run(ctx, codex.Text("Hello!"))
//
//	// With options:
//	thread := client.StartThread(
//		codex.WithModel("gpt-4"),
//		codex.WithSandboxMode(codex.SandboxWorkspaceWrite),
//	)
func (c *Codex) StartThread(opts ...ThreadOption) *Thread {
	threadOptions := applyThreadOptions(opts)
	return &Thread{
		exec:          c.exec,
		codexOptions:  c.options,
		threadOptions: threadOptions,
	}
}

// ResumeThread resumes a conversation based on the thread ID.
// Threads are persisted in ~/.codex/sessions.
//
// Example:
//
//	savedID := "thread_abc123"
//	thread := client.ResumeThread(savedID)
//	turn, err := thread.Run(ctx, codex.Text("Continue our conversation"))
func (c *Codex) ResumeThread(id string, opts ...ThreadOption) *Thread {
	threadOptions := applyThreadOptions(opts)
	return &Thread{
		exec:          c.exec,
		codexOptions:  c.options,
		threadOptions: threadOptions,
		id:            id,
	}
}
