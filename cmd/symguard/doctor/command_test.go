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
	code := Run(&buf)
	out := buf.String()

	if code != 0 {
		t.Errorf("Run() = %d, want 0 on an empty machine:\n%s", code, out)
	}
	for _, want := range []string{
		"symguard doctor",
		"Version:   test-1.2.3",
		"config           not configured (no config file found)",
		"policy           defaults only (no rules — deny by default)",
		"audit log        not initialized (created on first 'symguard decide')",
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
	code := Run(&buf)
	out := buf.String()

	if code != 1 {
		t.Errorf("Run() = %d, want 1 when a server is denied:\n%s", code, out)
	}
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
	// fail-closed error: the config line reports it, policy cannot load,
	// and the run exits non-zero (both failing checks count as issues).
	bad := filepath.Join(t.TempDir(), "bad.toml")
	if err := os.WriteFile(bad, []byte("this is not [valid toml"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	t.Setenv("SYMGUARD_CONFIG", bad)

	var buf strings.Builder
	code := Run(&buf)
	out := buf.String()

	if code != 1 {
		t.Errorf("Run() = %d, want 1 on a config error:\n%s", code, out)
	}
	if !strings.Contains(out, "config           error:") {
		t.Errorf("Run() output missing config error line:\n%s", out)
	}
	if !strings.Contains(out, "policy           not loaded (config error)") {
		t.Errorf("Run() output missing policy error line:\n%s", out)
	}
	if !strings.Contains(out, "2 issue(s) found") {
		t.Errorf("Run() output missing issue summary:\n%s", out)
	}
	if strings.Contains(out, "All basic checks passed") {
		t.Errorf("Run() reported all-clear despite a config error:\n%s", out)
	}
}

func TestRun_ConfigPresent(t *testing.T) {
	// A valid config file must flip the config line to ok and the policy
	// line to the real rule count instead of the old hardcoded stubs.
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("SYMGUARD_CONFIG", filepath.Join(dir, "config.toml"))
	cfg := "[defaults]\nshell = \"allow\"\nread_secret = \"deny\"\n\n[[rules]]\nmatch.server = \"symmemory\"\nmatch.tool = \"memory_search\"\ndecision = \"allow\"\n"
	if err := os.WriteFile(filepath.Join(dir, "config.toml"), []byte(cfg), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	doctorSetVersion(t, "test-1.2.3")

	var buf strings.Builder
	code := Run(&buf)
	out := buf.String()

	if code != 0 {
		t.Errorf("Run() = %d, want 0 for a healthy config:\n%s", code, out)
	}
	for _, want := range []string{
		"config           ok",
		"policy           ok (1 rule(s))",
		"All basic checks passed",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("Run() output missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "issue(s) found") {
		t.Errorf("Run() reported issues on a healthy config:\n%s", out)
	}
}

func TestRun_AuditLogWithoutAnchor(t *testing.T) {
	// An audit log whose chain anchor is missing loses truncation
	// detection: doctor must report it as a problem and exit non-zero.
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("SYMGUARD_CONFIG", filepath.Join(t.TempDir(), "missing.toml"))
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	logPath := filepath.Join(os.Getenv("XDG_DATA_HOME"), "symguard", "audit.log")
	if err := os.MkdirAll(filepath.Dir(logPath), 0o700); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(logPath, []byte("{\"entry\":1}\n"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	doctorSetVersion(t, "test-1.2.3")

	var buf strings.Builder
	code := Run(&buf)
	out := buf.String()

	if code != 1 {
		t.Errorf("Run() = %d, want 1 when the audit anchor is missing:\n%s", code, out)
	}
	if !strings.Contains(out, "audit log        error: anchor") || !strings.Contains(out, "1 issue(s) found") {
		t.Errorf("Run() output missing anchor problem:\n%s", out)
	}
}
