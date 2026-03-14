package sdd

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gentleman-programming/gentle-ai/internal/agents"
	"github.com/gentleman-programming/gentle-ai/internal/agents/claude"
	"github.com/gentleman-programming/gentle-ai/internal/agents/opencode"
	"github.com/gentleman-programming/gentle-ai/internal/assets"
	"github.com/gentleman-programming/gentle-ai/internal/model"
	// agents/cursor, agents/gemini, agents/vscode used via agents.NewAdapter()
)

func claudeAdapter() agents.Adapter   { return claude.NewAdapter() }
func opencodeAdapter() agents.Adapter { return opencode.NewAdapter() }

func TestInjectClaudeWritesSectionMarkers(t *testing.T) {
	home := t.TempDir()

	result, err := Inject(home, claudeAdapter(), "")
	if err != nil {
		t.Fatalf("Inject() error = %v", err)
	}
	if !result.Changed {
		t.Fatalf("Inject() first changed = false")
	}

	path := filepath.Join(home, ".claude", "CLAUDE.md")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	text := string(content)

	if !strings.Contains(text, "<!-- gentle-ai:sdd-orchestrator -->") {
		t.Fatal("CLAUDE.md missing open marker for sdd-orchestrator")
	}
	if !strings.Contains(text, "<!-- /gentle-ai:sdd-orchestrator -->") {
		t.Fatal("CLAUDE.md missing close marker for sdd-orchestrator")
	}
	if !strings.Contains(text, "sub-agent") {
		t.Fatal("CLAUDE.md missing real SDD orchestrator content (expected 'sub-agent')")
	}
	if !strings.Contains(text, "dependency") {
		t.Fatal("CLAUDE.md missing real SDD orchestrator content (expected 'dependency')")
	}
}

