package main

import (
	"bytes"
	"strings"
	"testing"
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
}

func TestCmdVersion(t *testing.T) {
	var buf bytes.Buffer
	cmdVersion(&buf)
	out := buf.String()
	if !strings.Contains(out, "symguard") {
		t.Error("expected 'symguard' in version output")
	}
	if version == "" {
		t.Error("version should not be empty")
	}
	if !strings.Contains(out, version) {
		t.Errorf("expected version %q in output", version)
	}
}

func TestCmdDoctor(t *testing.T) {
	var buf bytes.Buffer
	cmdDoctor(&buf)
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
	bt := buildTime()
	if bt == "" {
		t.Error("buildTime should not return empty string")
	}
	if !strings.Contains(bt, "compile-time placeholder") {
		t.Errorf("expected placeholder in buildTime, got: %s", bt)
	}
}
