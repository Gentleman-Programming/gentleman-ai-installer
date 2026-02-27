package permissions

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/gentleman-programming/gentleman-ai-installer/internal/model"
)

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

	path := filepath.Join(home, ".config", "opencode", "settings.json")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected settings file %q: %v", path, err)
	}
}

func TestInjectAddsEnvToDenyList(t *testing.T) {
	home := t.TempDir()

	if _, err := Inject(home, model.AgentClaudeCode); err != nil {
		t.Fatalf("Inject() error = %v", err)
	}

	settingsPath := filepath.Join(home, ".claude", "settings.json")
	content, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("read settings file %q: %v", settingsPath, err)
	}

	var settings map[string]any
	if err := json.Unmarshal(content, &settings); err != nil {
		t.Fatalf("unmarshal settings json: %v", err)
	}

	permissionsNode, ok := settings["permissions"].(map[string]any)
	if !ok {
		t.Fatalf("permissions node missing or invalid: %#v", settings["permissions"])
	}

	denyList, ok := permissionsNode["deny"].([]any)
	if !ok {
		t.Fatalf("deny list missing or invalid: %#v", permissionsNode["deny"])
	}

	for _, entry := range denyList {
		if value, ok := entry.(string); ok && value == ".env" {
			return
		}
	}

	t.Fatalf("deny list missing explicit .env rule: %#v", denyList)
}
