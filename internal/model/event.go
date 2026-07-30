// Package model defines the versioned, redaction-safe event contract for
// symguard's policy evaluation, approval flow, MCP proxy, and audit log.
//
// Schema version: 1 (increment on any backward-incompatible change to any type
// in this package). Additive changes (new source types, new decision values,
// new optional fields) extend the schema but do not change the version.
//
// # Core invariant
//
// Every event has a stable ID and enough provenance to correlate policy
// evaluation, approval, execution, and audit records. Sensitive arguments,
// tool output, paths, and credentials are redacted or represented by safe
// references before leaving the control boundary.
//
// # Control / Diagnostic separation
//
// ControlResponse contains only what the MCP proxy or agent runtime needs:
// allow, deny, ask, or a passthrough signal. All diagnostic context (rule
// expression, matched rule, evaluation trace) belongs in the ActionEvent's
// Evaluation field and is written to stderr or the audit sink, never to
// stdout in MCP transport mode.
package model

import "fmt"

// SchemaVersion is the current version of the ActionEvent schema.
// Increment on any backward-incompatible change; additive changes are
// compatible and do not change this constant.
const SchemaVersion = 1

// SourceType identifies where an ActionEvent originated.
type SourceType string

const (
	SourceProxy    SourceType = "proxy"    // MCP proxy intercept
	SourceHook     SourceType = "hook"     // Pre/post execution hook
	SourceArtifact SourceType = "artifact" // Declared tool or script output
	SourceScan     SourceType = "scan"     // Offline tool-scanning report
)

// ActionState describes the lifecycle phase of a tool call.
type ActionState string

const (
	ActionRequested ActionState = "requested"
	ActionApproved  ActionState = "approved"
	ActionDenied    ActionState = "denied"
	ActionStarted   ActionState = "started"
	ActionCompleted ActionState = "completed"
	ActionFailed    ActionState = "failed"
)

// RiskClass is a static classification assigned to a capability name.
type RiskClass string

const (
	RiskClassLow      RiskClass = "low"
	RiskClassMedium   RiskClass = "medium"
	RiskClassHigh     RiskClass = "high"
	RiskClassCritical RiskClass = "critical"
)

// Decision is the policy outcome for a tool call.
type Decision string

const (
	DecisionAllow    Decision = "allow"
	DecisionAsk      Decision = "ask"
	DecisionDeny     Decision = "deny"
	DecisionRedact   Decision = "redact"
	DecisionReadOnly Decision = "readonly"
	DecisionSandbox  Decision = "sandbox"
)

// AgentIdentity identifies the calling agent or client.
type AgentIdentity struct {
	AgentID   string `json:"agent_id"`
	SessionID string `json:"session_id,omitempty"`
	RunID     string `json:"run_id,omitempty"`
}

// ClientIdentity identifies the human or system that configured the agent.
type ClientIdentity struct {
	UserID string `json:"user_id,omitempty"`
	Host   string `json:"host,omitempty"`
}

// ToolCall captures an MCP tool request.
type ToolCall struct {
	Server     string `json:"server"`
	Tool       string `json:"tool"`
	Args       any    `json:"args,omitempty"`       // redacted before leaving the control boundary
	Result     any    `json:"result,omitempty"`      // redacted before leaving the control boundary
	ArgsRef    string `json:"args_ref,omitempty"`    // safe reference to the raw args
	ResultRef  string `json:"result_ref,omitempty"`  // safe reference to the raw result
	Capability string `json:"capability,omitempty"`  // risk-class key (e.g. "shell", "read_secret")
	RiskClass  string `json:"risk_class,omitempty"`  // resolved risk classification
	Error      string `json:"error,omitempty"`       // non-empty when the call failed
}

// Evaluation records the policy engine's reasoning (diagnostic only).
type Evaluation struct {
	MatchedRule   string   `json:"matched_rule,omitempty"`
	Decision      Decision `json:"decision"`
	Reason        string   `json:"reason,omitempty"`
	MarginalCheck bool     `json:"marginal_check,omitempty"`
}

// ActionEvent is the versioned, immutable record of one policy-relevant action.
type ActionEvent struct {
	ID           string         `json:"id"`
	SchemaVer    int            `json:"schema_version"`
	Source       SourceType     `json:"source"`
	PrevEventID  string         `json:"prev_event_id,omitempty"` // chain for audit log
	Agent        AgentIdentity  `json:"agent"`
	Client       ClientIdentity `json:"client,omitempty"`
	Call         ToolCall       `json:"call"`
	State        ActionState    `json:"state"`
	Evaluation   *Evaluation    `json:"evaluation,omitempty"`   // diagnostic only
	ControlResp  *ControlResponse `json:"control_response,omitempty"`
	Timestamp    string         `json:"timestamp"`              // RFC 3339
}

// ControlResponse is the immediate action the runtime should take.
// It contains zero diagnostic context.
type ControlResponse struct {
	Decision   Decision `json:"decision"`
	Reason     string   `json:"reason,omitempty"`      // user-facing, safe
	RetryAfter int      `json:"retry_after,omitempty"` // seconds
}

// EventID generates a stable, locally unique event ID.
// Format: "evt_<schema>_<source>_<counter>".
func EventID(source SourceType, counter int64) string {
	return fmt.Sprintf("evt_%d_%s_%d", SchemaVersion, source, counter)
}

// ValidateSource returns an error if the source type is unrecognized.
func ValidateSource(s SourceType) error {
	switch s {
	case SourceProxy, SourceHook, SourceArtifact, SourceScan:
		return nil
	}
	return fmt.Errorf("model: unknown source type %q", s)
}

// ValidateState returns an error if the action state is unrecognized.
func ValidateState(s ActionState) error {
	switch s {
	case ActionRequested, ActionApproved, ActionDenied,
		ActionStarted, ActionCompleted, ActionFailed:
		return nil
	}
	return fmt.Errorf("model: unknown action state %q", s)
}

// ValidateDecision returns an error if the decision value is unrecognized.
func ValidateDecision(d Decision) error {
	switch d {
	case DecisionAllow, DecisionAsk, DecisionDeny,
		DecisionRedact, DecisionReadOnly, DecisionSandbox:
		return nil
	}
	return fmt.Errorf("model: unknown decision %q", d)
}
