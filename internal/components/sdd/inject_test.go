package sdd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gentleman-programming/gentle-ai/internal/model"
)

func TestInjectOpenCodeWritesCommandFiles(t *testing.T) {
	home := t.TempDir()

	result, err := Inject(home, model.AgentOpenCode)
	if err != nil {
		t.Fatalf("Inject() error = %v", err)
	}
	if !result.Changed {
		t.Fatalf("Inject() first changed = false")
	}

	if len(result.Files) != len(OpenCodeCommands()) {
		t.Fatalf("Inject() files = %d", len(result.Files))
	}

	path := filepath.Join(home, ".config", "opencode", "commands", "sdd-init.md")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected command file %q: %v", path, err)
	}

	second, err := Inject(home, model.AgentOpenCode)
	if err != nil {
		t.Fatalf("Inject() second error = %v", err)
	}
	if second.Changed {
		t.Fatalf("Inject() second changed = true")
	}
}
