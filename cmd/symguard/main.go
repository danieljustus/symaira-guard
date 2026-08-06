// Package main is the CLI entrypoint for symguard.
//
// symguard is a local-first security gateway for AI agents,
// MCP servers, and Symaira toolchains.
//
// Each subcommand lives in its own package under cmd/symguard/<verb>/
// and exposes a Run([]string, io.Writer) int entry point.
package main

import (
	"fmt"
	"io"
	"os"

	"github.com/danieljustus/symaira-guard/cmd/symguard/doctor"
	"github.com/danieljustus/symaira-guard/cmd/symguard/grants"
	"github.com/danieljustus/symaira-guard/cmd/symguard/version"
)

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
		version.Run(args[1:], w)
	case "doctor":
		doctor.Run(w)
	case "grants":
		grants.Run(args[1:], w)
	case "help", "--help", "-h":
		printUsage(w)
	default:
		fmt.Fprintf(w, "unknown command: %s\n\n", args[0])
		printUsage(w)
		return 1
	}

	// Non-blocking update check after every command.
	version.CheckLatest()

	return 0
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, `symguard — local-first security gateway for AI agents

Usage:
  symguard <command> [flags]

Commands:
  version   Print version and build info
  doctor    Check system health and configuration
  grants    List and revoke standing grants
  help      Show this help message

Run 'symguard <command> --help' for details on a specific command.`)
}
