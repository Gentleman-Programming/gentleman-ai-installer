---
name: sdd-propose
description: >
  Create a change proposal with intent, scope, and approach.
  Trigger: When the orchestrator launches you to create or update a proposal for a change.
license: MIT
metadata:
  author: gentleman-programming
  version: "2.0"
---

## Purpose

You are a sub-agent responsible for creating PROPOSALS. You take the exploration analysis (or direct user input) and produce a structured `proposal.md` document inside the change folder.

## What You Receive

From the orchestrator:
- Change name (e.g., "add-dark-mode")
- Exploration analysis (from sdd-explore) OR direct user description
- Artifact store mode (`engram | openspec | hybrid | none`)

## Execution and Persistence Contract

Read and follow `skills/_shared/persistence-contract.md` for mode resolution rules.

- If mode is `engram`:

  **Read dependencies** (two-step — search returns truncated previews):
  1. `mem_search(query: "sdd/{change-name}/explore", project: "{project}")` → get observation ID (optional — may not exist)
  2. If found: `mem_get_observation(id: {id})` → full exploration content
  3. `mem_search(query: "sdd-init/{project}", project: "{project}")` → project context (optional)
  4. If found: `mem_get_observation(id: {id})` → full project context

  **Save your artifact**:
  ```
  mem_save(
    title: "sdd/{change-name}/proposal",
    topic_key: "sdd/{change-name}/proposal",
    type: "architecture",
    project: "{project}",
    content: "{your full proposal markdown}"
  )
  ```
  `topic_key` enables upserts — saving again updates, not duplicates.

  (See `skills/_shared/engram-convention.md` for full naming conventions.)
- If mode is `openspec`: Read and follow `skills/_shared/openspec-convention.md`.
- If mode is `hybrid`: Follow BOTH conventions — persist to Engram AND write to filesystem. Retrieve dependencies from Engram (primary) with filesystem fallback.
- If mode is `none`: Return result only. Never create or modify project files.
- Never force `openspec/` creation unless user requested file-based persistence or mode is `hybrid`.

## What to Do

### Step 1: Load Skill Registry

**Do this FIRST, before any other work.**

1. Try engram first: `mem_search(query: "skill-registry", project: "{project}")` → if found, `mem_get_observation(id)` for the full registry
2. If engram not available or not found: read `.atl/skill-registry.md` from the project root
3. If neither exists: proceed without skills (not an error)

From the registry, identify and read any skills whose triggers match your task. Also read any project convention files listed in the registry.

### Step 2: Create Change Directory

**IF mode is `openspec` or `hybrid`:** create the change folder structure:

```
openspec/changes/{change-name}/
└── proposal.md
```

**IF mode is `engram` or `none`:** Do NOT create any `openspec/` directories. Skip this step.

### Step 3: Read Existing Specs

**IF mode is `openspec` or `hybrid`:** If `openspec/specs/` has relevant specs, read them to understand current behavior that this change might affect.

**IF mode is `engram`:** Existing context was already retrieved from Engram in the Persistence Contract. Skip filesystem reads.

**IF mode is `none`:** Skip — no existing specs to read.

### Step 4: Write proposal.md

```markdown
# Proposal: {Change Title}

## Intent

{What problem are we solving? Why does this change need to happen?
Be specific about the user need or technical debt being addressed.}

## Scope

### In Scope
- {Concrete deliverable 1}
- {Concrete deliverable 2}
- {Concrete deliverable 3}

### Out of Scope
- {What we're explicitly NOT doing}
- {Future work that's related but deferred}

## Approach

{High-level technical approach. How will we solve this?
Reference the recommended approach from exploration if available.}

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `path/to/area` | New/Modified/Removed | {What changes} |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| {Risk description} | Low/Med/High | {How we mitigate} |

## Rollback Plan

{How to revert if something goes wrong. Be specific.}

## Dependencies

- {External dependency or prerequisite, if any}

## Success Criteria

- [ ] {How do we know this change succeeded?}
- [ ] {Measurable outcome}
```

### Step 5: Persist Artifact

**This step is MANDATORY — do NOT skip it.**

If mode is `engram`:
```
mem_save(
  title: "sdd/{change-name}/proposal",
  topic_key: "sdd/{change-name}/proposal",
  type: "architecture",
  project: "{project}",
  content: "{your full proposal markdown from Step 4}"
)
```

If mode is `openspec` or `hybrid`: the file was already written in Step 4.

If mode is `hybrid`: also call `mem_save` as above (write to BOTH backends).

If you skip this step, the next phase (sdd-spec) will NOT be able to find your proposal and the pipeline BREAKS.

### Step 6: Return Summary

Return to the orchestrator:

```markdown
## Proposal Created

**Change**: {change-name}
**Location**: `openspec/changes/{change-name}/proposal.md` (openspec/hybrid) | Engram `sdd/{change-name}/proposal` (engram) | inline (none)

### Summary
- **Intent**: {one-line summary}
- **Scope**: {N deliverables in, M items deferred}
- **Approach**: {one-line approach}
- **Risk Level**: {Low/Medium/High}

### Next Step
Ready for specs (sdd-spec) or design (sdd-design).
```

## Rules

- In `openspec` mode, ALWAYS create the `proposal.md` file
- If the change directory already exists with a proposal, READ it first and UPDATE it
- Keep the proposal CONCISE - it's a thinking tool, not a novel
- Every proposal MUST have a rollback plan
- Every proposal MUST have success criteria
- Use concrete file paths in "Affected Areas" when possible
- Apply any `rules.proposal` from `openspec/config.yaml`
- Return a structured envelope with: `status`, `executive_summary`, `detailed_report` (optional), `artifacts`, `next_recommended`, and `risks`
