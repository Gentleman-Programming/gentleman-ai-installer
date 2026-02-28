package skills

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gentleman-programming/gentle-ai/internal/model"
)

func TestInjectWritesSkillFilesForOpenCode(t *testing.T) {
	home := t.TempDir()

	result, err := Inject(home, model.AgentOpenCode, []model.SkillID{model.SkillTypeScript})
	if err != nil {
		t.Fatalf("Inject() error = %v", err)
	}
	if !result.Changed {
		t.Fatalf("Inject() first changed = false")
	}

	if len(result.Files) != 1 {
		t.Fatalf("Inject() files len = %d", len(result.Files))
	}

	path := filepath.Join(home, ".config", "opencode", "skill", "typescript", "SKILL.md")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected skill file %q: %v", path, err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if len(content) == 0 {
		t.Fatalf("skill file is empty")
	}

	// Idempotent: second inject should not change.
	second, err := Inject(home, model.AgentOpenCode, []model.SkillID{model.SkillTypeScript})
	if err != nil {
		t.Fatalf("Inject() second error = %v", err)
	}

	if second.Changed {
		t.Fatalf("Inject() second changed = true")
	}
}

func TestInjectWritesSkillFilesForClaude(t *testing.T) {
	home := t.TempDir()

	result, err := Inject(home, model.AgentClaudeCode, []model.SkillID{model.SkillReact19, model.SkillSDDInit})
	if err != nil {
		t.Fatalf("Inject() error = %v", err)
	}
	if !result.Changed {
		t.Fatalf("Inject() changed = false")
	}

	if len(result.Files) != 2 {
		t.Fatalf("Inject() files len = %d, want 2", len(result.Files))
	}

	for _, id := range []model.SkillID{model.SkillReact19, model.SkillSDDInit} {
		path := filepath.Join(home, ".claude", "skills", string(id), "SKILL.md")
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected skill file %q: %v", path, err)
		}
	}
}

func TestInjectSkipsUnknownSkillGracefully(t *testing.T) {
	home := t.TempDir()

	result, err := Inject(home, model.AgentOpenCode, []model.SkillID{
		model.SkillTypeScript,
		model.SkillID("nonexistent-skill"),
	})
	if err != nil {
		t.Fatalf("Inject() error = %v", err)
	}

	if len(result.Files) != 1 {
		t.Fatalf("Inject() files len = %d, want 1", len(result.Files))
	}

	if len(result.Skipped) != 1 {
		t.Fatalf("Inject() skipped len = %d, want 1", len(result.Skipped))
	}

	if result.Skipped[0] != "nonexistent-skill" {
		t.Fatalf("Inject() skipped[0] = %q, want nonexistent-skill", result.Skipped[0])
	}
}

func TestInjectRejectsUnsupportedAgent(t *testing.T) {
	home := t.TempDir()

	_, err := Inject(home, model.AgentID("cursor"), []model.SkillID{model.SkillTypeScript})
	if err == nil {
		t.Fatalf("Inject() expected error for unsupported agent")
	}

	if !strings.Contains(err.Error(), "does not support agent") {
		t.Fatalf("Inject() error = %v", err)
	}
}

func TestInjectUsesRealEmbeddedContent(t *testing.T) {
	home := t.TempDir()

	_, err := Inject(home, model.AgentClaudeCode, []model.SkillID{model.SkillTypeScript})
	if err != nil {
		t.Fatalf("Inject() error = %v", err)
	}

	path := filepath.Join(home, ".claude", "skills", "typescript", "SKILL.md")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	// Real embedded content should be substantial (not a one-line stub).
	if len(content) < 100 {
		t.Fatalf("skill file content looks like a stub (len=%d)", len(content))
	}
}

func TestSkillPathForAgent(t *testing.T) {
	path, err := SkillPathForAgent("/home/test", model.AgentClaudeCode, model.SkillTypeScript)
	if err != nil {
		t.Fatalf("SkillPathForAgent() error = %v", err)
	}
	want := "/home/test/.claude/skills/typescript/SKILL.md"
	if path != want {
		t.Fatalf("SkillPathForAgent() = %q, want %q", path, want)
	}

	path, err = SkillPathForAgent("/home/test", model.AgentOpenCode, model.SkillReact19)
	if err != nil {
		t.Fatalf("SkillPathForAgent() error = %v", err)
	}
	want = "/home/test/.config/opencode/skill/react-19/SKILL.md"
	if path != want {
		t.Fatalf("SkillPathForAgent() = %q, want %q", path, want)
	}
}
