// Package update provides update checking for symguard using corekit/updatecheck.
package update

import (
	"context"
	"fmt"
	"time"

	"github.com/danieljustus/symaira-corekit/updatecheck"
)

const (
	owner = "danieljustus"
	repo  = "symaira-guard"
)

// Info holds the result of an update check.
type Info struct {
	Current string // installed version
	Latest  string // latest release tag (empty if up to date)
	URL     string // release URL (empty if up to date)
}

// Check queries the latest GitHub release and returns update info.
// Errors are silently swallowed — never block or disrupt the CLI.
func Check(ctx context.Context, currentVersion string) *Info {
	if currentVersion == "" || currentVersion == "dev" {
		return nil
	}

	checker := updatecheck.NewChecker(owner, repo)
	checker.CacheTTL = 24 * time.Hour

	release, err := checker.Check(ctx, currentVersion)
	if err != nil {
		return nil // silently swallow errors
	}
	if release == nil {
		return nil // up to date
	}

	return &Info{
		Current: currentVersion,
		Latest:  release.TagName,
		URL:     release.HTMLURL,
	}
}

// Format returns a human-readable update message for stderr,
// or empty string if the version is current.
func Format(info *Info) string {
	if info == nil {
		return ""
	}
	return fmt.Sprintf("symguard: update available: %s → %s\n  %s\n",
		info.Current, info.Latest, info.URL)
}
