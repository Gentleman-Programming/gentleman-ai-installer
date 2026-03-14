---
name: skill-registry
description: >
  Create or update the skill registry for the current project. Scans user skills and project conventions, writes .atl/skill-registry.md, and saves to engram if available.
  Trigger: When user says "update skills", "skill registry", "actualizar skills", "update registry", or after installing/removing skills.
license: MIT
metadata:
  author: gentleman-programming
  version: "1.0"
---

## Purpose

You generate or update the **skill registry** — a catalog of all available skills (user-level and project-level) that sub-agents read before starting any task. This ensures every sub-agent knows what skills exist and can load the relevant ones.

## When to Run

- After installing or removing skills
- After setting up a new project
- When the user explicitly asks to update the registry
- As part of `sdd-init` (it calls this same logic)

## What to Do

### Step 1: Scan User Skills

1. Glob for `*/SKILL.md` files across ALL known skill directories. Check every path below — scan ALL that exist, not just the first match:

   **User-level (global skills):**
   - `~/.claude/skills/` — Claude Code
   - `~/.config/opencode/skills/` — OpenCode
   - `~/.gemini/skills/` — Gemini CLI
   - `~/.cursor/skills/` — Cursor
   - `~/.copilot/skills/` — VS Code Copilot
   - The parent directory of this skill file (catch-all for any tool)

   **Project-level (workspace skills):**
   - `{project-root}/.claude/skills/` — Claude Code
   - `{project-root}/.gemini/skills/` — Gemini CLI
   - `{project-root}/.agent/skills/` — Antigravity (workspace)
   - `{project-root}/skills/` — Generic

2. **SKIP `sdd-*` and `_shared`** — those are SDD workflow skills, not coding/task skills
3. Also **SKIP `skill-registry`** — that's this skill
4. **Deduplicate** — if the same skill name appears in multiple locations, keep the project-level version (more specific). If both are user-level, keep the first found.
5. For each skill found, read only the frontmatter (first 10 lines) to extract:
   - `name` field
   - `description` field → extract the trigger text (after "Trigger:" in the description)
6. Build a table of: Trigger | Skill Name | Full Path

### Step 2: Scan Project Conventions

1. Check the project root for convention files. Look for:
   - `agents.md` or `AGENTS.md`
   - `CLAUDE.md` (only project-level, not `~/.claude/CLAUDE.md`)
   - `.cursorrules`
   - `GEMINI.md`
   - `copilot-instructions.md`
2. **If an index file is found** (e.g., `agents.md`, `AGENTS.md`): READ its contents and extract all referenced file paths. These index files typically list project conventions with paths — extract every referenced path and include it in the registry table alongside the index file itself.
3. For non-index files (`.cursorrules`, `CLAUDE.md`, etc.): record the file directly.
4. The final table should include the index file AND all paths it references — zero extra hops for sub-agents.

### Step 3: Write the Registry

Build the registry markdown:

```markdown
# Skill Registry

As your FIRST step before starting any work, identify and load skills relevant to your task from this registry.

## User Skills

| Trigger | Skill | Path |
|---------|-------|------|
| {trigger from frontmatter} | {skill name} | {full path to SKILL.md} |
| ... | ... | ... |

## Project Conventions

| File | Path | Notes |
|------|------|-------|
| {index file} | {path} | Index — references files below |
| {referenced file} | {extracted path} | Referenced by {index file} |
| {standalone file} | {path} | |

Read the convention files listed above for project-specific patterns and rules. All referenced paths have been extracted — no need to read index files to discover more.
```

### Step 4: Persist the Registry

**This step is MANDATORY — do NOT skip it.**

#### A. Always write the file (guaranteed availability):

Create the `.atl/` directory in the project root if it doesn't exist, then write:

```
.atl/skill-registry.md
```

#### B. If engram is available, also save to engram (cross-session bonus):

```
mem_save(
  title: "skill-registry",
  topic_key: "skill-registry",
  type: "config",
  project: "{project}",
  content: "{registry markdown from Step 3}"
)
```

`topic_key` ensures upserts — running again updates the same observation.

### Step 5: Return Summary

```markdown
## Skill Registry Updated

**Project**: {project name}
**Location**: .atl/skill-registry.md
**Engram**: {saved / not available}

### User Skills Found
| Skill | Trigger |
|-------|---------|
| {name} | {trigger} |
| ... | ... |

### Project Conventions Found
| File | Path |
|------|------|
| {file} | {path} |

### Next Steps
Sub-agents will automatically load relevant skills from this registry.
To update after installing/removing skills, run this again.
```

## Rules

- ALWAYS write `.atl/skill-registry.md` regardless of any SDD persistence mode
- ALWAYS save to engram if the `mem_save` tool is available
- SKIP `sdd-*`, `_shared`, and `skill-registry` directories when scanning
- Only read frontmatter (first 10 lines) — do NOT read full skill files
- Include ALL convention index files found (not just the first)
- If no skills or conventions are found, write an empty registry (so sub-agents don't waste time searching)
- Add `.atl/` to the project's `.gitignore` if it exists and `.atl` is not already listed
