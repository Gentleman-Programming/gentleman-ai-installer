# Persistence Contract (shared across all SDD skills)

## Mode Resolution

The orchestrator passes `artifact_store.mode` with one of: `engram | openspec | hybrid | none`.

Default resolution (when orchestrator does not explicitly set a mode):
1. If Engram is available → use `engram`
2. Otherwise → use `none`

`openspec` and `hybrid` are NEVER used by default — only when the orchestrator explicitly passes them.

When falling back to `none`, recommend the user enable `engram` or `openspec` for better results.

## Behavior Per Mode

| Mode | Read from | Write to | Project files |
|------|-----------|----------|---------------|
| `engram` | Engram (see `engram-convention.md`) | Engram | Never |
| `openspec` | Filesystem (see `openspec-convention.md`) | Filesystem | Yes |
| `hybrid` | Engram (primary) + Filesystem (fallback) | Both Engram AND Filesystem | Yes |
| `none` | Orchestrator prompt context | Nowhere | Never |

### Hybrid Mode

`hybrid` persists every artifact to BOTH Engram and OpenSpec simultaneously. This provides:
- **Engram**: cross-session recovery, compaction survival, deterministic search
- **OpenSpec**: human-readable files in the project, version-controllable artifacts

**Read priority**: Engram first (faster, survives compaction). Fall back to filesystem if Engram returns no results.

**Write behavior**: Write to Engram (per `engram-convention.md`) AND to filesystem (per `openspec-convention.md`) for every artifact. Both writes MUST succeed for the operation to be considered complete.

**Token cost warning**: Hybrid mode consumes MORE tokens per operation than either single backend, because every read/write hits both stores. Use it when you need both cross-session persistence AND local file artifacts. If you only need one benefit, prefer `engram` or `openspec` alone.

## State Persistence (Orchestrator)

The orchestrator persists DAG state after each phase transition. This enables SDD recovery after context compaction.

| Mode | Persist State | Recover State |
|------|--------------|---------------|
| `engram` | `mem_save(topic_key: "sdd/{change-name}/state")` | `mem_search("sdd/*/state")` → `mem_get_observation(id)` |
| `openspec` | Write `openspec/changes/{change-name}/state.yaml` | Read `openspec/changes/{change-name}/state.yaml` |
| `hybrid` | Both: `mem_save` AND write `state.yaml` | Engram first; filesystem fallback |
| `none` | Not possible — state lives only in context | Not possible — warn user |

## Common Rules

- If mode is `none`, do NOT create or modify any project files. Return results inline only.
- If mode is `engram`, do NOT write any project files. Persist to Engram and return observation IDs.
- If mode is `openspec`, write files ONLY to the paths defined in `openspec-convention.md`.
- If mode is `hybrid`, persist to BOTH Engram AND filesystem. Follow both `engram-convention.md` and `openspec-convention.md` for each artifact.
- NEVER force `openspec/` creation unless the orchestrator explicitly passed `openspec` or `hybrid` mode.
- If you are unsure which mode to use, default to `none`.

## Sub-Agent Context Rules

Sub-agents launch with a fresh context and NO access to the orchestrator's instructions or memory protocol. The orchestrator controls what context they receive and sub-agents are responsible for persisting what they produce.

### Who reads, who writes

| Context | Who reads from backend | Who writes to backend |
|---------|----------------------|----------------------|
| Non-SDD (general task) | **Orchestrator** searches engram, passes summary in prompt | **Sub-agent** saves discoveries/decisions via `mem_save` |
| SDD (phase with dependencies) | **Sub-agent** reads artifacts directly from backend | **Sub-agent** saves its artifact |
| SDD (phase without dependencies, e.g. explore) | Nobody | **Sub-agent** saves its artifact |

### Why this split

- **Orchestrator reads for non-SDD**: It already has the engram protocol loaded. It knows what context is relevant. Sub-agents doing their own searches waste tokens on potentially irrelevant results.
- **Sub-agents read for SDD**: SDD artifacts are large (specs, designs). The orchestrator should NOT inline them — it passes artifact references (topic keys or file paths) and the sub-agent retrieves the full content.
- **Sub-agents always write**: They have the complete detail. By the time results flow back to the orchestrator, nuance is lost. Persist at the source.

