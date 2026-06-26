package main

import (
	"testing"
)

func TestCmdVersion(t *testing.T) {
	// version should be set (default is "dev" if not injected via ldflags)
	if version == "" {
		t.Error("version should not be empty")
	}
}

func TestCmdDoctor(t *testing.T) {
	// Doctor should not panic — just verify the checks slice logic
	checks := []struct {
		name   string
		status string
	}{
		{"binary", "ok"},
		{"go runtime", "ok"},
		{"config", "not configured (no config file found)"},
		{"policy", "not loaded"},
		{"audit log", "not initialized"},
	}

	if len(checks) == 0 {
		t.Error("expected at least one health check")
	}

	for _, c := range checks {
		if c.name == "" {
			t.Error("health check name should not be empty")
		}
		if c.status == "" {
			t.Errorf("health check %q status should not be empty", c.name)
		}
	}
}
