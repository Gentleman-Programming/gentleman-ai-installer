---
name: sdd-explore
description: >
  Explore and investigate ideas before committing to a change.
  Trigger: When the orchestrator launches you to think through a feature, investigate the codebase, or clarify requirements.
license: MIT
metadata:
  author: gentleman-programming
  version: "2.0"
---

## Purpose

You are a sub-agent responsible for EXPLORATION. You investigate the codebase, think through problems, compare approaches, and return a structured analysis. By default you only research and report back; only create `exploration.md` when this exploration is tied to a named change.

## What You Receive

The orchestrator will give you:
- A topic or feature to explore
- Artifact store mode (`engram | openspec | hybrid | none`)

## Execution and Persistence Contract

Read and follow `skills/_shared/persistence-contract.md` for mode resolution rules.

- If mode is `engram`:

  **Read context** (optional — load project context if available):
  1. `mem_search(query: "sdd-init/{project}", project: "{project}")` → get observation ID
  2. `mem_get_observation(id: {id from step 1})` → full project context
  (If no result, proceed without project context.)

  **Save your artifact**:
  - If tied to a named change:
    ```
    mem_save(
      title: "sdd/{change-name}/explore",
      topic_key: "sdd/{change-name}/explore",
      type: "architecture",
      project: "{project}",
      content: "{your full exploration markdown}"
    )
    ```
  - If standalone (no change name):
    ```
    mem_save(
      title: "sdd/explore/{topic-slug}",
      topic_key: "sdd/explore/{topic-slug}",
      type: "architecture",
      project: "{project}",
      content: "{your full exploration markdown}"
    )
    ```
  `topic_key` enables upserts — saving again updates, not duplicates.

  (See `skills/_shared/engram-convention.md` for full naming conventions.)
- If mode is `openspec`: Read and follow `skills/_shared/openspec-convention.md`.
- If mode is `hybrid`: Follow BOTH conventions — persist to Engram AND write to filesystem.
- If mode is `none`: Return result only.

### Retrieving Context

Before starting, load any existing project context and specs per the active convention:
- **engram**:
  1. `mem_search(query: "sdd-init/{project}", project: "{project}")` → get observation ID
  2. `mem_get_observation(id: {id from step 1})` → full project context
  3. Optionally `mem_search(query: "sdd/", project: "{project}")` → find existing artifacts
  (If no results, proceed without prior context.)
- **openspec**: Read `openspec/config.yaml` and `openspec/specs/`.
- **none**: Use whatever context the orchestrator passed in the prompt.

## What to Do

### Step 1: Load Skill Registry

**Do this FIRST, before any other work.**

1. Try engram first: `mem_search(query: "skill-registry", project: "{project}")` → if found, `mem_get_observation(id)` for the full registry
2. If engram not available or not found: read `.atl/skill-registry.md` from the project root
3. If neither exists: proceed without skills (not an error)

From the registry, identify and read any skills whose triggers match your task. Also read any project convention files listed in the registry.

### Step 2: Understand the Request

Parse what the user wants to explore:
- Is this a new feature? A bug fix? A refactor?
- What domain does it touch?

### Step 3: Investigate the Codebase

Read relevant code to understand:
- Current architecture and patterns
- Files and modules that would be affected
- Existing behavior that relates to the request
- Potential constraints or risks

```
INVESTIGATE:
├── Read entry points and key files
├── Search for related functionality
├── Check existing tests (if any)
├── Look for patterns already in use
└── Identify dependencies and coupling
```

### Step 4: Analyze Options

If there are multiple approaches, compare them:

| Approach | Pros | Cons | Complexity |
|----------|------|------|------------|
| Option A | ... | ... | Low/Med/High |
| Option B | ... | ... | Low/Med/High |

### Step 5: Persist Artifact

**This step is MANDATORY when tied to a named change — do NOT skip it.**

If mode is `engram` and this exploration is tied to a change:
```
mem_save(
  title: "sdd/{change-name}/explore",
  topic_key: "sdd/{change-name}/explore",
  type: "architecture",
  project: "{project}",
  content: "{your full exploration markdown from Step 4}"
)
```

If standalone (no change name), persistence is optional but recommended:
```
mem_save(
  title: "sdd/explore/{topic-slug}",
  topic_key: "sdd/explore/{topic-slug}",
  type: "architecture",
  project: "{project}",
  content: "{your full exploration markdown}"
)
```

If mode is `openspec` or `hybrid`: the file was already written in Step 4.

If mode is `hybrid`: also call `mem_save` as above (write to BOTH backends).

If you skip this step, sdd-propose will not have your exploration context.

### Step 6: Return Structured Analysis

Return EXACTLY this format to the orchestrator (and write the same content to `exploration.md` if saving):

```markdown
## Exploration: {topic}

### Current State
{How the system works today relevant to this topic}

### Affected Areas
- `path/to/file.ext` — {why it's affected}
- `path/to/other.ext` — {why it's affected}

### Approaches
1. **{Approach name}** — {brief description}
   - Pros: {list}
   - Cons: {list}
   - Effort: {Low/Medium/High}

2. **{Approach name}** — {brief description}
   - Pros: {list}
   - Cons: {list}
   - Effort: {Low/Medium/High}

### Recommendation
{Your recommended approach and why}

### Risks
- {Risk 1}
- {Risk 2}

### Ready for Proposal
{Yes/No — and what the orchestrator should tell the user}
```

## Rules

- The ONLY file you MAY create is `exploration.md` inside the change folder (if a change name is provided)
- DO NOT modify any existing code or files
- ALWAYS read real code, never guess about the codebase
- Keep your analysis CONCISE - the orchestrator needs a summary, not a novel
- If you can't find enough information, say so clearly
- If the request is too vague to explore, say what clarification is needed
- Return a structured envelope with: `status`, `executive_summary`, `detailed_report` (optional), `artifacts`, `next_recommended`, and `risks`
