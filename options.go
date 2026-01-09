package codex

// SandboxMode controls the filesystem sandbox granted to the agent.
type SandboxMode string

const (
	// SandboxReadOnly grants read-only access to the filesystem.
	SandboxReadOnly SandboxMode = "read-only"
	// SandboxWorkspaceWrite grants write access to the workspace directory.
	SandboxWorkspaceWrite SandboxMode = "workspace-write"
	// SandboxDangerFullAccess grants full filesystem access (use with caution).
	SandboxDangerFullAccess SandboxMode = "danger-full-access"
)

// ApprovalMode controls when the agent requests user approval.
type ApprovalMode string

const (
	// ApprovalNever never requests approval.
	ApprovalNever ApprovalMode = "never"
	// ApprovalOnRequest requests approval when the agent asks.
	ApprovalOnRequest ApprovalMode = "on-request"
	// ApprovalOnFailure requests approval when an operation fails.
	ApprovalOnFailure ApprovalMode = "on-failure"
	// ApprovalUntrusted requests approval for untrusted operations.
	ApprovalUntrusted ApprovalMode = "untrusted"
)

// ModelReasoningEffort controls the reasoning intensity of the model.
type ModelReasoningEffort string

const (
	// ReasoningMinimal uses minimal reasoning.
	ReasoningMinimal ModelReasoningEffort = "minimal"
	// ReasoningLow uses low reasoning effort.
	ReasoningLow ModelReasoningEffort = "low"
	// ReasoningMedium uses medium reasoning effort.
	ReasoningMedium ModelReasoningEffort = "medium"
	// ReasoningHigh uses high reasoning effort.
	ReasoningHigh ModelReasoningEffort = "high"
	// ReasoningXHigh uses extra-high reasoning effort.
	ReasoningXHigh ModelReasoningEffort = "xhigh"
)

// CodexOptions configures a Codex client.
type CodexOptions struct {
	// CodexPath points to a specific codex binary. When empty, the SDK
	// searches for codex in PATH.
	CodexPath string

	// BaseURL overrides the default API base URL. When empty, the CLI's
	// default value is used.
	BaseURL string

	// APIKey overrides the API key. When empty, the CLI falls back to
	// the CODEX_API_KEY environment variable.
	APIKey string

	// Env specifies environment variables passed to the Codex CLI process.
	// When provided, the SDK will not inherit variables from os.Environ().
	Env map[string]string
}

// Option is a functional option for configuring a Codex client.
type Option func(*CodexOptions)

// WithCodexPath sets a custom path to the codex binary.
// No-op when path is empty.
func WithCodexPath(path string) Option {
	return func(o *CodexOptions) {
		if path != "" {
			o.CodexPath = path
		}
	}
}

// WithBaseURL sets the API base URL.
// No-op when url is empty.
func WithBaseURL(url string) Option {
	return func(o *CodexOptions) {
		if url != "" {
			o.BaseURL = url
		}
	}
}

// WithAPIKey sets the API key.
// No-op when key is empty.
func WithAPIKey(key string) Option {
	return func(o *CodexOptions) {
		if key != "" {
			o.APIKey = key
		}
	}
}

// WithEnv sets custom environment variables for the CLI process.
// When set, os.Environ() will not be inherited.
func WithEnv(env map[string]string) Option {
	return func(o *CodexOptions) {
		o.Env = env
	}
}

