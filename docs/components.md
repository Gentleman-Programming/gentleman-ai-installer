# Components, Skills & Presets

← [Back to README](../README.md)

---

## Components

| Component | ID | Description |
|-----------|-----|-------------|
| Engram | `engram` | Persistent cross-session memory |
| SDD | `sdd` | Spec-Driven Development workflow (9 phases) |
| Skills | `skills` | Curated coding skill library |
| Context7 | `context7` | MCP server for live framework/library documentation |
| Persona | `persona` | Gentleman, neutral, or custom behavior mode |
| Permissions | `permissions` | Security-first defaults and guardrails |
| GGA | `gga` | Gentleman Guardian Angel — AI provider switcher |
| Theme | `theme` | Gentleman Kanagawa theme overlay |

## GGA Behavior

`gentle-ai --component gga` installs/provisions the `gga` binary globally on your machine.

It does **not** run project-level hook setup automatically (`gga init` / `gga install`) because that should be an explicit decision per repository.

After global install, enable GGA per project with:

```bash
gga init
gga install
```

---

## Skills

11 curated skill files organized by category, injected into your agent's configuration:

### SDD (Spec-Driven Development)

| Skill | ID | Description |
|-------|-----|-------------|
| SDD Init | `sdd-init` | Bootstrap SDD context in a project |
| SDD Explore | `sdd-explore` | Investigate codebase before committing to a change |
| SDD Propose | `sdd-propose` | Create change proposal with intent, scope, approach |
| SDD Spec | `sdd-spec` | Write specifications with requirements and scenarios |
| SDD Design | `sdd-design` | Technical design with architecture decisions |
| SDD Tasks | `sdd-tasks` | Break down a change into implementation tasks |
| SDD Apply | `sdd-apply` | Implement tasks following specs and design |
| SDD Verify | `sdd-verify` | Validate implementation matches specs |
| SDD Archive | `sdd-archive` | Sync delta specs to main specs and archive |

### Foundation

| Skill | ID | Description |
|-------|-----|-------------|
| Go Testing | `go-testing` | Go testing patterns including Bubbletea TUI testing |
| Skill Creator | `skill-creator` | Create new AI agent skills following the Agent Skills spec |

These foundation skills are installed by default with both `full-gentleman` and `ecosystem-only` presets.

---

## Presets

| Preset | ID | What's Included |
|--------|-----|-----------------|
| Full Gentleman | `full-gentleman` | All components + all skills + gentleman persona |
| Ecosystem Only | `ecosystem-only` | All components + P0 skills + gentleman persona |
| Minimal | `minimal` | Engram + Persona + Permissions only |
| Custom | `custom` | You pick components, skills, and persona individually |