func TestInjectClaudePreservesExistingSections(t *testing.T) {
	home := t.TempDir()
	claudeDir := filepath.Join(home, ".claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	existing := "# My Config\n\nSome user content.\n"
	if err := os.WriteFile(filepath.Join(claudeDir, "CLAUDE.md"), []byte(existing), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err := Inject(home, claudeAdapter(), "")
	if err != nil {
		t.Fatalf("Inject() error = %v", err)
	}

	content, err := os.ReadFile(filepath.Join(claudeDir, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	text := string(content)
	if !strings.Contains(text, "Some user content.") {
		t.Fatal("Existing user content was clobbered")
	}
	if !strings.Contains(text, "<!-- gentle-ai:sdd-orchestrator -->") {
		t.Fatal("SDD section was not injected")
	}
}

func TestInjectClaudeIsIdempotent(t *testing.T) {
	home := t.TempDir()

	first, err := Inject(home, claudeAdapter(), "")
	if err != nil {
		t.Fatalf("Inject() first error = %v", err)
	}
	if !first.Changed {
		t.Fatalf("Inject() first changed = false")
	}

	second, err := Inject(home, claudeAdapter(), "")
	if err != nil {
		t.Fatalf("Inject() second error = %v", err)
	}
	if second.Changed {
		t.Fatalf("Inject() second changed = true")
	}
}

func TestInjectOpenCodeWritesCommandFiles(t *testing.T) {
	home := t.TempDir()

	result, err := Inject(home, opencodeAdapter(), "")
	if err != nil {
		t.Fatalf("Inject() error = %v", err)
	}
	if !result.Changed {
		t.Fatalf("Inject() first changed = false")
	}

	if len(result.Files) == 0 {
		t.Fatal("Inject() returned no files")
	}

	commandPath := filepath.Join(home, ".config", "opencode", "commands", "sdd-init.md")
	content, err := os.ReadFile(commandPath)
	if err != nil {
		t.Fatalf("ReadFile(sdd-init.md) error = %v", err)
	}

	text := string(content)
	if !strings.Contains(text, "description") {
		t.Fatal("sdd-init.md missing frontmatter description — not real content")
	}

	settingsPath := filepath.Join(home, ".config", "opencode", "opencode.json")
	settingsContent, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("ReadFile(opencode.json) error = %v", err)
	}

	settingsText := string(settingsContent)
	if !strings.Contains(settingsText, `"agent"`) {
		t.Fatal("opencode.json missing agent key for SDD commands")
	}
	if !strings.Contains(settingsText, `"sdd-orchestrator"`) {
		t.Fatal("opencode.json missing sdd-orchestrator agent")
	}

	sharedPath := filepath.Join(home, ".config", "opencode", "skills", "_shared", "persistence-contract.md")
	if _, err := os.Stat(sharedPath); err != nil {
		t.Fatalf("expected shared SDD convention file %q: %v", sharedPath, err)
	}

	skillPath := filepath.Join(home, ".config", "opencode", "skills", "sdd-init", "SKILL.md")
	skillContent, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatalf("ReadFile(sdd-init SKILL.md) error = %v", err)
	}

	if !strings.Contains(string(skillContent), "sdd-init") {
		t.Fatal("SDD skill file missing expected content")
	}
}

func TestInjectOpenCodeIsIdempotent(t *testing.T) {
	home := t.TempDir()

	first, err := Inject(home, opencodeAdapter(), "")
	if err != nil {
		t.Fatalf("Inject() first error = %v", err)
	}
	if !first.Changed {
		t.Fatalf("Inject() first changed = false")
	}

	second, err := Inject(home, opencodeAdapter(), "")
	if err != nil {
		t.Fatalf("Inject() second error = %v", err)
	}
	if second.Changed {
		t.Fatalf("Inject() second changed = true")
	}
}

func TestInjectOpenCodeMigratesLegacyAgentsKey(t *testing.T) {
	home := t.TempDir()

	settingsPath := filepath.Join(home, ".config", "opencode", "opencode.json")
	if err := os.MkdirAll(filepath.Dir(settingsPath), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	legacy := `{
  "agents": {
    "legacy-agent": {
      "mode": "all",
      "prompt": "{file:./AGENTS.md}"
    }
  }
}`
	if err := os.WriteFile(settingsPath, []byte(legacy), 0o644); err != nil {
		t.Fatalf("WriteFile(opencode.json) error = %v", err)
	}

	if _, err := Inject(home, opencodeAdapter(), ""); err != nil {
		t.Fatalf("Inject() error = %v", err)
	}

	content, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("ReadFile(opencode.json) error = %v", err)
	}

	root := map[string]any{}
	if err := json.Unmarshal(content, &root); err != nil {
		t.Fatalf("Unmarshal(opencode.json) error = %v", err)
	}

	if _, hasLegacy := root["agents"]; hasLegacy {
		t.Fatal("opencode.json should not keep legacy agents key after migration")
	}

	agentRaw, ok := root["agent"]
	if !ok {
		t.Fatal("opencode.json missing agent key after migration")
	}

	agentMap, ok := agentRaw.(map[string]any)
	if !ok {
		t.Fatalf("opencode.json agent key has unexpected type: %T", agentRaw)
	}

	if _, ok := agentMap["legacy-agent"]; !ok {
		t.Fatal("legacy agent was not migrated under agent key")
	}
	if _, ok := agentMap["sdd-orchestrator"]; !ok {
		t.Fatal("sdd-orchestrator agent missing after merge")
	}
}

func TestInjectCursorWritesSDDOrchestratorAndSkills(t *testing.T) {
	home := t.TempDir()

	cursorAdapter, err := agents.NewAdapter("cursor")
	if err != nil {
		t.Fatalf("NewAdapter(cursor) error = %v", err)
	}

	result, injectErr := Inject(home, cursorAdapter, "")
	if injectErr != nil {
		t.Fatalf("Inject(cursor) error = %v", injectErr)
	}

	if !result.Changed {
		t.Fatal("Inject(cursor) changed = false")
	}

	// Should have SDD skill files AND the system prompt file.
	if len(result.Files) == 0 {
		t.Fatal("Inject(cursor) returned no files")
	}

	// Verify SDD orchestrator was injected into the system prompt file.
	promptPath := filepath.Join(home, ".cursor", "rules", "gentle-ai.mdc")
	content, readErr := os.ReadFile(promptPath)
	if readErr != nil {
		t.Fatalf("ReadFile(%q) error = %v", promptPath, readErr)
	}

	text := string(content)
	if !strings.Contains(text, "Spec-Driven Development") {
		t.Fatal("Cursor system prompt missing SDD orchestrator content")
	}
	if !strings.Contains(text, "sub-agent") {
		t.Fatal("Cursor system prompt missing SDD sub-agent references")
	}
}

func TestInjectGeminiWritesSDDOrchestratorAndSkills(t *testing.T) {
	home := t.TempDir()

	geminiAdapter, err := agents.NewAdapter("gemini-cli")
	if err != nil {
		t.Fatalf("NewAdapter(gemini-cli) error = %v", err)
	}

	result, injectErr := Inject(home, geminiAdapter, "")
	if injectErr != nil {
		t.Fatalf("Inject(gemini) error = %v", injectErr)
	}

	if !result.Changed {
		t.Fatal("Inject(gemini) changed = false")
	}

	// Verify SDD orchestrator was injected into GEMINI.md.
	promptPath := filepath.Join(home, ".gemini", "GEMINI.md")
	content, readErr := os.ReadFile(promptPath)
	if readErr != nil {
		t.Fatalf("ReadFile(%q) error = %v", promptPath, readErr)
	}

	text := string(content)
	if !strings.Contains(text, "Spec-Driven Development") {
		t.Fatal("Gemini system prompt missing SDD orchestrator content")
	}

	// Should also write SDD skill files.
	skillPath := filepath.Join(home, ".gemini", "skills", "sdd-init", "SKILL.md")
	if _, err := os.Stat(skillPath); err != nil {
		t.Fatalf("expected SDD skill file %q: %v", skillPath, err)
	}
}

func TestInjectVSCodeWritesSDDOrchestratorAndSkills(t *testing.T) {
	home := t.TempDir()

	vscodeAdapter, err := agents.NewAdapter("vscode-copilot")
	if err != nil {
		t.Fatalf("NewAdapter(vscode-copilot) error = %v", err)
	}

	result, injectErr := Inject(home, vscodeAdapter, "")
	if injectErr != nil {
		t.Fatalf("Inject(vscode) error = %v", injectErr)
	}

	if !result.Changed {
		t.Fatal("Inject(vscode) changed = false")
	}

	// Verify SDD orchestrator was injected into the VS Code instructions file.
	promptPath := vscodeAdapter.SystemPromptFile(home)
	content, readErr := os.ReadFile(promptPath)
	if readErr != nil {
		t.Fatalf("ReadFile(%q) error = %v", promptPath, readErr)
	}

	text := string(content)
	if !strings.Contains(text, "Spec-Driven Development") {
		t.Fatal("VS Code system prompt missing SDD orchestrator content")
	}

	// Should also write SDD skill files under ~/.copilot/skills/.
	skillPath := filepath.Join(home, ".copilot", "skills", "sdd-init", "SKILL.md")
	if _, err := os.Stat(skillPath); err != nil {
		t.Fatalf("expected SDD skill file %q: %v", skillPath, err)
	}

	sharedPath := filepath.Join(home, ".copilot", "skills", "_shared", "engram-convention.md")
	if _, err := os.Stat(sharedPath); err != nil {
		t.Fatalf("expected shared SDD convention file %q: %v", sharedPath, err)
	}
}

func TestInjectFileAppendSkipsIfAlreadyPresent(t *testing.T) {
	home := t.TempDir()

	cursorAdapter, err := agents.NewAdapter("cursor")
	if err != nil {
		t.Fatalf("NewAdapter(cursor) error = %v", err)
	}

	// First injection.
	first, firstErr := Inject(home, cursorAdapter, "")
	if firstErr != nil {
		t.Fatalf("Inject() first error = %v", firstErr)
	}
	if !first.Changed {
		t.Fatal("first Inject() changed = false")
	}

	// Second injection — SDD content is already there, should not duplicate.
	second, secondErr := Inject(home, cursorAdapter, "")
	if secondErr != nil {
		t.Fatalf("Inject() second error = %v", secondErr)
	}
	if second.Changed {
		t.Fatal("second Inject() changed = true — SDD orchestrator was duplicated")
	}
}

func TestInjectFileAppendSkipsLegacyHeading(t *testing.T) {
	home := t.TempDir()

	cursorAdapter, err := agents.NewAdapter("cursor")
	if err != nil {
		t.Fatalf("NewAdapter(cursor) error = %v", err)
	}

	promptPath := cursorAdapter.SystemPromptFile(home)
	if err := os.MkdirAll(filepath.Dir(promptPath), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	existing := "# Existing\n\n## Spec-Driven Development (SDD) Orchestrator\nAlready present.\n"
	if err := os.WriteFile(promptPath, []byte(existing), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	result, injectErr := Inject(home, cursorAdapter, "")
	if injectErr != nil {
		t.Fatalf("Inject() error = %v", injectErr)
	}
	if len(result.Files) == 0 {
		t.Fatal("Inject() returned no files")
	}

	content, readErr := os.ReadFile(promptPath)
	if readErr != nil {
		t.Fatalf("ReadFile() error = %v", readErr)
	}

	text := string(content)
	if strings.Count(text, "## Spec-Driven Development (SDD) Orchestrator") != 1 {
		t.Fatal("legacy SDD heading duplicated")
	}
}

func TestInjectOpenCodeMultiMode(t *testing.T) {
	home := t.TempDir()

	result, err := Inject(home, opencodeAdapter(), "multi")
	if err != nil {
		t.Fatalf("Inject(multi) error = %v", err)
	}
	if !result.Changed {
		t.Fatal("Inject(multi) changed = false")
	}

	settingsPath := filepath.Join(home, ".config", "opencode", "opencode.json")
	content, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("ReadFile(opencode.json) error = %v", err)
	}

	root := map[string]any{}
	if err := json.Unmarshal(content, &root); err != nil {
		t.Fatalf("Unmarshal(opencode.json) error = %v", err)
	}

	agentRaw, ok := root["agent"]
	if !ok {
		t.Fatal("opencode.json missing agent key")
	}

	agentMap, ok := agentRaw.(map[string]any)
	if !ok {
		t.Fatalf("agent key has unexpected type: %T", agentRaw)
	}

	// Multi overlay must contain orchestrator + 9 sub-agents = 10 agents.
	if len(agentMap) != 10 {
		t.Fatalf("agent count = %d, want 10", len(agentMap))
	}

	// Verify orchestrator is present.
	if _, ok := agentMap["sdd-orchestrator"]; !ok {
		t.Fatal("missing sdd-orchestrator agent")
	}

	// Verify representative sub-agents are present.
	for _, subAgent := range []string{"sdd-init", "sdd-apply", "sdd-verify", "sdd-explore", "sdd-propose", "sdd-spec", "sdd-design", "sdd-tasks", "sdd-archive"} {
		if _, ok := agentMap[subAgent]; !ok {
			t.Fatalf("missing sub-agent %q", subAgent)
		}
	}

	// Verify sub-agents have mode "subagent".
	applyRaw, _ := agentMap["sdd-apply"]
	applyAgent, ok := applyRaw.(map[string]any)
	if !ok {
		t.Fatalf("sdd-apply has unexpected type: %T", applyRaw)
	}
	if mode, _ := applyAgent["mode"].(string); mode != "subagent" {
		t.Fatalf("sdd-apply mode = %q, want %q", mode, "subagent")
	}
}

func TestInjectOpenCodeMultiModeIdempotent(t *testing.T) {
	home := t.TempDir()

	first, err := Inject(home, opencodeAdapter(), "multi")
	if err != nil {
		t.Fatalf("Inject(multi) first error = %v", err)
	}
	if !first.Changed {
		t.Fatal("Inject(multi) first changed = false")
	}

	second, err := Inject(home, opencodeAdapter(), "multi")
	if err != nil {
		t.Fatalf("Inject(multi) second error = %v", err)
	}
	if second.Changed {
		t.Fatal("Inject(multi) second changed = true — multi overlay was duplicated")
	}
}

func TestInjectOpenCodeEmptySDDModeDefaultsSingle(t *testing.T) {
	home := t.TempDir()

	result, err := Inject(home, opencodeAdapter(), "")
	if err != nil {
		t.Fatalf("Inject(\"\") error = %v", err)
	}
	if !result.Changed {
		t.Fatal("Inject(\"\") changed = false")
	}

	settingsPath := filepath.Join(home, ".config", "opencode", "opencode.json")
	content, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("ReadFile(opencode.json) error = %v", err)
	}

	root := map[string]any{}
	if err := json.Unmarshal(content, &root); err != nil {
		t.Fatalf("Unmarshal(opencode.json) error = %v", err)
	}

	agentRaw, ok := root["agent"]
	if !ok {
		t.Fatal("opencode.json missing agent key")
	}

	agentMap, ok := agentRaw.(map[string]any)
	if !ok {
		t.Fatalf("agent key has unexpected type: %T", agentRaw)
	}

	// Empty mode defaults to single — only sdd-orchestrator, no sub-agents.
	if _, ok := agentMap["sdd-orchestrator"]; !ok {
		t.Fatal("missing sdd-orchestrator agent")
	}
	if _, ok := agentMap["sdd-apply"]; ok {
		t.Fatal("sdd-apply sub-agent should NOT be present in single/default mode")
	}
}

func TestInjectClaudeIgnoresSDDMode(t *testing.T) {
	home := t.TempDir()

	// Inject with multi mode for Claude — should be ignored.
	resultMulti, err := Inject(home, claudeAdapter(), "multi")
	if err != nil {
		t.Fatalf("Inject(claude, multi) error = %v", err)
	}

	homeBaseline := t.TempDir()
	resultSingle, err := Inject(homeBaseline, claudeAdapter(), "single")
	if err != nil {
		t.Fatalf("Inject(claude, single) error = %v", err)
	}

	// Both should produce changed=true (first injection).
	if !resultMulti.Changed || !resultSingle.Changed {
		t.Fatal("first injection should be changed=true")
	}

	// Read and compare the CLAUDE.md files — content should be identical.
	multiContent, err := os.ReadFile(filepath.Join(home, ".claude", "CLAUDE.md"))
	if err != nil {
		t.Fatalf("ReadFile(multi) error = %v", err)
	}
	singleContent, err := os.ReadFile(filepath.Join(homeBaseline, ".claude", "CLAUDE.md"))
	if err != nil {
		t.Fatalf("ReadFile(single) error = %v", err)
	}

	if string(multiContent) != string(singleContent) {
		t.Fatal("Claude CLAUDE.md differs between multi and single sddMode — non-OpenCode agents should ignore sddMode")
	}
}

func TestInjectOpenCodeSingleToMultiSwitch(t *testing.T) {
	home := t.TempDir()

	// First: inject single mode.
	_, err := Inject(home, opencodeAdapter(), "single")
	if err != nil {
		t.Fatalf("Inject(single) error = %v", err)
	}

	settingsPath := filepath.Join(home, ".config", "opencode", "opencode.json")

	// Verify only orchestrator, no sub-agents.
	content, _ := os.ReadFile(settingsPath)
	if strings.Contains(string(content), `"sdd-apply"`) {
		t.Fatal("single mode should not have sdd-apply")
	}

	// Second: inject multi mode.
	result, err := Inject(home, opencodeAdapter(), "multi")
	if err != nil {
		t.Fatalf("Inject(multi) error = %v", err)
	}
	if !result.Changed {
		t.Fatal("switching from single to multi should produce changed=true")
	}

	content, err = os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("ReadFile(opencode.json) error = %v", err)
	}

	root := map[string]any{}
	if err := json.Unmarshal(content, &root); err != nil {
		t.Fatalf("Unmarshal(opencode.json) error = %v", err)
	}

	agentMap, _ := root["agent"].(map[string]any)
	if _, ok := agentMap["sdd-orchestrator"]; !ok {
		t.Fatal("missing sdd-orchestrator after switch to multi")
	}
	if _, ok := agentMap["sdd-apply"]; !ok {
		t.Fatal("missing sdd-apply after switch to multi")
	}
}

func TestInjectFileAppendSkipsAgentTeamsHeading(t *testing.T) {
	home := t.TempDir()

	cursorAdapter, err := agents.NewAdapter("cursor")
	if err != nil {
		t.Fatalf("NewAdapter(cursor) error = %v", err)
	}

	promptPath := cursorAdapter.SystemPromptFile(home)
	if err := os.MkdirAll(filepath.Dir(promptPath), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	existing := "# Existing\n\n## Agent Teams Orchestrator\nAlready present.\n"
	if err := os.WriteFile(promptPath, []byte(existing), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	result, injectErr := Inject(home, cursorAdapter, "")
	if injectErr != nil {
		t.Fatalf("Inject() error = %v", injectErr)
	}
	if len(result.Files) == 0 {
		t.Fatal("Inject() returned no files")
	}

	content, readErr := os.ReadFile(promptPath)
	if readErr != nil {
		t.Fatalf("ReadFile() error = %v", readErr)
	}

	text := string(content)
	if strings.Count(text, "## Agent Teams Orchestrator") != 1 {
		t.Fatal("agent teams heading duplicated")
	}
}

func TestInjectOpenCodeMultiModeWithModelAssignments(t *testing.T) {
	home := t.TempDir()

	assignments := map[string]model.ModelAssignment{
		"sdd-init":  {ProviderID: "anthropic", ModelID: "claude-sonnet-4-20250514"},
		"sdd-apply": {ProviderID: "openai", ModelID: "gpt-4o"},
	}

	result, err := Inject(home, opencodeAdapter(), "multi", assignments)
	if err != nil {
		t.Fatalf("Inject(multi, assignments) error = %v", err)
	}
	if !result.Changed {
		t.Fatal("Inject(multi, assignments) changed = false")
	}

	settingsPath := filepath.Join(home, ".config", "opencode", "opencode.json")
	content, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("ReadFile(opencode.json) error = %v", err)
	}

	root := map[string]any{}
	if err := json.Unmarshal(content, &root); err != nil {
		t.Fatalf("Unmarshal(opencode.json) error = %v", err)
	}

	agentMap, ok := root["agent"].(map[string]any)
	if !ok {
		t.Fatal("opencode.json missing agent map")
	}

	// Verify sdd-init has the assigned model.
	initAgent, ok := agentMap["sdd-init"].(map[string]any)
	if !ok {
		t.Fatal("sdd-init agent not found or wrong type")
	}
	if m, _ := initAgent["model"].(string); m != "anthropic/claude-sonnet-4-20250514" {
		t.Fatalf("sdd-init model = %q, want %q", m, "anthropic/claude-sonnet-4-20250514")
	}

	// Verify sdd-apply has the assigned model.
	applyAgent, ok := agentMap["sdd-apply"].(map[string]any)
	if !ok {
		t.Fatal("sdd-apply agent not found or wrong type")
	}
	if m, _ := applyAgent["model"].(string); m != "openai/gpt-4o" {
		t.Fatalf("sdd-apply model = %q, want %q", m, "openai/gpt-4o")
	}

	// Verify unassigned phases do NOT have a model field.
	verifyAgent, ok := agentMap["sdd-verify"].(map[string]any)
	if !ok {
		t.Fatal("sdd-verify agent not found or wrong type")
	}
	if _, hasModel := verifyAgent["model"]; hasModel {
		t.Fatal("sdd-verify should not have a model field when not assigned")
	}
}

func TestInjectOpenCodeMultiModeNoAssignmentsNoModel(t *testing.T) {
	home := t.TempDir()

	// Pass nil assignments — no model fields should be injected.
	result, err := Inject(home, opencodeAdapter(), "multi")
	if err != nil {
		t.Fatalf("Inject(multi) error = %v", err)
	}
	if !result.Changed {
		t.Fatal("Inject(multi) changed = false")
	}

	settingsPath := filepath.Join(home, ".config", "opencode", "opencode.json")
	content, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("ReadFile(opencode.json) error = %v", err)
	}

	root := map[string]any{}
	if err := json.Unmarshal(content, &root); err != nil {
		t.Fatalf("Unmarshal(opencode.json) error = %v", err)
	}

	agentMap, _ := root["agent"].(map[string]any)
	for _, phase := range []string{"sdd-init", "sdd-apply", "sdd-verify"} {
		agentDef, ok := agentMap[phase].(map[string]any)
		if !ok {
			continue
		}
		if _, hasModel := agentDef["model"]; hasModel {
			t.Fatalf("phase %q should not have model field when no assignments given", phase)
		}
	}
}

func TestInjectSingleModeIgnoresModelAssignments(t *testing.T) {
	home := t.TempDir()

	// Even if assignments are provided, single mode should ignore them.
	assignments := map[string]model.ModelAssignment{
		"sdd-init": {ProviderID: "anthropic", ModelID: "claude-sonnet-4-20250514"},
	}

	result, err := Inject(home, opencodeAdapter(), "single", assignments)
	if err != nil {
		t.Fatalf("Inject(single, assignments) error = %v", err)
	}
	if !result.Changed {
		t.Fatal("Inject(single, assignments) changed = false")
	}

	settingsPath := filepath.Join(home, ".config", "opencode", "opencode.json")
	content, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("ReadFile(opencode.json) error = %v", err)
	}

	// Single mode has no sub-agents, so model should not appear.
	if strings.Contains(string(content), `"model"`) {
		t.Fatal("single mode should not inject model assignments")
	}
}

