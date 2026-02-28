package engram

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gentleman-programming/gentle-ai/internal/model"
)

func TestInjectOpenCodeCreatesPluginAndSettings(t *testing.T) {
	home := t.TempDir()

	result, err := Inject(home, model.AgentOpenCode)
	if err != nil {
		t.Fatalf("Inject() error = %v", err)
	}
	if !result.Changed {
		t.Fatalf("Inject() first changed = false")
	}

	if len(result.Files) != 2 {
		t.Fatalf("Inject() files = %v", result.Files)
	}

	pluginPath := filepath.Join(home, ".config", "opencode", "plugins", "engram.ts")
	if _, err := os.Stat(pluginPath); err != nil {
		t.Fatalf("expected plugin file: %v", err)
	}

	settingsPath := filepath.Join(home, ".config", "opencode", "settings.json")
	settings, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("ReadFile(settings.json) error = %v", err)
	}

	if string(settings) == "" {
		t.Fatalf("settings.json is empty")
	}

	second, err := Inject(home, model.AgentOpenCode)
	if err != nil {
		t.Fatalf("Inject() second error = %v", err)
	}
	if second.Changed {
		t.Fatalf("Inject() second changed = true")
	}
}
