# Symaira Guard (`symguard`)

> A local-first security gateway for AI agents, MCP servers, and Symaira toolchains.

**Human control for agent autonomy.**

[![CI](https://github.com/danieljustus/symaira-guard/actions/workflows/ci.yml/badge.svg)](https://github.com/danieljustus/symaira-guard/actions/workflows/ci.yml)
[![Go](https://img.shields.io/github/go-mod/go-version/danieljustus/symaira-guard)](https://go.dev/)
[![License](https://img.shields.io/github/license/danieljustus/symaira-guard)](https://github.com/danieljustus/symaira-guard/blob/main/LICENSE)

---

## What

`symguard` sits between AI clients and the tools they call. It inspects every MCP tool call, classifies risk, enforces local policy, asks for human approval when needed, and records tamper-evident audit trails.

```
AI client / agent  →  symguard  →  MCP servers / CLIs / APIs / Symaira tools
```

The agent still gets useful tools. The human keeps enforceable boundaries.

## Why

MCP solved interoperability between AI clients and tool servers. It did not solve:

- Tool poisoning and rug-pull attacks (changed tool descriptions)
- Prompt injection escalating into tool calls
- Unbounded shell / filesystem / network / secret access
- Missing human approval for risky operations
- Secret exfiltration through tool output
- Cross-agent delegation risk

`symguard` is the missing local control layer.

## Current status (implemented today)

`symguard` is early-stage. The CLI currently implements only:

```bash
$ symguard version
symguard dev
  go      go1.26
  os/arch darwin/arm64
  built   2026-01-01 (compile-time placeholder)

$ symguard doctor
symguard doctor
...
  binary           ok
  go runtime       ok
  config           not configured (no config file found)
  policy           not loaded
  audit log        not initialized
  spawn allowlist  not configured (empty — deny by default)
  mcp servers      none discovered

All basic checks passed. Run 'symguard scan' after setup for full diagnostics.
```

`doctor` prints static health checks plus two live diagnostics: the
[spawn allowlist](docs/config/spawn-allowlist.md) verdict for every
discovered stdio MCP server, and plaintext-secret risks in their configs.
It reports and gates — it is not a secret store (resolution stays with
`symvault`). There is no `scan`, `policy`, `proxy`, `pin`, `audit`, or
`remote` subcommand yet — everything below this point is
**design intent, not shipped behavior**. Two internal packages exist to
support this direction: `internal/config` (TOML schema for defaults/rules,
not yet wired into the CLI) and `internal/discovery` (parses MCP config files
from Hermes/Claude Desktop/Cursor/VS Code/OpenCode, not yet exposed via any
command).

## What it does (planned)

### 1. Scan

Discover MCP servers configured across local AI clients and classify their tools by risk.

```bash
symguard scan                        # scan all clients
symguard scan --client hermes         # scan one client
symguard scan --format json           # machine-readable output
```

### 2. Policy

Define local rules that decide what gets through:

```toml
[defaults]
shell = "ask"
read_secret = "deny"
write_file = "ask"

[[rules]]
match.server = "symmemory"
match.tool = "memory_search"
decision = "allow"

[[rules]]
match.command_contains = ["rm -rf", "curl | sh"]
decision = "deny"
```

Decisions: `allow`, `ask`, `deny`, `redact`, `readonly`, `sandbox`. The TOML
schema for this exists in `internal/config`, but nothing evaluates it yet.

### 3. Proxy

Run as an MCP proxy that enforces policy per tool call:

```bash
symguard proxy --config ~/.config/symguard/config.toml
```

Each tool call is classified, policy-checked, optionally approved by a human, then forwarded upstream. Sensitive output can be redacted before it reaches the agent.

### 4. Pin

Store hashes of MCP tool descriptions and schemas. If a tool's description changes (hidden instructions, scope expansion), `symguard` flags it:

```
WARNING: Tool schema changed for server "filesystem" tool "read_file".
Policy: require re-approval
```

### 5. Audit

Append-only local audit log with hash chaining. Records what was requested, which policy matched, who approved, what executed, and what came back.

Hash chaining on its own only guards against **modifying** a retained entry —
it does not stop **truncating** the log (deleting the most recent entries
outright; a chain check over what remains still passes, because nothing
anchors how many entries should exist). `internal/audit` implements both
guarantees:

- **Modification detection** — the hash chain.
- **Truncation detection** — an external `ChainAnchor` checkpoint file
  (`.anchor`) records the last entry hash and total entry count after every
  write. On verification the anchor's count must match the observed entries.
  The anchor file should be stored with stricter permissions than the log
  itself; on Linux, `chattr +a` (append-only) prevents its deletion.

Until the anchor file is stored on a separate volume or with OS-level
append-only protection, truncation detection is robust against casual
tampering but not against a privileged attacker who can delete both files.
Document this limitation in deployments that accept it.

### 6. Remote access

Later phases add agent-aware remote MCP access over existing transports (SSH, Tailscale, LAN/mDNS) — not a new VPN, but policy and audit on top of tools you already trust.

None of sections 2–6 above are implemented yet — no policy engine, proxy,
pinning, audit log, or remote-access code exists in this repository today.

## Risk classes

| Risk | Examples | Default |
|------|----------|---------|
| `read_public` | docs, README, public web | allow |
| `read_private` | repo files, notes, local docs | allow or ask |
| `read_secret` | `.env`, SSH keys, vault entries | ask / deny |
| `write_file` | patch, overwrite, create file | ask |
| `shell` | command execution | ask |
| `network` | outbound API / web requests | ask |
| `browser` | cookies, sessions, web automation | ask |
| `credential_use` | using secrets without revealing them | ask once / scoped |
| `deploy` | release, push, infra mutation | ask every time |
| `destructive` | delete, wipe, reset, revoke | ask / deny |

The table above classifies a tool by its **name** alone — a necessary first
pass, but not sufficient on its own. `scan`'s risk classifier additionally
caps a tool's risk *downward* (never up) when the tool grants **zero marginal
capability** beyond one the same client already holds via another tool
already resolved to `allow` in that session. Example: a `read_file` tool
classifies as `read_private` above, but if the same client already has an
unrestricted `shell` tool resolved to `allow`, `read_file` grants no
capability `shell` didn't already grant — cat is a strict subset of shell —
so `scan` caps it down to `allow` and states why
(`no marginal capability over already-allowed tool: shell`). If `shell`
itself is `ask` or `deny`, `read_file` keeps its `read_private` classification
unchanged.

## Symaira ecosystem position

`symguard` is a **public, self-hosted core** tool. No Pro, tenant, or billing code.

```
┌─────────────────────────────────────────┐
│ AI clients / agents                     │
│ Hermes · Claude · Cursor · OpenCode ... │
└───────────────────┬─────────────────────┘
                    ▼
            ┌──────────────┐
            │  symguard    │  ← trust boundary
            └──────┬───────┘
                   ▼
    symvault · symmemory · symscope · symseek · ...
```

Optional runtime integrations, no compile-time dependencies on siblings.

## Principles

- **Local-first.** Policy decisions happen on your machine. No mandatory cloud account.
- **Boring is good.** No custom VPN, no NAT traversal, no WireGuard daemon. Reuse existing transports.
- **Discovery ≠ trust.** Finding a remote MCP server never auto-implies permission.
- **Agent identities.** Agents and runs are first-class identities with TTL and scoped grants.
- **Explainable.** Every decision has a reason. Simulate before acting. Diagnose after failing.

## Non-goals

Not a chat frontend, not a SIEM, not a cloud-only SaaS, not a VPN replacement, not a full endpoint protection platform.

> Classify agent tool calls, enforce local policy, ask the human when needed, and record what happened.

---

## Install

```bash
go install github.com/danieljustus/symaira-guard/cmd/symguard@latest
```

Or build from source:

## Build

Requires Go 1.26+. No external dependencies — only the Go standard library.

```bash
# Build the binary
make build

# Run tests
make test

# Lint (golangci-lint or go vet fallback)
make lint

# Set a version string at build time
make build VERSION=v1.0.0
```

Or directly with `go`:

```bash
go build -ldflags "-X main.version=dev" -o symguard ./cmd/symguard
go vet ./...
go test ./...
```

## Usage

```bash
# Version info (human-readable)
symguard version

# Version info (machine-readable JSON for GUI tools)
symguard version --json

# System health check
symguard doctor
```

## Development

```bash
git clone https://github.com/danieljustus/symaira-guard.git
cd symaira-guard

# Build
go build -o symguard ./cmd/symguard

# Test
go test ./...

# Lint
go vet ./...
```

## Status

Very early development. Implemented: `version` and `doctor` CLI commands,
plus two internal library packages (`config` schema, MCP client `discovery`)
that are not yet wired into any command. Everything else in this README
(scan, policy engine, proxy, pinning, audit, remote access) is design intent
only. See [docs/intern/IDEA.md](docs/intern/IDEA.md) for the full design document.
