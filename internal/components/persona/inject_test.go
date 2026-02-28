package persona

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gentleman-programming/gentle-ai/internal/agents"
	"github.com/gentleman-programming/gentle-ai/internal/agents/claude"
	"github.com/gentleman-programming/gentle-ai/internal/agents/opencode"
	"github.com/gentleman-programming/gentle-ai/internal/model"
)

func claudeAdapter() agents.Adapter   { return claude.NewAdapter() }
func opencodeAdapter() agents.Adapter { return opencode.NewAdapter() }

func TestInjectClaudeGentlemanWritesSectionWithRealContent(t *testing.T) {
	home := t.TempDir()

	result, err := Inject(home, claudeAdapter(), model.PersonaGentleman)
	if err != nil {
		t.Fatalf("Inject() error = %v", err)
	}
	if !result.Changed {
		t.Fatalf("Inject() changed = false")
	}

	path := filepath.Join(home, ".claude", "CLAUDE.md")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	text := string(content)
	if !strings.Contains(text, "<!-- gentle-ai:persona -->") {
		t.Fatal("CLAUDE.md missing open marker for persona")
	}
	if !strings.Contains(text, "<!-- /gentle-ai:persona -->") {
		t.Fatal("CLAUDE.md missing close marker for persona")
	}
	// Real content check — the embedded persona has these patterns.
	if !strings.Contains(text, "Senior Architect") {
		t.Fatal("CLAUDE.md missing real persona content (expected 'Senior Architect')")
	}
}

func TestInjectClaudeGentlemanWritesOutputStyleFile(t *testing.T) {
	home := t.TempDir()

	_, err := Inject(home, claudeAdapter(), model.PersonaGentleman)
	if err != nil {
		t.Fatalf("Inject() error = %v", err)
	}

	// Verify output-style file was written.
	stylePath := filepath.Join(home, ".claude", "output-styles", "gentleman.md")
	content, err := os.ReadFile(stylePath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", stylePath, err)
	}

	text := string(content)
	if !strings.Contains(text, "name: Gentleman") {
		t.Fatal("Output style file missing YAML frontmatter 'name: Gentleman'")
	}
	if !strings.Contains(text, "keep-coding-instructions: true") {
		t.Fatal("Output style file missing 'keep-coding-instructions: true'")
	}
	if !strings.Contains(text, "Gentleman Output Style") {
		t.Fatal("Output style file missing 'Gentleman Output Style' heading")
	}
}

