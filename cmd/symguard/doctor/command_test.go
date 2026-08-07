package doctor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// hermeticEnv points every host-dependent path (config, discovery, home) at
// fresh temp dirs so Run() sees a deterministic, empty machine.
func hermeticEnv(t *testing.T) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("SYMGUARD_CONFIG", filepath.Join(t.TempDir(), "missing.toml"))
}

// doctorSetVersion pins the version callback for the duration of the test
// and restores the default afterwards.
func doctorSetVersion(t *testing.T, v string) {
	t.Helper()
	SetVersion(func() string { return v })
	t.Cleanup(func() { SetVersion(func() string { return "dev" }) })
}

func TestRun_EmptyAllowlistNoServers(t *testing.T) {
	hermeticEnv(t)
	doctorSetVersion(t, "test-1.2.3")

	var buf strings.Builder
	Run(&buf)
	out := buf.String()

	for _, want := range []string{
		"symguard doctor",
		"Version:   test-1.2.3",
		"not configured (empty — deny by default)",
		"none discovered",
		"All basic checks passed",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("Run() output missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "issue(s) found") {
		t.Errorf("Run() reported issues on an empty machine:\n%s", out)
	}
}

func TestRun_DiscoveredServersDenied(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("SYMGUARD_CONFIG", filepath.Join(t.TempDir(), "missing.toml"))

	// One discovered Cursor server; the empty spawn allowlist denies it.
	cursorDir := filepath.Join(home, ".cursor")
	if err := os.MkdirAll(cursorDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	mcp := `{"mcpServers": {"demo": {"command": "/usr/bin/true", "args": ["--once"]}}}`
	if err := os.WriteFile(filepath.Join(cursorDir, "mcp.json"), []byte(mcp), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	doctorSetVersion(t, "test-1.2.3")

	var buf strings.Builder
	Run(&buf)
	out := buf.String()

	if !strings.Contains(out, "[DENIED]") {
		t.Errorf("Run() output missing DENIED verdict:\n%s", out)
	}
	if !strings.Contains(out, "demo (cursor/stdio)") {
		t.Errorf("Run() output missing discovered server line:\n%s", out)
	}
	if !strings.Contains(out, "1 issue(s) found") {
		t.Errorf("Run() output missing issue summary:\n%s", out)
	}
	if strings.Contains(out, "All basic checks passed") {
		t.Errorf("Run() reported all-clear despite a denied server:\n%s", out)
	}
}

func TestRun_ConfigError(t *testing.T) {
	hermeticEnv(t)
	doctorSetVersion(t, "test-1.2.3")

	// Invalid TOML at the resolved config path must surface as a
	// fail-closed allowlist error, not a panic or silent pass.
	bad := filepath.Join(t.TempDir(), "bad.toml")
	if err := os.WriteFile(bad, []byte("this is not [valid toml"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	t.Setenv("SYMGUARD_CONFIG", bad)

	var buf strings.Builder
	Run(&buf)
	out := buf.String()

	if !strings.Contains(out, "spawn allowlist") || !strings.Contains(out, "error:") {
		t.Errorf("Run() output missing allowlist error line:\n%s", out)
	}
	if !strings.Contains(out, "All basic checks passed") {
		t.Errorf("Run() output missing summary line:\n%s", out)
	}
}
