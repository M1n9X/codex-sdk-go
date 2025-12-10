package codex

import "testing"

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
