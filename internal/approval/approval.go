// Package approval defines the data contract for pending approval requests
// and human decisions in symguard's approval layer.
//
// The key design decision is embedded correlation: the ApprovalRequest
// carries the full original tool call (the ActionEvent from internal/model),
// so the frontend (TUI, CLI, or browser) never needs its own lookup state
// to know what it is approving. The ApprovalDecision echoes the same ID
// for correlation.
//
// This avoids a separate pending-request database or lookup table — the
// request IS the state, and matching on the echoed ID is sufficient for
// single-process, local-first operation.
package approval

import (
	"time"

	"github.com/danieljustus/symaira-guard/internal/model"
)

// Request is sent to the human for a decision on a pending tool call.
type Request struct {
	ID           string          `json:"id"`
	Hint         string          `json:"hint"`
	OriginalCall model.ActionEvent `json:"original_call"`
	Payload      any             `json:"payload,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	TTL          time.Duration   `json:"ttl,omitempty"`
}

// Decision is the human's response to an approval request.
// The ID must echo the Request.ID for correlation.
type Decision struct {
	ID        string        `json:"id"`
	Approved  bool          `json:"approved"`
	Payload   any           `json:"payload,omitempty"`
	Reason    string        `json:"reason,omitempty"`
	TTL       time.Duration `json:"ttl,omitempty"`
	DecidedAt time.Time     `json:"decided_at"`
}