func TestInjectClaudeGentlemanMergesOutputStyleIntoSettings(t *testing.T) {
	home := t.TempDir()

	// Pre-create a settings.json with some existing content.
	settingsDir := filepath.Join(home, ".claude")
	if err := os.MkdirAll(settingsDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	existingSettings := `{"permissions": {"allow": ["Read"]}, "syntaxHighlightingDisabled": true}`
	if err := os.WriteFile(filepath.Join(settingsDir, "settings.json"), []byte(existingSettings), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err := Inject(home, claudeAdapter(), model.PersonaGentleman)
	if err != nil {
		t.Fatalf("Inject() error = %v", err)
	}

	// Verify settings.json has outputStyle merged in.
	settingsPath := filepath.Join(home, ".claude", "settings.json")
	settingsContent, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", settingsPath, err)
	}

	var settings map[string]any
	if err := json.Unmarshal(settingsContent, &settings); err != nil {
		t.Fatalf("Unmarshal settings.json error = %v", err)
	}

	outputStyle, ok := settings["outputStyle"]
	if !ok {
		t.Fatal("settings.json missing 'outputStyle' key")
	}
	if outputStyle != "Gentleman" {
		t.Fatalf("settings.json outputStyle = %q, want %q", outputStyle, "Gentleman")
	}

	// Verify existing keys were preserved.
	if _, ok := settings["permissions"]; !ok {
		t.Fatal("settings.json lost 'permissions' key during merge")
	}
	if _, ok := settings["syntaxHighlightingDisabled"]; !ok {
		t.Fatal("settings.json lost 'syntaxHighlightingDisabled' key during merge")
	}
}

func TestInjectClaudeGentlemanReturnsAllFiles(t *testing.T) {
	home := t.TempDir()

	result, err := Inject(home, claudeAdapter(), model.PersonaGentleman)
	if err != nil {
		t.Fatalf("Inject() error = %v", err)
	}

	// Should return 3 files: CLAUDE.md, output-style, settings.json.
	if len(result.Files) != 3 {
		t.Fatalf("Inject() returned %d files, want 3: %v", len(result.Files), result.Files)
	}

	wantSuffixes := []string{"CLAUDE.md", "gentleman.md", "settings.json"}
	for _, suffix := range wantSuffixes {
		found := false
		for _, f := range result.Files {
			if strings.HasSuffix(f, suffix) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Inject() missing file with suffix %q in %v", suffix, result.Files)
		}
	}
}

func TestInjectClaudeNeutralWritesMinimalContent(t *testing.T) {
	home := t.TempDir()

	result, err := Inject(home, claudeAdapter(), model.PersonaNeutral)
	if err != nil {
		t.Fatalf("Inject() error = %v", err)
	}
	if !result.Changed {
		t.Fatalf("Inject() changed = false")
	}

	path := filepath.Join(home, ".claude", "CLAUDE.md")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	text := string(content)
	if !strings.Contains(text, "helpful") {
		t.Fatal("Neutral persona should contain 'helpful'")
	}
	// Should NOT have gentleman-specific content.
	if strings.Contains(text, "Senior Architect") {
		t.Fatal("Neutral persona should not contain gentleman content")
	}
}

func TestInjectClaudeNeutralDoesNotWriteOutputStyle(t *testing.T) {
	home := t.TempDir()

	result, err := Inject(home, claudeAdapter(), model.PersonaNeutral)
	if err != nil {
		t.Fatalf("Inject() error = %v", err)
	}

	// Should only return CLAUDE.md, no output-style file.
	if len(result.Files) != 1 {
		t.Fatalf("Neutral persona returned %d files, want 1: %v", len(result.Files), result.Files)
	}

	// Output-style file should NOT exist.
	stylePath := filepath.Join(home, ".claude", "output-styles", "gentleman.md")
	if _, err := os.Stat(stylePath); !os.IsNotExist(err) {
		t.Fatal("Neutral persona should NOT write output-style file")
	}
}

func TestInjectCustomClaudeDoesNothing(t *testing.T) {
	home := t.TempDir()

	result, err := Inject(home, claudeAdapter(), model.PersonaCustom)
	if err != nil {
		t.Fatalf("Inject() error = %v", err)
	}
	if result.Changed {
		t.Fatal("Custom persona should NOT change anything")
	}
	if len(result.Files) != 0 {
		t.Fatalf("Custom persona should return no files, got %v", result.Files)
	}

	// CLAUDE.md should NOT be created.
	claudeMD := filepath.Join(home, ".claude", "CLAUDE.md")
	if _, err := os.Stat(claudeMD); !os.IsNotExist(err) {
		t.Fatal("Custom persona should NOT create CLAUDE.md")
	}
}

func TestInjectCustomOpenCodeDoesNothing(t *testing.T) {
	home := t.TempDir()

	result, err := Inject(home, opencodeAdapter(), model.PersonaCustom)
	if err != nil {
		t.Fatalf("Inject() error = %v", err)
	}
	if result.Changed {
		t.Fatal("Custom persona (OpenCode) should NOT change anything")
	}
	if len(result.Files) != 0 {
		t.Fatalf("Custom persona (OpenCode) should return no files, got %v", result.Files)
	}

	// AGENTS.md should NOT be created.
	agentsMD := filepath.Join(home, ".config", "opencode", "AGENTS.md")
	if _, err := os.Stat(agentsMD); !os.IsNotExist(err) {
		t.Fatal("Custom persona (OpenCode) should NOT create AGENTS.md")
	}
}

func TestInjectOpenCodeGentlemanWritesAgentsFile(t *testing.T) {
	home := t.TempDir()

	result, err := Inject(home, opencodeAdapter(), model.PersonaGentleman)
	if err != nil {
		t.Fatalf("Inject() error = %v", err)
	}
	if !result.Changed {
		t.Fatalf("Inject() changed = false")
	}

	path := filepath.Join(home, ".config", "opencode", "AGENTS.md")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	text := string(content)
	if !strings.Contains(text, "Senior Architect") {
		t.Fatal("AGENTS.md missing real persona content")
	}
}

func TestInjectClaudeIsIdempotent(t *testing.T) {
	home := t.TempDir()

	first, err := Inject(home, claudeAdapter(), model.PersonaGentleman)
	if err != nil {
		t.Fatalf("Inject() first error = %v", err)
	}
	if !first.Changed {
		t.Fatalf("Inject() first changed = false")
	}

	second, err := Inject(home, claudeAdapter(), model.PersonaGentleman)
	if err != nil {
		t.Fatalf("Inject() second error = %v", err)
	}
	if second.Changed {
		t.Fatalf("Inject() second changed = true")
	}
}

func TestInjectOpenCodeIsIdempotent(t *testing.T) {
	home := t.TempDir()

	first, err := Inject(home, opencodeAdapter(), model.PersonaGentleman)
	if err != nil {
		t.Fatalf("Inject() first error = %v", err)
	}
	if !first.Changed {
		t.Fatalf("Inject() first changed = false")
	}

	second, err := Inject(home, opencodeAdapter(), model.PersonaGentleman)
	if err != nil {
		t.Fatalf("Inject() second error = %v", err)
	}
	if second.Changed {
		t.Fatalf("Inject() second changed = true")
	}
}

func TestInjectCursorGentlemanWritesRulesFileWithRealContent(t *testing.T) {
	home := t.TempDir()

	cursorAdapter, err := agents.NewAdapter("cursor")
	if err != nil {
		t.Fatalf("NewAdapter(cursor) error = %v", err)
	}

	result, injectErr := Inject(home, cursorAdapter, model.PersonaGentleman)
	if injectErr != nil {
		t.Fatalf("Inject(cursor) error = %v", injectErr)
	}

	if !result.Changed {
		t.Fatalf("Inject(cursor, gentleman) changed = false")
	}

	// Verify the generic persona content was used — not just neutral one-liner.
	path := filepath.Join(home, ".cursor", "rules", "gentle-ai.mdc")
	content, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, readErr)
	}

	text := string(content)
	if !strings.Contains(text, "Senior Architect") {
		t.Fatal("Cursor persona missing 'Senior Architect' — got neutral fallback instead of generic persona")
	}
	if !strings.Contains(text, "Skills") {
		t.Fatal("Cursor persona missing skills section")
	}
}

