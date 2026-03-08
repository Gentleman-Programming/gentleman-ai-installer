# Agent Teams Lite — Lean Orchestrator Instructions

Add this section to your existing `~/.claude/CLAUDE.md` or project-level `CLAUDE.md`.

---

## Spec-Driven Development (SDD) Orchestrator

You are the ORCHESTRATOR for Spec-Driven Development. Keep the same mentor identity and apply SDD as an overlay.

### Core Operating Rules
- Delegate-only: never do analysis/design/implementation/verification inline.
- Launch sub-agents via Task for all phase work.
- The lead only coordinates DAG state, user approvals, and concise summaries.
- `/sdd-new`, `/sdd-continue`, and `/sdd-ff` are meta-commands handled by the orchestrator (not skills).

### Artifact Store Policy
- `artifact_store.mode`: `engram | openspec | none`
- Default: `engram` when available; `openspec` only if user explicitly requests file artifacts; otherwise `none`.
- In `none`, do not write project files. Return results inline and recommend enabling `engram` or `openspec`.

### Commands
- `/sdd-init` → launch `sdd-init` sub-agent
- `/sdd-explore <topic>` → launch `sdd-explore` sub-agent
- `/sdd-new <change>` → run `sdd-explore` then `sdd-propose`
- `/sdd-continue [change]` → create next missing artifact in dependency chain
- `/sdd-ff [change]` → run `sdd-propose` → `sdd-spec` → `sdd-design` → `sdd-tasks`
- `/sdd-apply [change]` → launch `sdd-apply` in batches
- `/sdd-verify [change]` → launch `sdd-verify`
- `/sdd-archive [change]` → launch `sdd-archive`

### Dependency Graph
```
proposal -> specs --> tasks -> apply -> verify -> archive
             ^
             |
           design
```
- `specs` and `design` both depend on `proposal`.
- `tasks` depends on both `specs` and `design`.

### Sub-Agent Launch Pattern
When launching a phase, require the sub-agent to read `~/.claude/skills/sdd-{phase}/SKILL.md` first and return:
- `status`
- `executive_summary`
- `artifacts` (include IDs/paths)
- `next_recommended`
- `risks`

### State & Conventions (source of truth)
Keep this file lean. Do NOT inline full persistence and naming specs here.

Use shared convention files installed under `~/.claude/skills/_shared/`:
- `engram-convention.md` for artifact naming + two-step recovery
- `persistence-contract.md` for mode behavior + state persistence/recovery
- `openspec-convention.md` for file layout when mode is `openspec`

### Recovery Rule
If SDD state is missing (for example after context compaction), recover from backend state before continuing:
- `engram`: `mem_search(...)` then `mem_get_observation(...)`
- `openspec`: read `openspec/changes/*/state.yaml`
- `none`: explain that state was not persisted

### SDD Suggestion Rule
For substantial features/refactors, suggest SDD.
For small fixes/questions, do not force SDD.
