package skills

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gentleman-programming/gentleman-ai-installer/internal/model"
)

func TestInjectWritesSkillFilesForOpenCode(t *testing.T) {
	home := t.TempDir()

	result, err := Inject(home, model.AgentOpenCode, []SkillFile{{Name: "typescript", Content: []byte("# TS")}})
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

	second, err := Inject(home, model.AgentOpenCode, []SkillFile{{Name: "typescript", Content: []byte("# TS")}})
	if err != nil {
		t.Fatalf("Inject() second error = %v", err)
	}

	if second.Changed {
		t.Fatalf("Inject() second changed = true")
	}
}
