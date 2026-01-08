package codex

import (
	"bufio"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestResolveTargetTriple(t *testing.T) {
	tests := []struct {
		name    string
		goos    string
		goarch  string
		want    string
		wantErr bool
	}{
		{name: "linux_amd64", goos: "linux", goarch: "amd64", want: "x86_64-unknown-linux-musl"},
		{name: "linux_arm64", goos: "linux", goarch: "arm64", want: "aarch64-unknown-linux-musl"},
		{name: "darwin_amd64", goos: "darwin", goarch: "amd64", want: "x86_64-apple-darwin"},
		{name: "darwin_arm64", goos: "darwin", goarch: "arm64", want: "aarch64-apple-darwin"},
		{name: "windows_amd64", goos: "windows", goarch: "amd64", want: "x86_64-pc-windows-msvc"},
		{name: "windows_arm64", goos: "windows", goarch: "arm64", want: "aarch64-pc-windows-msvc"},
		{name: "unsupported_arch", goos: "linux", goarch: "ppc64", wantErr: true},
		{name: "unsupported_os", goos: "plan9", goarch: "amd64", wantErr: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveTargetTriple(tt.goos, tt.goarch)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got triple=%s", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

// TestExecEarlyExit tests that the SDK properly handles when the codex process
// exits early (before stdout is fully drained), ensuring no hangs occur and
// stderr is properly captured.
func TestExecEarlyExit(t *testing.T) {
	// Create a fake codex script that exits immediately with an error
	// but writes to both stderr and stdout
	fakeCodexScript := createFakeCodexScript(t)
	defer os.Remove(fakeCodexScript)

	exec, err := newExec(fakeCodexScript, nil)
	if err != nil {
		t.Fatalf("failed to create exec: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	stream, err := exec.Run(ctx, ExecArgs{
		Input: "test input",
	})
	if err != nil {
		t.Fatalf("failed to start exec: %v", err)
	}
	defer stream.Close()

	// Read from stdout (should complete without hanging)
	scanner := bufio.NewScanner(stream.Stdout())
	for scanner.Scan() {
		// Consume stdout
	}
	if err := scanner.Err(); err != nil {
		t.Logf("scanner error (may be expected if process exits early): %v", err)
	}

	// Wait should return an error with stderr content
	waitErr := stream.Wait()
	if waitErr == nil {
		t.Fatal("expected Wait to return error, got nil")
	}

	errMsg := waitErr.Error()
	if !strings.Contains(errMsg, "codex exec exited with code") {
		t.Errorf("expected error message to contain 'codex exec exited with code', got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "early exit error") {
		t.Errorf("expected error message to contain stderr output 'early exit error', got: %s", errMsg)
	}

	// Verify the test completed within timeout (no hang)
	select {
	case <-ctx.Done():
		t.Fatal("test timed out, likely due to hang")
	default:
		// Test completed successfully
	}
}

// createFakeCodexScript creates a temporary script that simulates codex exec
// exiting early with stderr output.
func createFakeCodexScript(t *testing.T) string {
	t.Helper()

	var scriptContent string
	var scriptName string

	if runtime.GOOS == "windows" {
		scriptName = "fake-codex.bat"
		scriptContent = `@echo off
echo {"type":"test"}
echo early exit error 1>&2
exit /b 2
`
	} else {
		scriptName = "fake-codex.sh"
		scriptContent = `#!/bin/sh
echo '{"type":"test"}'
echo 'early exit error' >&2
exit 2
`
	}

	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, scriptName)

	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		t.Fatalf("failed to create fake codex script: %v", err)
	}

	return scriptPath
}

// TestExecStreamReadAndWait tests that reading stdout and calling Wait works correctly.
func TestExecStreamReadAndWait(t *testing.T) {
	// Create a fake codex script that outputs multiple lines
	fakeCodexScript := createFakeCodexMultilineScript(t)
	defer os.Remove(fakeCodexScript)

	exec, err := newExec(fakeCodexScript, nil)
	if err != nil {
		t.Fatalf("failed to create exec: %v", err)
	}

	ctx := context.Background()
	stream, err := exec.Run(ctx, ExecArgs{
		Input: "test input",
	})
	if err != nil {
		t.Fatalf("failed to start exec: %v", err)
	}
	defer stream.Close()

	// Read all output
	var lines []string
	scanner := bufio.NewScanner(stream.Stdout())
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scanner error: %v", err)
	}

	// Verify we got expected output
	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d", len(lines))
	}

	// Wait should succeed
	if err := stream.Wait(); err != nil {
		t.Errorf("Wait returned unexpected error: %v", err)
	}
}

func createFakeCodexMultilineScript(t *testing.T) string {
	t.Helper()

	var scriptContent string
	var scriptName string

	if runtime.GOOS == "windows" {
		scriptName = "fake-codex-multiline.bat"
		scriptContent = `@echo off
echo {"type":"line1"}
echo {"type":"line2"}
echo {"type":"line3"}
exit /b 0
`
	} else {
		scriptName = "fake-codex-multiline.sh"
		scriptContent = `#!/bin/sh
echo '{"type":"line1"}'
echo '{"type":"line2"}'
echo '{"type":"line3"}'
exit 0
`
	}

	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, scriptName)

	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		t.Fatalf("failed to create fake codex script: %v", err)
	}

	return scriptPath
}

// TestExecWithRealCodex tests exec with the real codex binary if available.
// This test is skipped if codex is not found in PATH.
func TestExecWithRealCodex(t *testing.T) {
	// Check if codex is available
	_, err := exec.LookPath("codex")
	if err != nil {
		t.Skip("codex binary not found in PATH, skipping integration test")
	}

	exec, err := newExec("", nil)
	if err != nil {
		t.Fatalf("failed to create exec: %v", err)
	}

	// Use a short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stream, err := exec.Run(ctx, ExecArgs{
		Input:            "echo hello",
		SkipGitRepoCheck: true,
	})
	if err != nil {
		t.Fatalf("failed to start exec: %v", err)
	}
	defer stream.Close()

	// Read some output (we don't care about the exact content)
	scanner := bufio.NewScanner(stream.Stdout())
	lineCount := 0
	for scanner.Scan() && lineCount < 100 {
		lineCount++
		t.Logf("Line %d: %s", lineCount, scanner.Text())
	}

	// Note: We might not get a clean exit due to timeout or API requirements,
	// but we should not hang
	if err := stream.Wait(); err != nil {
		t.Logf("Wait returned error (may be expected): %v", err)
	}
}
