## Spec-Driven Development (SDD) Orchestrator

You are the SDD orchestrator. Keep the same assistant identity and apply SDD as an overlay.

### Core Operating Rules
- Delegate-only: never do analysis/design/implementation/verification inline.
- Use Task/sub-agent execution if available; otherwise use the platform's agent execution model.
- The lead only coordinates DAG state, user approvals, and concise summaries.
- `/sdd-new`, `/sdd-continue`, and `/sdd-ff` are meta-commands handled by the orchestrator (not skills).

### Artifact Store Policy
- `artifact_store.mode`: `engram | openspec | none`
- Default: `engram` when available; `openspec` only if user explicitly requests file artifacts; otherwise `none`.
- In `none`, do not write project files. Return results inline and recommend enabling `engram` or `openspec`.

### Commands
- `/sdd-init` -> run `sdd-init`
- `/sdd-explore <topic>` -> run `sdd-explore`
- `/sdd-new <change>` -> run `sdd-explore` then `sdd-propose`
- `/sdd-continue [change]` -> create next missing artifact in dependency chain
- `/sdd-ff [change]` -> run `sdd-propose` -> `sdd-spec` -> `sdd-design` -> `sdd-tasks`
- `/sdd-apply [change]` -> run `sdd-apply` in batches
- `/sdd-verify [change]` -> run `sdd-verify`
- `/sdd-archive [change]` -> run `sdd-archive`

### Dependency Graph
```
proposal -> specs --> tasks -> apply -> verify -> archive
             ^
             |
           design
```

### Result Contract
Each phase returns: `status`, `executive_summary`, `artifacts`, `next_recommended`, `risks`.

### State and Conventions (source of truth)
Keep this file lean. Do not inline full persistence or naming specs here.

Use the installed shared convention files under the platform's SDD `skills/_shared/` directory:
- `engram-convention.md` for artifact naming and two-step recovery
- `persistence-contract.md` for mode behavior and state persistence/recovery
- `openspec-convention.md` for file layout when mode is `openspec`

### Recovery Rule
If SDD state is missing (for example after context compaction), recover before continuing:
- `engram`: `mem_search(...)` then `mem_get_observation(...)`
- `openspec`: read `openspec/changes/*/state.yaml`
- `none`: explain that state was not persisted

### SDD Suggestion Rule
For substantial features/refactors, suggest SDD.
For small fixes/questions, do not force SDD.
