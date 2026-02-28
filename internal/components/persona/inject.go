package persona

import (
	"fmt"
	"os"
	"path/filepath"

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

func Inject(homeDir string, agent model.AgentID, persona model.PersonaID) (InjectionResult, error) {
	switch agent {
	case model.AgentClaudeCode:
		return injectClaude(homeDir, persona)
	case model.AgentOpenCode:
		return injectOpenCode(homeDir, persona)
	default:
		return InjectionResult{}, fmt.Errorf("persona injector does not support agent %q", agent)
	}
}

func personaContent(agent model.AgentID, persona model.PersonaID) string {
	switch persona {
	case model.PersonaNeutral:
		return neutralPersonaContent
	case model.PersonaCustom:
		// Custom persona does nothing — the user keeps their own personality.
		// The SDD orchestrator is injected separately by the SDD component.
		return ""
	default:
		// Gentleman persona — read from embedded assets.
		switch agent {
		case model.AgentClaudeCode:
			return assets.MustRead("claude/persona-gentleman.md")
		case model.AgentOpenCode:
			return assets.MustRead("opencode/persona-gentleman.md")
		default:
			return neutralPersonaContent
		}
	}
}

func injectClaude(homeDir string, persona model.PersonaID) (InjectionResult, error) {
	// Custom persona does nothing — user keeps their own CLAUDE.md content.
	if persona == model.PersonaCustom {
		return InjectionResult{}, nil
	}

	files := make([]string, 0, 3)
	changed := false

	// 1. Inject persona content into CLAUDE.md section.
	claudeMDPath := filepath.Join(homeDir, ".claude", "CLAUDE.md")
	content := personaContent(model.AgentClaudeCode, persona)

	existing, err := readFileOrEmpty(claudeMDPath)
	if err != nil {
		return InjectionResult{}, err
	}

	updated := filemerge.InjectMarkdownSection(existing, "persona", content)

	writeResult, err := filemerge.WriteFileAtomic(claudeMDPath, []byte(updated), 0o644)
	if err != nil {
		return InjectionResult{}, err
	}
	changed = changed || writeResult.Changed
	files = append(files, claudeMDPath)

	// 2. Gentleman-only: write output-style file and merge outputStyle into settings.json.
	if persona == model.PersonaGentleman {
		outputStylePath := filepath.Join(homeDir, ".claude", "output-styles", "gentleman.md")
		outputStyleContent := assets.MustRead("claude/output-style-gentleman.md")

		styleResult, err := filemerge.WriteFileAtomic(outputStylePath, []byte(outputStyleContent), 0o644)
		if err != nil {
			return InjectionResult{}, err
		}
		changed = changed || styleResult.Changed
		files = append(files, outputStylePath)

		// Merge "outputStyle": "Gentleman" into ~/.claude/settings.json.
		settingsPath := filepath.Join(homeDir, ".claude", "settings.json")
		settingsResult, err := mergeJSONFile(settingsPath, outputStyleOverlayJSON)
		if err != nil {
			return InjectionResult{}, err
		}
		changed = changed || settingsResult.Changed
		files = append(files, settingsPath)
	}

	return InjectionResult{Changed: changed, Files: files}, nil
}

func injectOpenCode(homeDir string, persona model.PersonaID) (InjectionResult, error) {
	// Custom persona does nothing — user keeps their own AGENTS.md content.
	if persona == model.PersonaCustom {
		return InjectionResult{}, nil
	}

	content := personaContent(model.AgentOpenCode, persona)
	agentsPath := filepath.Join(homeDir, ".config", "opencode", "AGENTS.md")

	writeResult, err := filemerge.WriteFileAtomic(agentsPath, []byte(content), 0o644)
	if err != nil {
		return InjectionResult{}, err
	}

	return InjectionResult{Changed: writeResult.Changed, Files: []string{agentsPath}}, nil
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
