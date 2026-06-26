// Package config defines the TOML configuration schema for symguard and
// provides a loader that resolves XDG Base Directory paths with environment
// variable overrides.
//
// The configuration file lives at:
//
//	$XDG_CONFIG_HOME/symguard/config.toml
//
// with a fallback to ~/.config/symguard/config.toml when XDG_CONFIG_HOME is
// unset. The SYMGUARD_CONFIG environment variable overrides both.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Decision represents an allowed policy decision for a tool call.
type Decision string

const (
	Allow    Decision = "allow"
	Ask      Decision = "ask"
	Deny     Decision = "deny"
	Redact   Decision = "redact"
	ReadOnly Decision = "readonly"
	Sandbox  Decision = "sandbox"
)

// Defaults maps capability names to default decisions. Keys correspond to
// risk classes defined in IDEA.md (e.g. "shell", "read_secret", "write_file",
// "network").
type Defaults map[string]Decision

// RuleMatch defines the matching criteria for a policy rule. At least one
// field must be set. Multiple fields are ANDed together.
type RuleMatch struct {
	Server           string   `toml:"server,omitempty"`
	Tool             string   `toml:"tool,omitempty"`
	Capability       string   `toml:"capability,omitempty"`
	CommandContains  []string `toml:"command_contains,omitempty"`
}

// Rule maps a match pattern to a policy decision. Rules are evaluated in
// order; the first matching rule wins.
type Rule struct {
	Match    RuleMatch `toml:"match"`
	Decision Decision  `toml:"decision"`
}

// ProxyConfig holds upstream MCP server configuration for proxy mode.
type ProxyConfig struct {
	Upstream string `toml:"upstream,omitempty"`
}

// AuditConfig controls the append-only audit log.
type AuditConfig struct {
	Path       string `toml:"path,omitempty"`
	Encrypt    bool   `toml:"encrypt,omitempty"`
	EncryptAge string `toml:"encrypt_age,omitempty"`
}

// RemoteTarget describes a known remote MCP target.
type RemoteTarget struct {
	Name           string   `toml:"name"`
	Provider       string   `toml:"provider"`
	Host           string   `toml:"host"`
	AllowedServers []string `toml:"allowed_servers,omitempty"`
	TrustLevel     string   `toml:"trust_level,omitempty"`
	Labels         []string `toml:"labels,omitempty"`
}

// Config is the top-level TOML configuration structure for symguard.
type Config struct {
	Defaults Defaults `toml:"defaults"`
	Rules    []Rule   `toml:"rules"`
	Proxy    ProxyConfig `toml:"proxy"`
	Audit    AuditConfig `toml:"audit"`
	Remote   []RemoteTarget `toml:"remote"`
}

// DefaultConfig returns a Config with sensible defaults. When no config file
// is present, Load returns this value.
func DefaultConfig() *Config {
	return &Config{
		Defaults: Defaults{
			"shell":       Ask,
			"read_secret": Deny,
			"write_file":  Ask,
			"network":     Ask,
		},
		Rules:  nil,
		Proxy:  ProxyConfig{},
		Audit: AuditConfig{
			Path: "symguard-audit.log",
		},
		Remote: nil,
	}
}

// DefaultPath returns the XDG Base Directory path for the config file:
//
//	$XDG_CONFIG_HOME/symguard/config.toml
//
// When XDG_CONFIG_HOME is unset, it falls back to ~/.config/symguard/config.toml.
func DefaultPath() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "symguard", "config.toml")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		// Last resort: relative path.
		return filepath.Join(".config", "symguard", "config.toml")
	}
	return filepath.Join(home, ".config", "symguard", "config.toml")
}

// ConfigPath resolves the config file path. It checks SYMGUARD_CONFIG first,
// then falls back to the XDG default path.
func ConfigPath() string {
	if env := os.Getenv("SYMGUARD_CONFIG"); env != "" {
		return env
	}
	return DefaultPath()
}

// Load reads the TOML configuration from the resolved path. When the file
// does not exist it returns DefaultConfig with no error. When the file exists
// but contains invalid TOML or schema violations, it returns a descriptive
// error.
func Load() (*Config, error) {
	return LoadFrom(ConfigPath())
}

// LoadFrom reads the TOML configuration from the given path. When the file
// does not exist it returns DefaultConfig with no error. When the file exists
// but contains invalid TOML, it returns a descriptive error.
func LoadFrom(path string) (*Config, error) {
	cfg := DefaultConfig()

	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return cfg, nil
	}
	if err != nil {
		return nil, fmt.Errorf("config: stat %s: %w", path, err)
	}

	meta, err := toml.DecodeFile(path, cfg)
	if err != nil {
		return nil, fmt.Errorf("config: parse %s: %w", path, err)
	}

	// Warn about unknown keys so users catch typos early.
	if undecoded := meta.Undecoded(); len(undecoded) > 0 {
		for _, key := range undecoded {
			fmt.Fprintf(os.Stderr, "config: warning: unknown key %q in %s\n", key, path)
		}
	}

	if err := validate(cfg); err != nil {
		return nil, fmt.Errorf("config: validate %s: %w", path, err)
	}

	return cfg, nil
}

// validate checks the decoded config for obvious policy errors.
func validate(cfg *Config) error {
	validDecisions := map[Decision]bool{
		Allow: true, Ask: true, Deny: true, Redact: true, ReadOnly: true, Sandbox: true,
	}

	for cap, d := range cfg.Defaults {
		if !validDecisions[d] {
			return fmt.Errorf("defaults.%q: invalid decision %q (allowed: allow, ask, deny, redact, readonly, sandbox)", cap, d)
		}
	}

	for i, rule := range cfg.Rules {
		if !validDecisions[rule.Decision] {
			return fmt.Errorf("rules[%d].decision: invalid decision %q", i, rule.Decision)
		}
		if rule.Match.Server == "" && rule.Match.Tool == "" && rule.Match.Capability == "" && len(rule.Match.CommandContains) == 0 {
			return fmt.Errorf("rules[%d]: match must specify at least one criterion (server, tool, capability, command_contains)", i)
		}
	}

	return nil
}
