# Persistence Contract (shared across all SDD skills)

## Mode Resolution

The orchestrator passes `artifact_store.mode` with one of: `engram | openspec | none`.

Default resolution (when orchestrator does not explicitly set a mode):
1. If Engram is available → use `engram`
2. Otherwise → use `none`

`openspec` is NEVER used by default — only when the orchestrator explicitly passes `openspec`.

When falling back to `none`, recommend the user enable `engram` or `openspec` for better results.

## Behavior Per Mode

| Mode | Read from | Write to | Project files |
|------|-----------|----------|---------------|
| `engram` | Engram (see `engram-convention.md`) | Engram | Never |
| `openspec` | Filesystem (see `openspec-convention.md`) | Filesystem | Yes |
| `none` | Orchestrator prompt context | Nowhere | Never |

## State Persistence (Orchestrator)

The orchestrator persists DAG state after each phase transition. This enables SDD recovery after context compaction.

| Mode | Persist State | Recover State |
|------|--------------|---------------|
| `engram` | `mem_save(topic_key: "sdd/{change-name}/state")` | `mem_search("sdd/*/state")` → `mem_get_observation(id)` |
| `openspec` | Write `openspec/changes/{change-name}/state.yaml` | Read `openspec/changes/{change-name}/state.yaml` |
| `none` | Not possible — state lives only in context | Not possible — warn user |

## Common Rules

- If mode is `none`, do NOT create or modify any project files. Return results inline only.
- If mode is `engram`, do NOT write any project files. Persist to Engram and return observation IDs.
- If mode is `openspec`, write files ONLY to the paths defined in `openspec-convention.md`.
- NEVER force `openspec/` creation unless the orchestrator explicitly passed `openspec` mode.
- If you are unsure which mode to use, default to `none`.

## Detail Level

The orchestrator may also pass `detail_level`: `concise | standard | deep`.
This controls output verbosity but does NOT affect what gets persisted — always persist the full artifact.
