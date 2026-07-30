// Package version implements the `symguard version` command.
package version

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"time"

	"github.com/danieljustus/symaira-guard/internal/update"
)

// version is set at build time via ldflags.
var version = "dev"

// Run prints version and build information to w.
func Run(w io.Writer) {
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

// CheckLatest queries GitHub and prints an update notice to stderr
// if a newer release is available. Errors are silently swallowed.
func CheckLatest() {
	info := update.Check(context.Background(), version)
	if msg := update.Format(info); msg != "" {
		fmt.Fprint(os.Stderr, msg)
	}
}

// SetVersion sets the version string at build time.
// Called from main via ldflags.
func SetVersion(v string) {
	version = v
}
