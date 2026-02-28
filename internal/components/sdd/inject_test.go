package sdd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gentleman-programming/gentle-ai/internal/model"
)

func TestInjectClaudeWritesSectionMarkers(t *testing.T) {
	home := t.TempDir()

	result, err := Inject(home, model.AgentClaudeCode)
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
	// Real content checks — the embedded asset has these patterns.
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

	_, err := Inject(home, model.AgentClaudeCode)
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

	first, err := Inject(home, model.AgentClaudeCode)
	if err != nil {
		t.Fatalf("Inject() first error = %v", err)
	}
	if !first.Changed {
		t.Fatalf("Inject() first changed = false")
	}

	second, err := Inject(home, model.AgentClaudeCode)
	if err != nil {
		t.Fatalf("Inject() second error = %v", err)
	}
	if second.Changed {
		t.Fatalf("Inject() second changed = true")
	}
}

func TestInjectOpenCodeWritesCommandFiles(t *testing.T) {
	home := t.TempDir()

	result, err := Inject(home, model.AgentOpenCode)
	if err != nil {
		t.Fatalf("Inject() error = %v", err)
	}
	if !result.Changed {
		t.Fatalf("Inject() first changed = false")
	}

	// Should have command files + SDD skill files.
	if len(result.Files) == 0 {
		t.Fatal("Inject() returned no files")
	}

	// Check a command file has real content.
	commandPath := filepath.Join(home, ".config", "opencode", "commands", "sdd-init.md")
	content, err := os.ReadFile(commandPath)
	if err != nil {
		t.Fatalf("ReadFile(sdd-init.md) error = %v", err)
	}

	text := string(content)
	if !strings.Contains(text, "description") {
		t.Fatal("sdd-init.md missing frontmatter description — not real content")
	}

	// Check a skill file was written.
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

	first, err := Inject(home, model.AgentOpenCode)
	if err != nil {
		t.Fatalf("Inject() first error = %v", err)
	}
	if !first.Changed {
		t.Fatalf("Inject() first changed = false")
	}

	second, err := Inject(home, model.AgentOpenCode)
	if err != nil {
		t.Fatalf("Inject() second error = %v", err)
	}
	if second.Changed {
		t.Fatalf("Inject() second changed = true")
	}
}

func TestInjectUnsupportedAgentReturnsError(t *testing.T) {
	home := t.TempDir()

	_, err := Inject(home, model.AgentID("cursor"))
	if err == nil {
		t.Fatal("expected error for unsupported agent")
	}
}