func TestCommandsAssetDir(t *testing.T) {
	tests := []struct {
		agentID model.AgentID
		want    string
	}{
		{model.AgentClaudeCode, "claude/commands"},
		{model.AgentOpenCode, "opencode/commands"},
		{model.AgentCursor, ""},
		{model.AgentGeminiCLI, ""},
		{model.AgentVSCodeCopilot, ""},
	}
	for _, tt := range tests {
		if got := commandsAssetDir(tt.agentID); got != tt.want {
			t.Errorf("commandsAssetDir(%q) = %q, want %q", tt.agentID, got, tt.want)
		}
	}
}

func TestSDDCommandNames(t *testing.T) {
	names := SDDCommandNames()
	if len(names) != 3 {
		t.Fatalf("SDDCommandNames() returned %d names, want 3: %v", len(names), names)
	}

	wantNames := map[string]bool{
		"sdd-new":      true,
		"sdd-continue": true,
		"sdd-ff":       true,
	}

	for _, name := range names {
		if !wantNames[name] {
			t.Errorf("SDDCommandNames() contains unexpected name %q", name)
		}
	}

	// Cross-check against the embedded asset files.
	entries, err := fs.ReadDir(assets.FS, "claude/commands")
	if err != nil {
		t.Fatalf("fs.ReadDir(claude/commands) error = %v", err)
	}
	fileCount := 0
	for _, e := range entries {
		if !e.IsDir() {
			fileCount++
		}
	}
	if fileCount != len(names) {
		t.Errorf("SDDCommandNames() has %d entries but claude/commands has %d files — lists are out of sync", len(names), fileCount)
	}
}

