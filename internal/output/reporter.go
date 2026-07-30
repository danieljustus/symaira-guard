// Package output provides a Reporter interface and format-specific
// implementations for rendering symguard results.
//
// Supported formats:
//   - "table" (default): human-readable tabular output
//   - "json": JSON-formatted output
//
// The package has no dependency on scan, discovery, or other subsystem
// internals — it only knows how to render a result value it's given.
package output

import (
	"encoding/json"
	"fmt"
	"io"
)

// Reporter renders a result value to the given writer.
type Reporter interface {
	// Print renders the result to w and returns any error.
	Print(w io.Writer, result any) error
}

// NewReporter returns a Reporter for the given format string.
// Supported formats: "table" (default), "json".
func NewReporter(format string) Reporter {
	switch format {
	case "json":
		return &jsonReporter{}
	case "table", "":
		return &tableReporter{}
	default:
		return &tableReporter{}
	}
}

type jsonReporter struct{}

func (r *jsonReporter) Print(w io.Writer, result any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}

type tableReporter struct{}

func (r *tableReporter) Print(w io.Writer, result any) error {
	// For structured results, attempt a simple key-value rendering.
	// Future implementations may use richer formatting for known types.
	switch v := result.(type) {
	case string:
		_, err := fmt.Fprintln(w, v)
		return err
	case fmt.Stringer:
		_, err := fmt.Fprintln(w, v.String())
		return err
	default:
		// Fall back to compact JSON for unknown types.
		enc := json.NewEncoder(w)
		enc.SetIndent("", "")
		return enc.Encode(v)
	}
}

// FormatNames returns the list of supported format names.
func FormatNames() []string {
	return []string{"table", "json"}
}
