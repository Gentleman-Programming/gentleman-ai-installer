# Contributing to Gentle AI <!-- v1.0 -->

Thank you for your interest in contributing to **Gentle AI** (`gga`) — a Go TUI installer for AI agent environments.

Before you dive in, please read this guide fully. We have a structured workflow to keep the project organized and maintainable.

---

## Table of Contents

- [Issue-First Workflow](#issue-first-workflow)
- [Label System](#label-system)
- [Development Setup](#development-setup)
- [Testing](#testing)
- [Commit Convention](#commit-convention)
- [Pull Request Rules](#pull-request-rules)
- [Code of Conduct](#code-of-conduct)

---

## Issue-First Workflow

**No PR without an issue. No exceptions.**

This project follows a strict issue-first workflow:

1. **Open an issue** using the appropriate template ([Bug Report](https://github.com/Gentleman-Programming/gentle-ai/issues/new?template=bug_report.yml) or [Feature Request](https://github.com/Gentleman-Programming/gentle-ai/issues/new?template=feature_request.yml))
2. **Wait for approval** — a maintainer will add the `status:approved` label when the issue is ready to be worked on
3. **Comment on the issue** to let others know you're working on it
4. **Open a PR** referencing the approved issue

PRs that are not linked to an approved issue will be **automatically rejected** by CI.

---

## Label System

### Type Labels (applied to PRs)

| Label | Description |
|-------|-------------|
| `type:bug` | Bug fix |
| `type:feature` | New feature or enhancement |
| `type:refactor` | Code refactoring, no functional changes |
| `type:docs` | Documentation only |
| `type:test` | Test coverage additions |
| `type:chore` | Build, CI, tooling changes |
| `type:breaking` | Breaking change |

### Status Labels (applied to Issues)

| Label | Description |
|-------|-------------|
| `status:needs-review` | Newly opened, awaiting maintainer review |
| `status:approved` | Approved for implementation — work can begin |
| `status:in-progress` | Being worked on |
| `status:blocked` | Blocked by another issue or external dependency |
| `status:wont-fix` | Out of scope or won't be addressed |

### Priority Labels

| Label | Description |
|-------|-------------|
| `priority:critical` | Blocking issues, security vulnerabilities |
| `priority:high` | Important, affects many users |
| `priority:medium` | Normal priority |
| `priority:low` | Nice to have |

---

## Development Setup

### Prerequisites

- Go 1.22+
- Docker (for E2E tests)
- Git

### Clone and Build

```bash
git clone https://github.com/Gentleman-Programming/gentle-ai.git
cd gentle-ai
go build -o gga .
```

### Run Locally

```bash
./gga
```

---

## Testing

### Unit Tests

Run the full unit test suite:

```bash
go test ./...
```

Run tests for a specific package:

```bash
go test ./internal/tui/...
```

Run with verbose output:

```bash
go test -v ./...
```

### E2E Tests

E2E tests are Docker-based shell scripts. Docker must be running.

```bash
cd e2e
chmod +x docker-test.sh
./docker-test.sh
```

> ⚠️ E2E tests spin up containers to simulate real installation environments. They may take a few minutes to complete.

---

## Commit Convention

This project uses [Conventional Commits](https://www.conventionalcommits.org/).

### Format

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

### Types

| Type | When to Use |
|------|-------------|
| `feat` | New feature |
| `fix` | Bug fix |
| `refactor` | Code change that's neither fix nor feature |
| `docs` | Documentation changes |
| `test` | Adding or updating tests |
| `chore` | Build, CI, tooling |
| `perf` | Performance improvement |

### Scopes (optional but recommended)

`tui`, `cli`, `installer`, `catalog`, `system`, `agent`, `e2e`, `ci`, `docs`

### Examples

```
feat(tui): add progress bar to installation steps
fix(agent): correct Claude Code detection on macOS
docs: update contributing guide with E2E test instructions
chore(ci): add unit tests job to CI workflow
test(installer): add coverage for catalog step execution
```

### Breaking Changes

For breaking changes, add `!` after the type and include a `BREAKING CHANGE:` footer:

```
feat(cli)!: rename --config flag to --config-file

BREAKING CHANGE: the --config flag has been renamed to --config-file.
Update your scripts and aliases accordingly.
```

---

## Pull Request Rules

### Before Opening a PR

- [ ] There is a linked approved issue (`Closes #<N>`)
- [ ] All unit tests pass (`go test ./...`)
- [ ] E2E tests pass (`cd e2e && ./docker-test.sh`)
- [ ] Commits follow Conventional Commits format
- [ ] Code is self-reviewed

### PR Title

Use the same Conventional Commits format as commit messages:

```
feat(tui): add keyboard shortcut help overlay
fix(agent): handle missing HOME env var gracefully
```

### Automated PR Checks

All PRs go through automated checks:

| Check | What It Verifies |
|-------|-----------------|
| **Check Issue Reference** | PR body contains `Closes/Fixes/Resolves #N` |
| **Check Issue Has status:approved** | The linked issue has been approved by a maintainer |
| **Check PR Has type:* Label** | Exactly one `type:*` label is applied |
| **Unit Tests** | `go test ./...` passes |
| **E2E Tests** | `cd e2e && ./docker-test.sh` passes |

**All checks must pass** before a PR can be merged.

### Linking Your Issue

In the PR body, include one of:

```
Closes #42
Fixes #42
Resolves #42
```

---

## Code of Conduct

Be respectful. We're building something together.

- Critique code, not people
- Be constructive in reviews
- Welcome newcomers

Violations may result in removal from the project.

---

## Questions?

Use [GitHub Discussions](https://github.com/Gentleman-Programming/gentle-ai/discussions) — not issues — for questions, ideas, and general conversation.