func TestInjectClaudeWritesCommandFiles(t *testing.T) {
	home := t.TempDir()

	result, err := Inject(home, claudeAdapter(), "")
	if err != nil {
		t.Fatalf("Inject() error = %v", err)
	}
	if !result.Changed {
		t.Fatalf("Inject() changed = false")
	}

	expectedCommands := SDDCommandNames()
	for _, cmd := range expectedCommands {
		path := filepath.Join(home, ".claude", "commands", cmd+".md")
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("ReadFile(%s.md) error = %v", cmd, err)
		}
		text := string(content)

		if !strings.Contains(text, "description:") {
			t.Errorf("%s.md missing description frontmatter", cmd)
		}
		if strings.Contains(text, "agent:") {
			t.Errorf("%s.md contains agent: field — should not be in Claude commands", cmd)
		}
		if strings.Contains(text, "subtask:") {
			t.Errorf("%s.md contains subtask: field — should not be in Claude commands", cmd)
		}
	}

	// Verify skill path references use Claude path, not OpenCode.
	newPath := filepath.Join(home, ".claude", "commands", "sdd-new.md")
	newContent, err := os.ReadFile(newPath)
	if err != nil {
		t.Fatalf("ReadFile(sdd-new.md) error = %v", err)
	}
	if strings.Contains(string(newContent), "opencode/skills") {
		t.Fatal("sdd-new.md references opencode skills path — should use ~/.claude/skills/")
	}
}

