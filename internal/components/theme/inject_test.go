package theme

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gentleman-programming/gentleman-ai-installer/internal/model"
)

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

	path := filepath.Join(home, ".claude", "settings.json")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected settings file %q: %v", path, err)
	}
}
