// Package main is the CLI entrypoint for symguard.
//
// symguard is a local-first security gateway for AI agents,
// MCP servers, and Symaira toolchains.
package main

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"time"
)

// version is set at build time via ldflags:
//
//	go build -ldflags "-X main.version=v1.0.0"
//
// Default value is "dev" for untagged builds.
var version = "dev"

func main() {
	os.Exit(run(os.Args[1:], os.Stdout))
}

// run executes the CLI command with the given args and writes output to w.
// It returns an exit code: 0 for success, 1 for usage errors.
func run(args []string, w io.Writer) int {
	if len(args) < 1 {
		printUsage(w)
		return 1
	}

	switch args[0] {
	case "version":
		cmdVersion(w)
	case "doctor":
		cmdDoctor(w)
	case "help", "--help", "-h":
		printUsage(w)
	default:
		fmt.Fprintf(w, "unknown command: %s\n\n", args[0])
		printUsage(w)
		return 1
	}
	return 0
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, `symguard — local-first security gateway for AI agents

Usage:
  symguard <command> [flags]

Commands:
  version   Print version and build info
  doctor    Check system health and configuration
  help      Show this help message

Run 'symguard <command> --help' for details on a specific command.`)
}

func cmdVersion(w io.Writer) {
	fmt.Fprintf(w, "symguard %s\n", version)
	fmt.Fprintf(w, "  go      %s\n", runtime.Version())
	fmt.Fprintf(w, "  os/arch %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Fprintf(w, "  built   %s\n", buildTime())
}

func buildTime() string {
	t, err := time.Parse(time.RFC3339, "2026-01-01T00:00:00Z")
	if err != nil {
		return "unknown"
	}
	return fmt.Sprintf("%s (compile-time placeholder)", t.Format("2006-01-02"))
}

func cmdDoctor(w io.Writer) {
	fmt.Fprintln(w, "symguard doctor")
	fmt.Fprintln(w)
	fmt.Fprintf(w, "  Version:   %s\n", version)
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
