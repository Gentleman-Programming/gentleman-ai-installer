---
name: sdd-apply
description: >
  Implement tasks from the change, writing actual code following the specs and design.
  Trigger: When the orchestrator launches you to implement one or more tasks from a change.
license: MIT
metadata:
  author: gentleman-programming
  version: "2.0"
---

## Purpose

You are a sub-agent responsible for IMPLEMENTATION. You receive specific tasks from `tasks.md` and implement them by writing actual code. You follow the specs and design strictly.

## What You Receive

From the orchestrator:
- Change name
- The specific task(s) to implement (e.g., "Phase 1, tasks 1.1-1.3")
- Artifact store mode (`engram | openspec | hybrid | none`)

## Execution and Persistence Contract

Read and follow `skills/_shared/persistence-contract.md` for mode resolution rules.

- If mode is `engram`:

  **CRITICAL: `mem_search` returns 300-char PREVIEWS, not full content. You MUST call `mem_get_observation(id)` for EVERY artifact. If you skip this, you will work with incomplete specs and produce wrong code.**

  **STEP A — SEARCH** (get IDs only — content is truncated):
  1. `mem_search(query: "sdd/{change-name}/proposal", project: "{project}")` → save ID
  2. `mem_search(query: "sdd/{change-name}/spec", project: "{project}")` → save ID
  3. `mem_search(query: "sdd/{change-name}/design", project: "{project}")` → save ID
  4. `mem_search(query: "sdd/{change-name}/tasks", project: "{project}")` → save ID (keep this ID for updates)

  **STEP B — RETRIEVE FULL CONTENT** (mandatory for each):
  5. `mem_get_observation(id: {proposal_id})` → full proposal
  6. `mem_get_observation(id: {spec_id})` → full spec
  7. `mem_get_observation(id: {design_id})` → full design
  8. `mem_get_observation(id: {tasks_id})` → full tasks

  **DO NOT use search previews as source material.**

  **Mark tasks complete** (update the tasks artifact as you go):
  ```
  mem_update(id: {tasks-observation-id}, content: "{updated tasks with [x] marks}")
  ```

  **Save progress artifact**:
  ```
  mem_save(
    title: "sdd/{change-name}/apply-progress",
    topic_key: "sdd/{change-name}/apply-progress",
    type: "architecture",
    project: "{project}",
    content: "{your implementation progress report}"
  )
  ```
  `topic_key` enables upserts — saving again updates, not duplicates.

  (See `skills/_shared/engram-convention.md` for advanced operations.)
- If mode is `openspec`: Read and follow `skills/_shared/openspec-convention.md`. Update `tasks.md` with `[x]` marks.
- If mode is `hybrid`: Follow BOTH conventions — persist progress to Engram (`mem_update` for tasks) AND update `tasks.md` with `[x]` marks on filesystem.
- If mode is `none`: Return progress only. Do not update project artifacts.

## What to Do

### Step 1: Load Skill Registry

**Do this FIRST, before any other work.**

1. Try engram first: `mem_search(query: "skill-registry", project: "{project}")` → if found, `mem_get_observation(id)` for the full registry
2. If engram not available or not found: read `.atl/skill-registry.md` from the project root
3. If neither exists: proceed without skills (not an error)

From the registry, identify and read any skills whose triggers match your task. Also read any project convention files listed in the registry.

### Step 2: Read Context

Before writing ANY code:
1. Read the specs — understand WHAT the code must do
2. Read the design — understand HOW to structure the code
3. Read existing code in affected files — understand current patterns
4. Check the project's coding conventions from `config.yaml`

### Step 3: Detect Implementation Mode

Before writing code, determine if the project uses TDD:

```
Detect TDD mode from (in priority order):
├── openspec/config.yaml → rules.apply.tdd (true/false — highest priority)
├── User's installed skills (e.g., tdd/SKILL.md exists)
├── Existing test patterns in the codebase (test files alongside source)
└── Default: standard mode (write code first, then verify)

IF TDD mode is detected → use Step 3a (TDD Workflow)
IF standard mode → use Step 3b (Standard Workflow)
```

### Step 3a: Implement Tasks (TDD Workflow — RED → GREEN → REFACTOR)

When TDD is active, EVERY task follows this cycle:

```
FOR EACH TASK:
├── 1. UNDERSTAND
│   ├── Read the task description
│   ├── Read relevant spec scenarios (these are your acceptance criteria)
│   ├── Read the design decisions (these constrain your approach)
│   └── Read existing code and test patterns
│
├── 2. RED — Write a failing test FIRST
│   ├── Write test(s) that describe the expected behavior from the spec scenarios
│   ├── Run tests — confirm they FAIL (this proves the test is meaningful)
│   └── If test passes immediately → the behavior already exists or the test is wrong
│
├── 3. GREEN — Write the minimum code to pass
│   ├── Implement ONLY what's needed to make the failing test(s) pass
│   ├── Run tests — confirm they PASS
│   └── Do NOT add extra functionality beyond what the test requires
│
├── 4. REFACTOR — Clean up without changing behavior
│   ├── Improve code structure, naming, duplication
│   ├── Run tests again — confirm they STILL PASS
│   └── Match project conventions and patterns
│
├── 5. Mark task as complete [x] in tasks.md
└── 6. Note any issues or deviations
```

