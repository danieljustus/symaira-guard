// Package audit defines the types for symguard's append-only audit log,
// evidence references, and portable security case bundles.
//
// Evidence references point to supporting data without copying it, using
// ref:// URIs. Case bundles package events, decisions, and audit records
// into a portable, redaction-safe directory structure.
package audit

// BundleSchemaVersion is the current version of the case bundle format.
const BundleSchemaVersion = 1

// EvidenceRef is a URI-style reference to supporting evidence without
// copying the raw data. Formats:
//
//	ref://file/path:line         local file and line number
//	ref://event/<event-id>       reference to another event in the case
//	ref://hash/<sha256>          content-addressed reference
type EvidenceRef string

// Manifest describes a portable security case bundle.
type Manifest struct {
	SchemaVersion int               `json:"schema_version"`
	CaseID        string            `json:"case_id"`
	CreatedAt     string            `json:"created_at"` // RFC 3339
	RecordCounts  map[string]int    `json:"record_counts"`
	Digests       map[string]string `json:"digests"` // SHA-256 per stream
	SourceRepo    string            `json:"source_repo,omitempty"`
	Unsigned      bool              `json:"unsigned"`
}

// CaseBundle is a portable collection of security evidence.
type CaseBundle struct {
	Manifest  Manifest   `json:"manifest"`
	Events    []Event    `json:"events,omitempty"`
	Decisions []Decision `json:"decisions,omitempty"`
}

// Event is a normalized audit event record.
type Event struct {
	ID         string        `json:"id"`
	SchemaVer  int           `json:"schema_version"`
	EventType  string        `json:"event_type"`
	AgentID    string        `json:"agent_id,omitempty"`
	Server     string        `json:"server,omitempty"`
	Tool       string        `json:"tool,omitempty"`
	Capability string        `json:"capability,omitempty"`
	Decision   string        `json:"decision"`
	Evidence   []EvidenceRef `json:"evidence,omitempty"`
	Redacted   bool          `json:"redacted"`
	Timestamp  string        `json:"timestamp"` // RFC 3339
}

// Decision is a portable approval decision record.
type Decision struct {
	ID        string `json:"id"`
	RequestID string `json:"request_id"`
	Approved  bool   `json:"approved"`
	Reason    string `json:"reason,omitempty"`
	DecidedAt string `json:"decided_at"` // RFC 3339
}

// NewManifest creates a manifest with default values for the current schema.
func NewManifest(caseID string) Manifest {
	return Manifest{
		SchemaVersion: BundleSchemaVersion,
		CaseID:        caseID,
		RecordCounts:  make(map[string]int),
		Digests:       make(map[string]string),
		Unsigned:      true,
	}
}