// ThreadOptions configures how a thread interacts with the Codex CLI.
type ThreadOptions struct {
	// Model selects the model identifier to run the agent with.
	Model string

	// SandboxMode controls the filesystem sandbox granted to the agent.
	SandboxMode SandboxMode

	// WorkingDirectory sets the directory provided to --cd when launching the CLI.
	WorkingDirectory string

	// SkipGitRepoCheck skips the Git repository check (--skip-git-repo-check).
	SkipGitRepoCheck bool

	// ModelReasoningEffort sets the reasoning intensity of the model.
	ModelReasoningEffort ModelReasoningEffort

	// NetworkAccessEnabled enables network access for the agent.
	// Use a pointer to distinguish between unset and false.
	NetworkAccessEnabled *bool

	// WebSearchEnabled enables web search for the agent.
	// Use a pointer to distinguish between unset and false.
	WebSearchEnabled *bool

	// ApprovalPolicy sets when the agent requests user approval.
	ApprovalPolicy ApprovalMode

	// AdditionalDirectories specifies additional directories accessible to the agent.
	AdditionalDirectories []string
}

// ThreadOption is a functional option for configuring a Thread.
type ThreadOption func(*ThreadOptions)

// WithModel sets the model identifier.
// No-op when model is empty.
func WithModel(model string) ThreadOption {
	return func(o *ThreadOptions) {
		if model != "" {
			o.Model = model
		}
	}
}

// WithSandboxMode sets the sandbox mode.
func WithSandboxMode(mode SandboxMode) ThreadOption {
	return func(o *ThreadOptions) {
		o.SandboxMode = mode
	}
}

// WithWorkingDirectory sets the working directory.
// No-op when dir is empty.
func WithWorkingDirectory(dir string) ThreadOption {
	return func(o *ThreadOptions) {
		if dir != "" {
			o.WorkingDirectory = dir
		}
	}
}

// WithSkipGitRepoCheck skips the Git repository check.
func WithSkipGitRepoCheck() ThreadOption {
	return func(o *ThreadOptions) {
		o.SkipGitRepoCheck = true
	}
}

// WithModelReasoningEffort sets the reasoning effort level.
func WithModelReasoningEffort(effort ModelReasoningEffort) ThreadOption {
	return func(o *ThreadOptions) {
		o.ModelReasoningEffort = effort
	}
}

// WithNetworkAccess enables or disables network access.
func WithNetworkAccess(enabled bool) ThreadOption {
	return func(o *ThreadOptions) {
		o.NetworkAccessEnabled = &enabled
	}
}

// WithWebSearch enables or disables web search.
func WithWebSearch(enabled bool) ThreadOption {
	return func(o *ThreadOptions) {
		o.WebSearchEnabled = &enabled
	}
}

// WithApprovalPolicy sets the approval policy.
func WithApprovalPolicy(policy ApprovalMode) ThreadOption {
	return func(o *ThreadOptions) {
		o.ApprovalPolicy = policy
	}
}

// WithAdditionalDirectories adds directories accessible to the agent.
func WithAdditionalDirectories(dirs ...string) ThreadOption {
	return func(o *ThreadOptions) {
		o.AdditionalDirectories = append(o.AdditionalDirectories, dirs...)
	}
}

// TurnOptions configures a single turn when running the agent.
type TurnOptions struct {
	// OutputSchema describes the expected JSON structure when requesting
	// structured output. The value must marshal to a JSON object.
	OutputSchema any
}

// TurnOption is a functional option for configuring a Turn.
type TurnOption func(*TurnOptions)

// WithOutputSchema sets the expected output schema for structured output.
func WithOutputSchema(schema any) TurnOption {
	return func(o *TurnOptions) {
		o.OutputSchema = schema
	}
}

// applyCodexOptions applies functional options to CodexOptions.
func applyCodexOptions(opts []Option) CodexOptions {
	var options CodexOptions
	for _, opt := range opts {
		opt(&options)
	}
	return options
}

// applyThreadOptions applies functional options to ThreadOptions.
func applyThreadOptions(opts []ThreadOption) ThreadOptions {
	var options ThreadOptions
	for _, opt := range opts {
		opt(&options)
	}
	return options
}

// applyTurnOptions applies functional options to TurnOptions.
func applyTurnOptions(opts []TurnOption) TurnOptions {
	var options TurnOptions
	for _, opt := range opts {
		opt(&options)
	}
	return options
}
