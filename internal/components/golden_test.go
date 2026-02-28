package components_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/gentleman-programming/gentle-ai/internal/components/mcp"
	"github.com/gentleman-programming/gentle-ai/internal/components/sdd"
	"github.com/gentleman-programming/gentle-ai/internal/components/skills"
)

func TestGoldenConfigs(t *testing.T) {
	presetsJSON, err := json.MarshalIndent(skills.Presets(), "", "  ")
	if err != nil {
		t.Fatalf("MarshalIndent() error = %v", err)
	}
	presetsJSON = append(presetsJSON, '\n')

	commands := sdd.OpenCodeCommands()
	if len(commands) == 0 {
		t.Fatalf("OpenCodeCommands() returned no commands")
	}
	commandMarkdown := []byte("# " + commands[0].Name + "\n\n" + commands[0].Description + "\n\n" + commands[0].Body + "\n")

	tests := []struct {
		name    string
		path    string
		content []byte
	}{
		{name: "context7 server", path: "context7-server.json", content: mcp.DefaultContext7ServerJSON()},
		{name: "context7 overlay", path: "context7-overlay.json", content: mcp.DefaultContext7OverlayJSON()},
		{name: "skills presets", path: "skills-presets.json", content: presetsJSON},
		{name: "sdd command markdown", path: "sdd-command-sdd-init.md", content: commandMarkdown},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			goldenPath := filepath.Join(goldenDir(t), tc.path)
			want, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("ReadFile(%q) error = %v", goldenPath, err)
			}

			if string(tc.content) != string(want) {
				t.Fatalf("golden mismatch for %s\nwant:\n%s\n\ngot:\n%s", tc.path, want, tc.content)
			}
		})
	}
}

func goldenDir(t *testing.T) string {
	t.Helper()
	return filepath.Join("..", "..", "testdata", "golden")
}
