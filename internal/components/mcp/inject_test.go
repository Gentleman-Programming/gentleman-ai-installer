package mcp

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gentleman-programming/gentleman-ai-installer/internal/model"
)

func TestInjectOpenCodeMergesContext7AndIsIdempotent(t *testing.T) {
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

	settingsPath := filepath.Join(home, ".config", "opencode", "settings.json")
	settings, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("ReadFile(settings.json) error = %v", err)
	}

	if len(settings) == 0 {
		t.Fatalf("settings.json is empty")
	}
}

func TestInjectClaudeWritesContext7FileAndIsIdempotent(t *testing.T) {
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

	path := filepath.Join(home, ".claude", "mcp", "context7.json")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected context7 file %q: %v", path, err)
	}
}
