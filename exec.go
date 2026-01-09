package codex

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
)

const (
	internalOriginatorEnv = "CODEX_INTERNAL_ORIGINATOR_OVERRIDE"
	goSDKOriginator       = "codex_sdk_go"
)

// ExecArgs contains all arguments for running the codex CLI.
type ExecArgs struct {
	Input                 string
	BaseURL               string
	APIKey                string
	ThreadID              string
	Images                []string
	Model                 string
	SandboxMode           SandboxMode
	WorkingDirectory      string
	SkipGitRepoCheck      bool
	OutputSchemaFile      string
	ModelReasoningEffort  ModelReasoningEffort
	NetworkAccessEnabled  *bool
	WebSearchEnabled      *bool
	ApprovalPolicy        ApprovalMode
	AdditionalDirectories []string
}

// Exec manages execution of the codex CLI binary.
type Exec struct {
	path string
	env  map[string]string
}

// newExec creates a new Exec instance.
func newExec(pathOverride string, env map[string]string) (*Exec, error) {
	path := pathOverride
	if path == "" {
		var err error
		path, err = findCodexPath()
		if err != nil {
			return nil, err
		}
	}
	return &Exec{path: path, env: env}, nil
}

// ExecStream provides access to the running codex process.
type ExecStream struct {
	stdout    io.ReadCloser
	waitOnce  sync.Once
	waitErr   error
	waitFn    func() error
	closeOnce sync.Once
	closeErr  error
}

// Stdout returns a reader for the process stdout.
func (s *ExecStream) Stdout() io.ReadCloser {
	return s.stdout
}

// Wait blocks until the process exits and returns any error.
func (s *ExecStream) Wait() error {
	s.waitOnce.Do(func() {
		if s.waitFn != nil {
			s.waitErr = s.waitFn()
		}
	})
	return s.waitErr
}

// Close closes the stdout reader.
func (s *ExecStream) Close() error {
	s.closeOnce.Do(func() {
		if s.stdout != nil {
			s.closeErr = s.stdout.Close()
		}
	})
	return s.closeErr
}

// Run starts the codex CLI with the given arguments.
func (e *Exec) Run(ctx context.Context, args ExecArgs) (*ExecStream, error) {
	commandArgs := []string{"exec", "--experimental-json"}

	if args.Model != "" {
		commandArgs = append(commandArgs, "--model", args.Model)
	}

	if args.SandboxMode != "" {
		commandArgs = append(commandArgs, "--sandbox", string(args.SandboxMode))
	}

	if args.WorkingDirectory != "" {
		commandArgs = append(commandArgs, "--cd", args.WorkingDirectory)
	}

	for _, dir := range args.AdditionalDirectories {
		if dir != "" {
			commandArgs = append(commandArgs, "--add-dir", dir)
		}
	}

	if args.SkipGitRepoCheck {
		commandArgs = append(commandArgs, "--skip-git-repo-check")
	}

	if args.OutputSchemaFile != "" {
		commandArgs = append(commandArgs, "--output-schema", args.OutputSchemaFile)
	}

	if args.ModelReasoningEffort != "" {
		commandArgs = append(commandArgs, "--config", fmt.Sprintf(`model_reasoning_effort="%s"`, args.ModelReasoningEffort))
	}

	if args.NetworkAccessEnabled != nil {
		commandArgs = append(commandArgs, "--config", fmt.Sprintf("sandbox_workspace_write.network_access=%t", *args.NetworkAccessEnabled))
	}

	if args.WebSearchEnabled != nil {
		commandArgs = append(commandArgs, "--config", fmt.Sprintf("features.web_search_request=%t", *args.WebSearchEnabled))
	}

	if args.ApprovalPolicy != "" {
		commandArgs = append(commandArgs, "--config", fmt.Sprintf(`approval_policy="%s"`, args.ApprovalPolicy))
	}

	for _, image := range args.Images {
		if image != "" {
			commandArgs = append(commandArgs, "--image", image)
		}
	}

	if args.ThreadID != "" {
		commandArgs = append(commandArgs, "resume", args.ThreadID)
	}

	cmd := exec.CommandContext(ctx, e.path, commandArgs...)
	cmd.Env = e.buildEnvironment(args.BaseURL, args.APIKey)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("open stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("open stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("open stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start codex exec: %w", err)
	}

	stderrBuf := bytes.NewBuffer(nil)
	stderrDone := make(chan struct{})
	go func() {
		defer close(stderrDone)
		_, _ = io.Copy(stderrBuf, stderr)
	}()

	writeErrCh := make(chan error, 1)
	go func() {
		defer stdin.Close()
		_, err := io.WriteString(stdin, args.Input)
		writeErrCh <- err
	}()

	waitFn := func() error {
		// Wait for process to complete
		err := cmd.Wait()

		// Check if write to stdin failed
		writeErr := <-writeErrCh
		if writeErr != nil {
			return fmt.Errorf("write to codex stdin: %w", writeErr)
		}

		// Ensure stderr is fully drained before checking exit status
		<-stderrDone

		// Check if process exited with error
		if err != nil {
			var exitErr *exec.ExitError
			if errors.As(err, &exitErr) {
				stderrText := strings.TrimSpace(stderrBuf.String())
				return &ErrExecFailed{
					ExitCode: exitErr.ExitCode(),
					Stderr:   stderrText,
					Err:      err,
				}
			}
			return fmt.Errorf("codex exec failed: %w", err)
		}

		return nil
	}

	return &ExecStream{stdout: stdout, waitFn: waitFn}, nil
}

