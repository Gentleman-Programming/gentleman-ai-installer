package sdd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gentleman-programming/gentle-ai/internal/agents"
	"github.com/gentleman-programming/gentle-ai/internal/agents/claude"
	"github.com/gentleman-programming/gentle-ai/internal/agents/opencode"
	// agents/cursor, agents/gemini, agents/vscode used via agents.NewAdapter()
)

func claudeAdapter() agents.Adapter   { return claude.NewAdapter() }
func opencodeAdapter() agents.Adapter { return opencode.NewAdapter() }

func TestInjectClaudeWritesSectionMarkers(t *testing.T) {
	home := t.TempDir()

	result, err := Inject(home, claudeAdapter())
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

	_, err := Inject(home, claudeAdapter())
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

	first, err := Inject(home, claudeAdapter())
	if err != nil {
		t.Fatalf("Inject() first error = %v", err)
	}
	if !first.Changed {
		t.Fatalf("Inject() first changed = false")
	}

	second, err := Inject(home, claudeAdapter())
	if err != nil {
		t.Fatalf("Inject() second error = %v", err)
	}
	if second.Changed {
		t.Fatalf("Inject() second changed = true")
	}
}

func TestInjectOpenCodeWritesCommandFiles(t *testing.T) {
	home := t.TempDir()

	result, err := Inject(home, opencodeAdapter())
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

	skillPath := filepath.Join(home, ".config", "opencode", "skill", "sdd-init", "SKILL.md")
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

	first, err := Inject(home, opencodeAdapter())
	if err != nil {
		t.Fatalf("Inject() first error = %v", err)
	}
	if !first.Changed {
		t.Fatalf("Inject() first changed = false")
	}

	second, err := Inject(home, opencodeAdapter())
	if err != nil {
		t.Fatalf("Inject() second error = %v", err)
	}
	if second.Changed {
		t.Fatalf("Inject() second changed = true")
	}
}

func TestInjectCursorWritesSDDOrchestratorAndSkills(t *testing.T) {
	home := t.TempDir()

	cursorAdapter, err := agents.NewAdapter("cursor")
	if err != nil {
		t.Fatalf("NewAdapter(cursor) error = %v", err)
	}

	result, injectErr := Inject(home, cursorAdapter)
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

	result, injectErr := Inject(home, geminiAdapter)
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

	result, injectErr := Inject(home, vscodeAdapter)
	if injectErr != nil {
		t.Fatalf("Inject(vscode) error = %v", injectErr)
	}

	if !result.Changed {
		t.Fatal("Inject(vscode) changed = false")
	}

	// Verify SDD orchestrator was injected into copilot-instructions.md.
	promptPath := filepath.Join(home, ".github", "copilot-instructions.md")
	content, readErr := os.ReadFile(promptPath)
	if readErr != nil {
		t.Fatalf("ReadFile(%q) error = %v", promptPath, readErr)
	}

	text := string(content)
	if !strings.Contains(text, "Spec-Driven Development") {
		t.Fatal("VS Code system prompt missing SDD orchestrator content")
	}

	// Should also write SDD skill files under ~/.vscode/skills/.
	skillPath := filepath.Join(home, ".vscode", "skills", "sdd-init", "SKILL.md")
	if _, err := os.Stat(skillPath); err != nil {
		t.Fatalf("expected SDD skill file %q: %v", skillPath, err)
	}
}

func TestInjectFileAppendSkipsIfAlreadyPresent(t *testing.T) {
	home := t.TempDir()

	cursorAdapter, err := agents.NewAdapter("cursor")
	if err != nil {
		t.Fatalf("NewAdapter(cursor) error = %v", err)
	}

	// First injection.
	first, firstErr := Inject(home, cursorAdapter)
	if firstErr != nil {
		t.Fatalf("Inject() first error = %v", firstErr)
	}
	if !first.Changed {
		t.Fatal("first Inject() changed = false")
	}

	// Second injection — SDD content is already there, should not duplicate.
	second, secondErr := Inject(home, cursorAdapter)
	if secondErr != nil {
		t.Fatalf("Inject() second error = %v", secondErr)
	}
	if second.Changed {
		t.Fatal("second Inject() changed = true — SDD orchestrator was duplicated")
	}
}