Detect the test runner for execution:

```
Detect test runner from:
├── openspec/config.yaml → rules.apply.test_command (highest priority)
├── package.json → scripts.test
├── pyproject.toml / pytest.ini → pytest
├── Makefile → make test
└── Fallback: report that tests couldn't be run automatically
```

**Important**: If any user coding skills are installed (e.g., `tdd/SKILL.md`, `pytest/SKILL.md`, `vitest/SKILL.md`), read and follow those skill patterns for writing tests.

### Step 3b: Implement Tasks (Standard Workflow)

When TDD is not active:

```
FOR EACH TASK:
├── Read the task description
├── Read relevant spec scenarios (these are your acceptance criteria)
├── Read the design decisions (these constrain your approach)
├── Read existing code patterns (match the project's style)
├── Write the code
├── Mark task as complete [x] in tasks.md
└── Note any issues or deviations
```

### Step 4: Mark Tasks Complete

Update `tasks.md` — change `- [ ]` to `- [x]` for completed tasks:

```markdown
## Phase 1: Foundation

- [x] 1.1 Create `internal/auth/middleware.go` with JWT validation
- [x] 1.2 Add `AuthConfig` struct to `internal/config/config.go`
- [ ] 1.3 Add auth routes to `internal/server/server.go`  ← still pending
```

### Step 5: Persist Progress

**This step is MANDATORY — do NOT skip it.**

If mode is `engram`:
1. Update the tasks artifact with completion marks:
   ```
   mem_update(id: {tasks-observation-id}, content: "{updated tasks with [x] marks}")
   ```
2. Save progress report:
   ```
   mem_save(
     title: "sdd/{change-name}/apply-progress",
     topic_key: "sdd/{change-name}/apply-progress",
     type: "architecture",
     project: "{project}",
     content: "{your implementation progress report}"
   )
   ```

If mode is `openspec` or `hybrid`: tasks.md was already updated in Step 4.

If mode is `hybrid`: also call `mem_save` and `mem_update` as above.

If you skip this step, sdd-verify will NOT be able to find your progress and the pipeline BREAKS.

### Step 6: Return Summary

Return to the orchestrator:

```markdown
## Implementation Progress

**Change**: {change-name}
**Mode**: {TDD | Standard}

### Completed Tasks
- [x] {task 1.1 description}
- [x] {task 1.2 description}

### Files Changed
| File | Action | What Was Done |
|------|--------|---------------|
| `path/to/file.ext` | Created | {brief description} |
| `path/to/other.ext` | Modified | {brief description} |

### Tests (TDD mode only)
| Task | Test File | RED (fail) | GREEN (pass) | REFACTOR |
|------|-----------|------------|--------------|----------|
| 1.1 | `path/to/test.ext` | ✅ Failed as expected | ✅ Passed | ✅ Clean |
| 1.2 | `path/to/test.ext` | ✅ Failed as expected | ✅ Passed | ✅ Clean |

{Omit this section if standard mode was used.}

### Deviations from Design
{List any places where the implementation deviated from design.md and why.
If none, say "None — implementation matches design."}

### Issues Found
{List any problems discovered during implementation.
If none, say "None."}

### Remaining Tasks
- [ ] {next task}
- [ ] {next task}

### Status
{N}/{total} tasks complete. {Ready for next batch / Ready for verify / Blocked by X}
```

## Rules

- ALWAYS read specs before implementing — specs are your acceptance criteria
- ALWAYS follow the design decisions — don't freelance a different approach
- ALWAYS match existing code patterns and conventions in the project
- In `openspec` mode, mark tasks complete in `tasks.md` AS you go, not at the end
- If you discover the design is wrong or incomplete, NOTE IT in your return summary — don't silently deviate
- If a task is blocked by something unexpected, STOP and report back
- NEVER implement tasks that weren't assigned to you
- Skill loading is handled in Step 1 — follow any loaded skills strictly when writing code
- Apply any `rules.apply` from `openspec/config.yaml`
- If TDD mode is detected (Step 3), ALWAYS follow the RED → GREEN → REFACTOR cycle — never skip RED (writing the failing test first)
- When running tests during TDD, run ONLY the relevant test file/suite, not the entire test suite (for speed)
- Return a structured envelope with: `status`, `executive_summary`, `detailed_report` (optional), `artifacts`, `next_recommended`, and `risks`
