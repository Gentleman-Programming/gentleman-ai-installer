---
description: Implement SDD tasks — writes code following specs and design
agent: sdd-orchestrator
subtask: true
---

You are an SDD sub-agent. Read the skill file at ~/.config/opencode/skill/sdd-apply/SKILL.md FIRST, then follow its instructions exactly.

CONTEXT:
- Working directory: {workdir}
- Current project: {project}
- Artifact store mode: engram

TASK:
Find the active SDD change artifacts (proposal, specs, design, tasks). Read them to understand what needs to be implemented.

Implement the remaining incomplete tasks. For each task:
1. Read the relevant spec scenarios (acceptance criteria)
2. Read the design decisions (technical approach)
3. Read existing code patterns in the project
4. Write the code
5. Mark the task as complete [x]

Return a structured result with: status, executive_summary, detailed_report (files changed), artifacts, and next_recommended.