func TestInjectClaudeCommandsIdempotent(t *testing.T) {
	home := t.TempDir()

	first, err := Inject(home, claudeAdapter(), "")
	if err != nil {
		t.Fatalf("Inject() first error = %v", err)
	}
	if !first.Changed {
		t.Fatalf("Inject() first changed = false")
	}

	second, err := Inject(home, claudeAdapter(), "")
	if err != nil {
		t.Fatalf("Inject() second error = %v", err)
	}
	if second.Changed {
		t.Fatalf("Inject() second changed = true — command files changed on second run")
	}

	// Verify files are still present and unchanged.
	for _, cmd := range SDDCommandNames() {
		path := filepath.Join(home, ".claude", "commands", cmd+".md")
		if _, err := os.Stat(path); err != nil {
			t.Errorf("command file %s.md missing after second inject: %v", cmd, err)
		}
	}
}

func TestInjectModelAssignmentsFunction(t *testing.T) {
	overlayJSON := []byte(`{
  "agent": {
    "sdd-init": {"mode": "subagent", "prompt": "test"},
    "sdd-apply": {"mode": "subagent", "prompt": "test"}
  }
}`)

	assignments := map[string]model.ModelAssignment{
		"sdd-init": {ProviderID: "anthropic", ModelID: "claude-sonnet-4-20250514"},
	}

	result, err := injectModelAssignments(overlayJSON, assignments)
	if err != nil {
		t.Fatalf("injectModelAssignments() error = %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("Unmarshal result error = %v", err)
	}

	agents := parsed["agent"].(map[string]any)
	initAgent := agents["sdd-init"].(map[string]any)
	if m, _ := initAgent["model"].(string); m != "anthropic/claude-sonnet-4-20250514" {
		t.Fatalf("sdd-init model = %q, want %q", m, "anthropic/claude-sonnet-4-20250514")
	}

	// sdd-apply should NOT have a model field.
	applyAgent := agents["sdd-apply"].(map[string]any)
	if _, ok := applyAgent["model"]; ok {
		t.Fatal("sdd-apply should not have model field when not in assignments")
	}
}