// buildEnvironment constructs the environment for the CLI process.
func (e *Exec) buildEnvironment(baseURL, apiKey string) []string {
	envMap := make(map[string]string)

	if e.env != nil {
		// Use custom environment
		for k, v := range e.env {
			envMap[k] = v
		}
	} else {
		// Inherit from os.Environ
		for _, kv := range os.Environ() {
			if idx := strings.IndexByte(kv, '='); idx >= 0 {
				envMap[kv[:idx]] = kv[idx+1:]
			}
		}
	}

	// Set SDK originator if not already set
	if value, ok := envMap[internalOriginatorEnv]; !ok || value == "" {
		envMap[internalOriginatorEnv] = goSDKOriginator
	}

	// Override with provided values
	if baseURL != "" {
		envMap["OPENAI_BASE_URL"] = baseURL
	}
	if apiKey != "" {
		envMap["CODEX_API_KEY"] = apiKey
	}

	// Convert to slice
	env := make([]string, 0, len(envMap))
	for k, v := range envMap {
		env = append(env, k+"="+v)
	}
	sort.Strings(env)
	return env
}

// findCodexPath searches for the codex binary in PATH.
func findCodexPath() (string, error) {
	if bundled := bundledCodexPath(); bundled != "" {
		return bundled, nil
	}

	codexPath, err := exec.LookPath("codex")
	if err != nil {
		return "", fmt.Errorf("%w: %v (ensure codex is installed and in PATH)", ErrCodexNotFound, err)
	}
	return codexPath, nil
}

func bundledCodexPath() string {
	targetTriple, err := resolveTargetTriple(runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return ""
	}

	// Locate the directory containing this source file.
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok || currentFile == "" {
		return ""
	}

	root := filepath.Dir(currentFile)
	binaryName := "codex"
	if runtime.GOOS == "windows" {
		binaryName = "codex.exe"
	}

	candidate := filepath.Join(root, "vendor", targetTriple, "codex", binaryName)
	if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
		return candidate
	}
	return ""
}

func resolveTargetTriple(goos, goarch string) (string, error) {
	switch goos {
	case "linux", "android":
		switch goarch {
		case "amd64":
			return "x86_64-unknown-linux-musl", nil
		case "arm64":
			return "aarch64-unknown-linux-musl", nil
		}
	case "darwin":
		switch goarch {
		case "amd64":
			return "x86_64-apple-darwin", nil
		case "arm64":
			return "aarch64-apple-darwin", nil
		}
	case "windows":
		switch goarch {
		case "amd64":
			return "x86_64-pc-windows-msvc", nil
		case "arm64":
			return "aarch64-pc-windows-msvc", nil
		}
	}
	return "", fmt.Errorf("unsupported platform %s/%s", goos, goarch)
}