func TestInjectGeminiGentlemanWritesSystemPromptWithRealContent(t *testing.T) {
	home := t.TempDir()

	geminiAdapter, err := agents.NewAdapter("gemini-cli")
	if err != nil {
		t.Fatalf("NewAdapter(gemini-cli) error = %v", err)
	}

	result, injectErr := Inject(home, geminiAdapter, model.PersonaGentleman)
	if injectErr != nil {
		t.Fatalf("Inject(gemini) error = %v", injectErr)
	}

	if !result.Changed {
		t.Fatal("Inject(gemini, gentleman) changed = false")
	}

	path := filepath.Join(home, ".gemini", "GEMINI.md")
	content, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, readErr)
	}

	text := string(content)
	if !strings.Contains(text, "Senior Architect") {
		t.Fatal("Gemini persona missing 'Senior Architect'")
	}
}

func TestInjectVSCodeGentlemanWritesCopilotInstructions(t *testing.T) {
	home := t.TempDir()

	vscodeAdapter, err := agents.NewAdapter("vscode-copilot")
	if err != nil {
		t.Fatalf("NewAdapter(vscode-copilot) error = %v", err)
	}

	result, injectErr := Inject(home, vscodeAdapter, model.PersonaGentleman)
	if injectErr != nil {
		t.Fatalf("Inject(vscode) error = %v", injectErr)
	}

	if !result.Changed {
		t.Fatal("Inject(vscode, gentleman) changed = false")
	}

	path := filepath.Join(home, ".github", "copilot-instructions.md")
	content, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, readErr)
	}

	text := string(content)
	if !strings.Contains(text, "Senior Architect") {
		t.Fatal("VS Code persona missing 'Senior Architect'")
	}
}
