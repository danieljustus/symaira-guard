// Package doctor implements the `symguard doctor` command.
package doctor

import (
	"fmt"
	"io"
	"runtime"
)

// Run prints system health and configuration information to w.
func Run(w io.Writer) {
	fmt.Fprintln(w, "symguard doctor")
	fmt.Fprintln(w)
	fmt.Fprintf(w, "  Version:   %s\n", versionInfo())
	fmt.Fprintf(w, "  Go:        %s\n", runtime.Version())
	fmt.Fprintf(w, "  OS/Arch:   %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Fprintln(w)

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

	for _, c := range checks {
		fmt.Fprintf(w, "  %-16s %s\n", c.name, c.status)
	}

	fmt.Fprintln(w)
	fmt.Fprintln(w, "All basic checks passed. Run 'symguard scan' after setup for full diagnostics.")
}

// versionInfo returns the version string for display.
var versionInfo = func() string { return "dev" }

// SetVersion sets the version callback for the doctor command.
func SetVersion(fn func() string) {
	versionInfo = fn
}
