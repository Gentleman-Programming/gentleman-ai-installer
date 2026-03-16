package cli

import (
	"path/filepath"
	"testing"

	"github.com/gentleman-programming/gentle-ai/internal/model"
)

func TestComponentPathsSDDIncludesSystemPromptForAllSupportedAgents(t *testing.T) {
	home := t.TempDir()
	adapters := resolveAdapters([]model.AgentID{
		model.AgentClaudeCode,
		model.AgentOpenCode,
		model.AgentGeminiCLI,
		model.AgentCursor,
		model.AgentVSCodeCopilot,
	})

	paths := componentPaths(home, model.Selection{}, adapters, model.ComponentSDD)

	for _, adapter := range adapters {
		p := adapter.SystemPromptFile(home)
		if !containsPath(paths, p) {
			t.Fatalf("componentPaths(sdd) missing system prompt path %q\npaths=%v", p, paths)
		}
	}
}

func TestComponentPathsSDDIncludesOpenCodeSettingsAndCommands(t *testing.T) {
	home := t.TempDir()
	adapters := resolveAdapters([]model.AgentID{model.AgentOpenCode})

	paths := componentPaths(home, model.Selection{}, adapters, model.ComponentSDD)

	settings := filepath.Join(home, ".config", "opencode", "opencode.json")
	if !containsPath(paths, settings) {
		t.Fatalf("componentPaths(sdd) missing OpenCode settings path %q\npaths=%v", settings, paths)
	}

	command := filepath.Join(home, ".config", "opencode", "commands", "sdd-init.md")
	if !containsPath(paths, command) {
		t.Fatalf("componentPaths(sdd) missing OpenCode command path %q\npaths=%v", command, paths)
	}
}

func TestComponentPathsSDDMultiIncludesOpenCodePlugin(t *testing.T) {
	home := t.TempDir()
	adapters := resolveAdapters([]model.AgentID{model.AgentOpenCode})

	paths := componentPaths(home, model.Selection{SDDMode: model.SDDModeMulti}, adapters, model.ComponentSDD)

	plugin := filepath.Join(home, ".config", "opencode", "plugins", "background-agents.ts")
	if !containsPath(paths, plugin) {
		t.Fatalf("componentPaths(sdd multi) missing OpenCode plugin path %q\npaths=%v", plugin, paths)
	}
}

func TestComponentPathsSDDSingleExcludesOpenCodePlugin(t *testing.T) {
	home := t.TempDir()
	adapters := resolveAdapters([]model.AgentID{model.AgentOpenCode})

	paths := componentPaths(home, model.Selection{SDDMode: model.SDDModeSingle}, adapters, model.ComponentSDD)

	plugin := filepath.Join(home, ".config", "opencode", "plugins", "background-agents.ts")
	if containsPath(paths, plugin) {
		t.Fatalf("componentPaths(sdd single) unexpectedly included OpenCode plugin path %q\npaths=%v", plugin, paths)
	}
}

func TestComponentPathsSDDIncludesSkillsAndSharedConventions(t *testing.T) {
	home := t.TempDir()
	adapters := resolveAdapters([]model.AgentID{model.AgentGeminiCLI})

	paths := componentPaths(home, model.Selection{}, adapters, model.ComponentSDD)

	// Verify all four shared convention files are reported.
	for _, sharedFile := range []string{
		"persistence-contract.md",
		"engram-convention.md",
		"openspec-convention.md",
		"sdd-phase-common.md",
	} {
		shared := filepath.Join(home, ".gemini", "skills", "_shared", sharedFile)
		if !containsPath(paths, shared) {
			t.Fatalf("componentPaths(sdd) missing shared convention path %q\npaths=%v", shared, paths)
		}
	}

	skill := filepath.Join(home, ".gemini", "skills", "sdd-verify", "SKILL.md")
	if !containsPath(paths, skill) {
		t.Fatalf("componentPaths(sdd) missing SDD skill path %q\npaths=%v", skill, paths)
	}
}

func containsPath(paths []string, want string) bool {
	for _, p := range paths {
		if p == want {
			return true
		}
	}
	return false
}