### Orchestrator prompt instructions for sub-agents

When launching a sub-agent, the orchestrator MUST include persistence instructions in the prompt:

**Non-SDD**:
```
PERSISTENCE (MANDATORY):
If you make important discoveries, decisions, or fix bugs, you MUST save them
to engram before returning:
  mem_save(title: "{short description}", type: "{decision|bugfix|discovery|pattern}",
           project: "{project}", content: "{What, Why, Where, Learned}")
Do NOT return without saving what you learned. This is how the team builds
persistent knowledge across sessions.
```

**SDD (with dependencies)**:
```
Artifact store mode: {engram|openspec|hybrid|none}
Read these artifacts before starting (two-step — search returns truncated previews):
  mem_search(query: "sdd/{change-name}/{type}", project: "{project}") → get ID
  mem_get_observation(id: {id}) → full content (REQUIRED for SDD dependencies)

PERSISTENCE (MANDATORY — do NOT skip):
After completing your work, you MUST call:
  mem_save(
    title: "sdd/{change-name}/{artifact-type}",
    topic_key: "sdd/{change-name}/{artifact-type}",
    type: "architecture",
    project: "{project}",
    content: "{your full artifact markdown}"
  )
If you return without calling mem_save, the next phase CANNOT find your artifact
and the pipeline BREAKS.
```

**SDD (no dependencies)**:
```
Artifact store mode: {engram|openspec|hybrid|none}

PERSISTENCE (MANDATORY — do NOT skip):
After completing your work, you MUST call:
  mem_save(
    title: "sdd/{change-name}/{artifact-type}",
    topic_key: "sdd/{change-name}/{artifact-type}",
    type: "architecture",
    project: "{project}",
    content: "{your full artifact markdown}"
  )
If you return without calling mem_save, the next phase CANNOT find your artifact
and the pipeline BREAKS.
```

## Skill Registry

The skill registry is a catalog of all available skills (user-level + project-level) that sub-agents read before starting any task. It is **infrastructure, not an SDD artifact** — it exists independently of any persistence mode.

### Where the registry lives

The registry is ALWAYS written to `.atl/skill-registry.md` in the project root, regardless of mode. If engram is available, it's ALSO saved there as a cross-session bonus.

| Source | Location | Priority |
|--------|----------|----------|
| Engram | `topic_key: "skill-registry"` | Read FIRST (fast, cross-session) |
| File | `.atl/skill-registry.md` | Fallback if engram not available |

### How to generate/update

Run the `skill-registry` skill, or run `sdd-init` (which includes registry generation).

### Sub-agent skill loading protocol

**EVERY sub-agent MUST check the skill registry as its FIRST step**, before starting any work:

```
1. Try engram first: mem_search(query: "skill-registry", project: "{project}")
   → if found: mem_get_observation(id) → full registry
2. If engram not available or not found: read .atl/skill-registry.md
3. If neither exists: proceed without skills (not an error)
4. From the registry, identify skills matching your task:
   - Writing React code? → load react-19
   - Reviewing a PR? → load pr-review
   - Creating a Jira task? → load jira-task
   - Writing tests? → load pytest/playwright
5. Read those specific SKILL.md files
6. Also read any project convention files listed in the registry (index files are already expanded — all referenced paths are in the table)
7. THEN proceed with your actual task
```

The orchestrator MUST include this instruction in ALL sub-agent prompts:
```
SKILL LOADING (do this FIRST):
Check for available skills:
  1. Try: mem_search(query: "skill-registry", project: "{project}")
  2. Fallback: read .atl/skill-registry.md
Load and follow any skills relevant to your task.
```

### When the registry doesn't exist

If neither engram nor the file has a registry, the sub-agent proceeds without skills. This is not an error — skills are optional enhancement. Recommend the user run `/skill-registry` or `/sdd-init` to generate it.

## Detail Level

The orchestrator may also pass `detail_level`: `concise | standard | deep`.
This controls output verbosity but does NOT affect what gets persisted — always persist the full artifact.
