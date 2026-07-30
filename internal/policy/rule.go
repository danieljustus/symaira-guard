// Package policy defines the versioned rule catalog and evaluation engine
// for symguard's policy subsystem.
//
// Rules are evaluated in order of precedence (lower number = higher priority).
// The first matching rule whose Decision is not a pass-through wins.
// If no rule matches, the capability-based Defaults from config apply.
//
// The policy package is independent from MCP transport, UI, and audit storage.
// It consumes model types for events and decisions but does not import
// transport or storage packages.
package policy

import (
	"fmt"

	"github.com/danieljustus/symaira-guard/internal/model"
)

// RuleID is a stable, unique rule identifier.
type RuleID string

// Version is a semver-compatible rule version.
type Version string

// Precedence defines rule ordering (lower = evaluated first).
type Precedence int

// Rule is a single policy rule with stable identity and deterministic ordering.
type Rule struct {
	ID         RuleID          `json:"id"`
	Version    Version         `json:"version"`
	Precedence Precedence      `json:"precedence"`
	Decision   model.Decision  `json:"decision"`
	Match      MatchCriteria   `json:"match"`
	Reason     string          `json:"reason,omitempty"`
	ObserveOnly bool           `json:"observe_only,omitempty"`
}

// MatchCriteria defines what a rule matches against.
// All non-empty fields must match (AND semantics).
type MatchCriteria struct {
	Server          string   `json:"server,omitempty"`
	Tool            string   `json:"tool,omitempty"`
	Capability      string   `json:"capability,omitempty"`
	Remote          string   `json:"remote,omitempty"`
	CommandContains []string `json:"command_contains,omitempty"`
}

// Result is the structured outcome of a policy evaluation.
type Result struct {
	Rule        *Rule          `json:"rule,omitempty"`
	Decision    model.Decision `json:"decision"`
	Reason      string         `json:"reason,omitempty"`
	Matched     bool           `json:"matched"`
	Precedence  Precedence     `json:"precedence,omitempty"`
}

// Catalog is a versioned, validated collection of rules.
type Catalog struct {
	Rules      []Rule    `json:"rules"`
	Version    Version   `json:"version"`
}

// Evaluate checks a tool call against the catalog and returns the
// first matching rule's decision. Returns the default decision from
// the provided Defaults if no rule matches.
func (c *Catalog) Evaluate(call model.ToolCall, defaults model.Decision) Result {
	for _, rule := range c.Rules {
		if matches(rule.Match, call) {
			return Result{
				Rule:       &rule,
				Decision:   rule.Decision,
				Reason:     rule.Reason,
				Matched:    true,
				Precedence: rule.Precedence,
			}
		}
	}
	return Result{
		Decision: defaults,
		Reason:   "default capability decision",
		Matched:  false,
	}
}

// matches checks whether a tool call satisfies the match criteria.
func matches(m MatchCriteria, call model.ToolCall) bool {
	if m.Server != "" && m.Server != call.Server {
		return false
	}
	if m.Tool != "" && m.Tool != call.Tool {
		return false
	}
	if m.Capability != "" && m.Capability != call.Capability {
		return false
	}
	if len(m.CommandContains) > 0 {
		args, ok := call.Args.(string)
		if !ok {
			return false
		}
		matched := false
		for _, substr := range m.CommandContains {
			if contains(args, substr) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	return true
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Validate checks the catalog for structural errors.
func (c *Catalog) Validate() error {
	seen := make(map[RuleID]bool)
	prevPrecedence := Precedence(-1)

	for i, rule := range c.Rules {
		if rule.ID == "" {
			return fmt.Errorf("rules[%d]: empty rule ID", i)
		}
		if seen[rule.ID] {
			return fmt.Errorf("rules[%d]: duplicate rule ID %q", i, rule.ID)
		}
		seen[rule.ID] = true

		if err := model.ValidateDecision(rule.Decision); err != nil {
			return fmt.Errorf("rules[%d] %q: %w", i, rule.ID, err)
		}

		if rule.Match.Server == "" && rule.Match.Tool == "" &&
			rule.Match.Capability == "" && rule.Match.Remote == "" &&
			len(rule.Match.CommandContains) == 0 {
			return fmt.Errorf("rules[%d] %q: match must specify at least one criterion", i, rule.ID)
		}

		if rule.Precedence <= prevPrecedence {
			return fmt.Errorf("rules[%d] %q: precedence %d not strictly greater than previous %d",
				i, rule.ID, rule.Precedence, prevPrecedence)
		}
		prevPrecedence = rule.Precedence
	}
	return nil
}

// NewCatalog creates a validated catalog from rules.
func NewCatalog(rules []Rule, version Version) (*Catalog, error) {
	c := &Catalog{Rules: rules, Version: version}
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return c, nil
}
