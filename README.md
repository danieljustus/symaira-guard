# Symaira Guard (`symguard`)

> A local-first security gateway for AI agents, MCP servers, and Symaira toolchains.

**Human control for agent autonomy.**

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

## What it does

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

Decisions: `allow`, `ask`, `deny`, `redact`, `readonly`, `sandbox`.

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

### 6. Remote access

Later phases add agent-aware remote MCP access over existing transports (SSH, Tailscale, LAN/mDNS) — not a new VPN, but policy and audit on top of tools you already trust.

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

## Status

Early development. See [docs/intern/IDEA.md](docs/intern/IDEA.md) for the full design document.
