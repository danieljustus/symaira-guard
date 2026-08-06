package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/danieljustus/symaira-guard/cmd/symguard/doctor"
	"github.com/danieljustus/symaira-guard/cmd/symguard/version"
)

func TestRun_NoArgs(t *testing.T) {
	var buf bytes.Buffer
	code := run(nil, &buf)
	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}
	out := buf.String()
	if !strings.Contains(out, "symguard") {
		t.Error("expected usage message on no args")
	}
}

func TestRun_Help(t *testing.T) {
	tests := []struct {
		name string
		arg  string
	}{
		{"help", "help"},
		{"--help", "--help"},
		{"-h", "-h"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			code := run([]string{tt.arg}, &buf)
			if code != 0 {
				t.Errorf("expected exit code 0, got %d", code)
			}
			out := buf.String()
			if !strings.Contains(out, "symguard") || !strings.Contains(out, "Commands:") {
				t.Error("expected usage message")
			}
		})
	}
}

func TestRun_UnknownCommand(t *testing.T) {
	var buf bytes.Buffer
	code := run([]string{"bogus"}, &buf)
	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}
	out := buf.String()
	if !strings.Contains(out, "unknown command: bogus") {
		t.Errorf("expected unknown command error, got: %s", out)
	}
}

func TestRun_Version(t *testing.T) {
	var buf bytes.Buffer
	code := run([]string{"version"}, &buf)
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
	out := buf.String()
	if !strings.Contains(out, "symguard") {
		t.Error("expected version output to contain 'symguard'")
	}
	if !strings.Contains(out, "go") {
		t.Error("expected version output to contain go version")
	}
}

func TestRun_Doctor(t *testing.T) {
	var buf bytes.Buffer
	code := run([]string{"doctor"}, &buf)
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
	out := buf.String()
	if !strings.Contains(out, "symguard doctor") {
		t.Error("expected doctor header")
	}
	if !strings.Contains(out, "Version:") {
		t.Error("expected version in doctor output")
	}
	if !strings.Contains(out, "All basic checks passed") {
		t.Error("expected completion message")
	}
}

func TestRun_GrantsList(t *testing.T) {
	t.Setenv("SYMGUARD_DATA", t.TempDir())
	var buf bytes.Buffer
	code := run([]string{"grants", "list"}, &buf)
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
	out := buf.String()
	if !strings.Contains(out, "No active grants.") {
		t.Errorf("expected empty grants list, got: %s", out)
	}
}

func TestPrintUsage(t *testing.T) {
	var buf bytes.Buffer
	printUsage(&buf)
	out := buf.String()
	if !strings.Contains(out, "symguard") {
		t.Error("expected 'symguard' in usage")
	}
	if !strings.Contains(out, "Commands:") {
		t.Error("expected 'Commands:' in usage")
	}
	if !strings.Contains(out, "version") {
		t.Error("expected 'version' in usage")
	}
	if !strings.Contains(out, "doctor") {
		t.Error("expected 'doctor' in usage")
	}
	if !strings.Contains(out, "grants") {
		t.Error("expected 'grants' in usage")
	}
}

func TestCmdVersion(t *testing.T) {
	var buf bytes.Buffer
	version.Run(nil, &buf)
	out := buf.String()
	if !strings.Contains(out, "symguard") {
		t.Error("expected 'symguard' in version output")
	}
	if !strings.Contains(out, "go") {
		t.Error("expected go version in output")
	}
}

func TestCmdDoctor(t *testing.T) {
	var buf bytes.Buffer
	doctor.Run(&buf)
	out := buf.String()
	if !strings.Contains(out, "symguard doctor") {
		t.Error("expected doctor header")
	}
	if !strings.Contains(out, "binary") {
		t.Error("expected 'binary' check")
	}
	if !strings.Contains(out, "config") {
		t.Error("expected 'config' check")
	}
}

func TestBuildTime(t *testing.T) {
	// buildTime is an unexported function in the version package.
	// Test that version output includes the placeholder.
	var buf bytes.Buffer
	version.Run(nil, &buf)
	out := buf.String()
	if !strings.Contains(out, "compile-time placeholder") {
		t.Errorf("expected placeholder in version output, got: %s", out)
	}
}

func TestRun_VersionJSON(t *testing.T) {
	var buf bytes.Buffer
	code := run([]string{"version", "--json"}, &buf)
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
	out := buf.String()
	if !strings.Contains(out, `"tool"`) {
		t.Error("expected JSON output with 'tool' field")
	}
	if !strings.Contains(out, `"version"`) {
		t.Error("expected JSON output with 'version' field")
	}
	if !strings.Contains(out, `"schema_version"`) {
		t.Error("expected JSON output with 'schema_version' field")
	}
}

func TestCmdVersionJSON(t *testing.T) {
	var buf bytes.Buffer
	version.Run([]string{"--json"}, &buf)
	out := buf.String()
	if !strings.Contains(out, `"tool":"symguard"`) {
		t.Errorf("expected JSON tool field, got: %s", out)
	}
	if !strings.Contains(out, `"schema_version":1`) {
		t.Errorf("expected schema_version 1, got: %s", out)
	}
}

func TestCmdVersionPlainIsStringFormat(t *testing.T) {
	var buf bytes.Buffer
	version.Run(nil, &buf)
	out := buf.String()
	// Info.String() returns "tool vX.Y.Z"
	if !strings.Contains(out, "symguard") {
		t.Error("expected 'symguard' in output")
	}
	if !strings.Contains(out, "go") {
		t.Error("expected go version in output")
	}
}
