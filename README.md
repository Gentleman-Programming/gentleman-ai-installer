# AI Gentle Stack

> **One command. Any agent. Any OS. The Gentleman AI ecosystem -- configured and ready.**

AI Gentle Stack takes whatever AI coding agent(s) you use and supercharges them with the Gentleman ecosystem: persistent memory, spec-driven development, coding skills, MCP servers, code review, and a teaching-oriented persona.

## What it does

This is NOT an AI agent installer. Most agents are easy to install. This is an **ecosystem configurator** -- it injects the full Gentleman stack into your AI tools:

- **Engram** -- persistent cross-session memory
- **SDD** -- Spec-Driven Development workflow (plan before you code)
- **Skills** -- curated coding patterns for modern stacks (React 19, Next.js 15, TypeScript, Tailwind 4, etc.)
- **MCP servers** -- Context7 for real-time library documentation
- **GGA** -- Guardian Angel AI code review on every commit
- **Persona & config** -- security-first permissions, teaching-oriented persona, themes

**Before**: "I installed Claude Code / OpenCode, but it's just a chatbot that writes code."

**After**: Your agent now has memory, skills, workflow, MCP tools, and a persona that actually teaches you.

## Supported platforms

| Platform | Package Manager | Status |
|----------|----------------|--------|
| macOS (Apple Silicon + Intel) | Homebrew | Supported |
| Linux (Ubuntu/Debian) | apt | Supported |
| Linux (Arch) | pacman | Supported |

Derivatives are detected via `ID_LIKE` in `/etc/os-release` (Linux Mint, Pop!_OS, Manjaro, EndeavourOS, etc.).

Unsupported platforms are rejected at startup with a clear error message.

## Supported agents

| Agent | Tier | Status |
|-------|------|--------|
| Claude Code | Full | Supported |
| OpenCode | Full | Supported |

## Components

| Component | Description |
|-----------|-------------|
| `engram` | Persistent memory system |
| `sdd` | Spec-Driven Development skills (9 phases) |
| `skills` | Coding skills library (React, TS, Tailwind, etc.) |
| `context7` | MCP server for live library documentation |
| `persona` | Gentleman mentor persona |
| `permissions` | Security-first defaults (blocks .env access) |
| `gga` | Guardian Angel code review (optional) |

## Quick start

### Prerequisites

- Go 1.22+
- `git` in PATH
- Platform-specific:
  - **macOS**: Homebrew
  - **Ubuntu/Debian**: `apt-get`, `sudo`
  - **Arch**: `pacman`, `sudo`

### Dry run (see what would happen)

```bash
go run ./cmd/gentleman-ai install --dry-run \
  --agent claude-code,opencode \
  --component engram,sdd,skills,context7,persona,permissions
```

### Install

```bash
go run ./cmd/gentleman-ai install \
  --agent claude-code,opencode \
  --preset full-gentleman
```

## CLI flags

| Flag | Description |
|------|-------------|
| `--agent`, `--agents` | Agents to configure (comma-separated) |
| `--component`, `--components` | Components to install |
| `--skill`, `--skills` | Skills to install |
| `--persona` | Persona: `gentleman`, `neutral`, `custom` |
| `--preset` | Preset: `full-gentleman`, `ecosystem-only`, `minimal`, `custom` |
| `--dry-run` | Preview plan without applying changes |

## Architecture

```
cmd/gentleman-ai/          CLI entrypoint
internal/
  app/                     Command dispatch + runtime wiring
  system/                  OS/distro detection, platform profiles, support guards
  cli/                     Install flags, validation, orchestration, dry-run
  planner/                 Dependency graph, resolution, ordering, review payloads
  installcmd/              Profile-aware command resolver (brew/apt/pacman)
  pipeline/                Staged execution + rollback orchestration
  backup/                  Config snapshot + restore
  components/              Per-component install/inject logic
    engram/  sdd/  skills/  mcp/  persona/  theme/  permissions/  gga/
  agents/                  Agent adapters
    claude/  opencode/
  verify/                  Post-apply health checks + reporting
  tui/                     Bubbletea TUI (Rose Pine theme)
    styles/  screens/
e2e/                       Docker-based E2E tests (Ubuntu + Arch)
testdata/                  Golden test fixtures
docs/                      Additional documentation
```

## Testing

```bash
# Run all tests
go test ./...

# Dry-run smoke test
go run ./cmd/gentleman-ai install --dry-run

# Docker E2E (requires Docker)
cd e2e && ./docker-test.sh
```

## Documentation

- [Quick Start](docs/quickstart.md)
- [Non-Interactive Mode](docs/non-interactive.md)
- [Backup & Rollback](docs/rollback.md)
- [Docker E2E Testing](docs/docker-e2e-testing.md)

## Relationship to Gentleman.Dots

| | Gentleman.Dots | AI Gentle Stack |
|--|---------------|-----------------|
| **Purpose** | Dev environment (editors, shells, terminals) | AI development layer (agents, memory, skills) |
| **Installs** | Neovim, Fish/Zsh, Tmux/Zellij, Ghostty | Claude Code, OpenCode, Engram, SDD, MCP, Skills |
| **Overlap** | None -- complementary | None -- different layer |

Install Gentleman.Dots first for your dev environment, then AI Gentle Stack for the AI layer on top.

## License

MIT
