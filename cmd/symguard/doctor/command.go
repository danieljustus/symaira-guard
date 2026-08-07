// Package doctor implements the `symguard doctor` command.
//
// Doctor reports system health and configuration information, including the
// spawn allowlist gate for discovered stdio MCP servers and plaintext-secret
// risks in their configs. It reports and gates; it never becomes a secret
// store — resolution stays with symvault.
package doctor

import (
	"fmt"
	"io"
	"runtime"

	"github.com/danieljustus/symaira-guard/internal/config"
	"github.com/danieljustus/symaira-guard/internal/discovery"
	"github.com/danieljustus/symaira-guard/internal/spawn"
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

	// Spawn allowlist and discovered MCP servers: doctor reports and gates,
	// it does not become a secret store.
	var (
		serverChecks []ServerCheck
		issues       int
	)
	cfg, err := config.Load()
	if err != nil {
		checks = append(checks, struct{ name, status string }{"spawn allowlist", fmt.Sprintf("error: %v", err)})
	} else {
		allowlist := spawn.NewAllowlist(cfg.Spawn.Allowlist)
		if allowlist.Len() == 0 {
			checks = append(checks, struct{ name, status string }{"spawn allowlist", "not configured (empty — deny by default)"})
		} else {
			checks = append(checks, struct{ name, status string }{"spawn allowlist", fmt.Sprintf("ok (%d entries)", allowlist.Len())})
		}

		servers, discErr := discovery.DiscoverAll()
		switch {
		case discErr != nil:
			checks = append(checks, struct{ name, status string }{"mcp servers", fmt.Sprintf("error: %v", discErr)})
		case len(servers) == 0:
			checks = append(checks, struct{ name, status string }{"mcp servers", "none discovered"})
		default:
			checks = append(checks, struct{ name, status string }{"mcp servers", fmt.Sprintf("%d discovered", len(servers))})
			serverChecks = checkServers(servers, allowlist)
			issues = issueCount(serverChecks)
		}
	}

	for _, c := range checks {
		fmt.Fprintf(w, "  %-16s %s\n", c.name, c.status)
	}

	printServerChecks(w, serverChecks)
	printSecretRisks(w, serverChecks)

	fmt.Fprintln(w)
	if issues == 0 {
		fmt.Fprintln(w, "All basic checks passed. Run 'symguard scan' after setup for full diagnostics.")
	} else {
		fmt.Fprintf(w, "%d issue(s) found. See details above.\n", issues)
	}
}

// versionInfo returns the version string for display.
var versionInfo = func() string { return "dev" }

// SetVersion sets the version callback for the doctor command.
func SetVersion(fn func() string) {
	versionInfo = fn
}
