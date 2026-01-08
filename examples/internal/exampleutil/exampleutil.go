package exampleutil

import (
	"os"
	"path/filepath"

	"github.com/M1n9X/codex-sdk-go"
)

// ClientOptions returns Codex client options that mirror the TypeScript
// samples' codexPathOverride helper:
//  1. Respect CODEX_EXECUTABLE if set.
//  2. Otherwise, prefer a repo-local debug build at ../../codex-rs/target/debug/codex.
//  3. Fall back to PATH/bundled resolution.
func ClientOptions() []codex.Option {
	var opts []codex.Option

	if path := os.Getenv("CODEX_EXECUTABLE"); path != "" {
		opts = append(opts, codex.WithCodexPath(path))
		return opts
	}

	// Try local debug build (useful when hacking on codex-rs alongside the SDK).
	if cwd, err := os.Getwd(); err == nil {
		candidate := filepath.Clean(filepath.Join(cwd, "..", "..", "codex-rs", "target", "debug", "codex"))
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			opts = append(opts, codex.WithCodexPath(candidate))
			return opts
		}
	}

	return opts
}
