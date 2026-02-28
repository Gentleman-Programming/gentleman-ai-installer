package persona

import (
	"fmt"
	"os"

	"github.com/gentleman-programming/gentle-ai/internal/agents"
	"github.com/gentleman-programming/gentle-ai/internal/assets"
	"github.com/gentleman-programming/gentle-ai/internal/components/filemerge"
	"github.com/gentleman-programming/gentle-ai/internal/model"
)

type InjectionResult struct {
	Changed bool
	Files   []string
}

const neutralPersonaContent = "Be helpful, direct, and technically precise. Focus on accuracy and clarity.\n"

// outputStyleOverlayJSON is the settings.json overlay to enable the Gentleman output style.
var outputStyleOverlayJSON = []byte("{\n  \"outputStyle\": \"Gentleman\"\n}\n")

// openCodeAgentOverlayJSON defines Tab-switchable agents for OpenCode.
// "gentleman" is the primary agent, "sdd-orchestrator" is available via Tab.
// Both reference AGENTS.md via {file:./AGENTS.md} for their system prompt.
var openCodeAgentOverlayJSON = []byte("{\n  \"agent\": {\n    \"gentleman\": {\n      \"mode\": \"primary\",\n      \"description\": \"Senior Architect mentor - helpful first, challenging when it matters\",\n      \"prompt\": \"{file:./AGENTS.md}\",\n      \"tools\": {\n        \"write\": true,\n        \"edit\": true\n      }\n    },\n    \"sdd-orchestrator\": {\n      \"mode\": \"all\",\n      \"description\": \"Gentleman personality + SDD delegate-only orchestrator\",\n      \"prompt\": \"{file:./AGENTS.md}\",\n      \"tools\": {\n        \"read\": true,\n        \"write\": true,\n        \"edit\": true,\n        \"bash\": true\n      }\n    }\n  }\n}\n")

func Inject(homeDir string, adapter agents.Adapter, persona model.PersonaID) (InjectionResult, error) {
	if !adapter.SupportsSystemPrompt() {
		return InjectionResult{}, nil
	}

	// Custom persona does nothing — user keeps their own config.
	if persona == model.PersonaCustom {
		return InjectionResult{}, nil
	}

	files := make([]string, 0, 3)
	changed := false

	content := personaContent(adapter.Agent(), persona)
	if content == "" {
		return InjectionResult{}, nil
	}

	// 1. Inject persona content based on system prompt strategy.
	switch adapter.SystemPromptStrategy() {
	case model.StrategyMarkdownSections:
		promptPath := adapter.SystemPromptFile(homeDir)
		existing, err := readFileOrEmpty(promptPath)
		if err != nil {
			return InjectionResult{}, err
		}

		updated := filemerge.InjectMarkdownSection(existing, "persona", content)

		writeResult, err := filemerge.WriteFileAtomic(promptPath, []byte(updated), 0o644)
		if err != nil {
			return InjectionResult{}, err
		}
		changed = changed || writeResult.Changed
		files = append(files, promptPath)

	case model.StrategyFileReplace:
		promptPath := adapter.SystemPromptFile(homeDir)
		writeResult, err := filemerge.WriteFileAtomic(promptPath, []byte(content), 0o644)
		if err != nil {
			return InjectionResult{}, err
		}
		changed = changed || writeResult.Changed
		files = append(files, promptPath)

	case model.StrategyAppendToFile:
		promptPath := adapter.SystemPromptFile(homeDir)
		writeResult, err := filemerge.WriteFileAtomic(promptPath, []byte(content), 0o644)
		if err != nil {
			return InjectionResult{}, err
		}
		changed = changed || writeResult.Changed
		files = append(files, promptPath)
	}

	// 2. OpenCode agent definitions — Tab-switchable agents in opencode.json.
	if adapter.Agent() == model.AgentOpenCode && persona != model.PersonaCustom {
		settingsPath := adapter.SettingsPath(homeDir)
		if settingsPath != "" {
			agentResult, err := mergeJSONFile(settingsPath, openCodeAgentOverlayJSON)
			if err != nil {
				return InjectionResult{}, err
			}
			changed = changed || agentResult.Changed
			files = append(files, settingsPath)
		}
	}

	// 3. Gentleman-only: write output style + merge into settings (if agent supports it).
	if persona == model.PersonaGentleman && adapter.SupportsOutputStyles() {
		outputStyleDir := adapter.OutputStyleDir(homeDir)
		if outputStyleDir != "" {
			outputStylePath := outputStyleDir + "/gentleman.md"
			outputStyleContent := assets.MustRead("claude/output-style-gentleman.md")

			styleResult, err := filemerge.WriteFileAtomic(outputStylePath, []byte(outputStyleContent), 0o644)
			if err != nil {
				return InjectionResult{}, err
			}
			changed = changed || styleResult.Changed
			files = append(files, outputStylePath)
		}

		// Merge "outputStyle": "Gentleman" into settings.
		settingsPath := adapter.SettingsPath(homeDir)
		if settingsPath != "" {
			settingsResult, err := mergeJSONFile(settingsPath, outputStyleOverlayJSON)
			if err != nil {
				return InjectionResult{}, err
			}
			changed = changed || settingsResult.Changed
			files = append(files, settingsPath)
		}
	}

	return InjectionResult{Changed: changed, Files: files}, nil
}

func personaContent(agent model.AgentID, persona model.PersonaID) string {
	switch persona {
	case model.PersonaNeutral:
		return neutralPersonaContent
	case model.PersonaCustom:
		return ""
	default:
		// Gentleman persona — try agent-specific asset, then generic fallback.
		switch agent {
		case model.AgentClaudeCode:
			return assets.MustRead("claude/persona-gentleman.md")
		case model.AgentOpenCode:
			return assets.MustRead("opencode/persona-gentleman.md")
		default:
			// Generic persona includes Gentleman personality + skills table + SDD orchestrator.
			// Used by Gemini CLI, Cursor, VS Code Copilot, and any future agents.
			return assets.MustRead("generic/persona-gentleman.md")
		}
	}
}

func mergeJSONFile(path string, overlay []byte) (filemerge.WriteResult, error) {
	baseJSON, err := osReadFile(path)
	if err != nil {
		return filemerge.WriteResult{}, err
	}

	merged, err := filemerge.MergeJSONObjects(baseJSON, overlay)
	if err != nil {
		return filemerge.WriteResult{}, err
	}

	return filemerge.WriteFileAtomic(path, merged, 0o644)
}

var osReadFile = func(path string) ([]byte, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read json file %q: %w", path, err)
	}

	return content, nil
}

func readFileOrEmpty(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("read file %q: %w", path, err)
	}
	return string(data), nil
}
