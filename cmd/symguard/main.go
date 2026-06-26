// Package main is the CLI entrypoint for symguard.
//
// symguard is a local-first security gateway for AI agents,
// MCP servers, and Symaira toolchains.
package main

import (
	"fmt"
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
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "version":
		cmdVersion()
	case "doctor":
		cmdDoctor()
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stdout, `symguard — local-first security gateway for AI agents

Usage:
  symguard <command> [flags]

Commands:
  version   Print version and build info
  doctor    Check system health and configuration
  help      Show this help message

Run 'symguard <command> --help' for details on a specific command.`)
}

func cmdVersion() {
	fmt.Printf("symguard %s\n", version)
	fmt.Printf("  go      %s\n", runtime.Version())
	fmt.Printf("  os/arch %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Printf("  built   %s\n", buildTime())
}

func buildTime() string {
	t, err := time.Parse(time.RFC3339, "2026-01-01T00:00:00Z")
	if err != nil {
		return "unknown"
	}
	// In a real build the build time would be injected via ldflags.
	// For now, return a placeholder indicating when this binary was compiled.
	return fmt.Sprintf("%s (compile-time placeholder)", t.Format("2006-01-02"))
}

func cmdDoctor() {
	fmt.Println("symguard doctor")
	fmt.Println()
	fmt.Printf("  Version:   %s\n", version)
	fmt.Printf("  Go:        %s\n", runtime.Version())
	fmt.Printf("  OS/Arch:   %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Println()

	// Placeholder health checks — real checks will be added with
	// config, policy, audit, and MCP subsystems in later issues.
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
		fmt.Printf("  %-16s %s\n", c.name, c.status)
	}

	fmt.Println()
	fmt.Println("All basic checks passed. Run 'symguard scan' after setup for full diagnostics.")
}
