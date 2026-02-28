package sdd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/gentleman-programming/gentle-ai/internal/agents"
	"github.com/gentleman-programming/gentle-ai/internal/assets"
	"github.com/gentleman-programming/gentle-ai/internal/components/filemerge"
	"github.com/gentleman-programming/gentle-ai/internal/model"
)

type InjectionResult struct {
	Changed bool
	Files   []string
}

func Inject(homeDir string, adapter agents.Adapter) (InjectionResult, error) {
	if !adapter.SupportsSystemPrompt() {
		return InjectionResult{}, nil
	}

	files := make([]string, 0)
	changed := false

	// 1. Inject SDD orchestrator into system prompt.
	switch adapter.SystemPromptStrategy() {
	case model.StrategyMarkdownSections:
		result, err := injectMarkdownSections(homeDir, adapter)
		if err != nil {
			return InjectionResult{}, err
		}
		changed = changed || result.Changed
		files = append(files, result.Files...)

	case model.StrategyFileReplace, model.StrategyAppendToFile:
		// For FileReplace/AppendToFile agents, the SDD orchestrator is included
		// in the generic persona asset. However, if the user chose neutral or
		// custom persona, the SDD content must still be injected. We append the
		// SDD orchestrator section to the existing system prompt file so it is
		// always present regardless of persona choice.
		result, err := injectFileAppend(homeDir, adapter)
		if err != nil {
			return InjectionResult{}, err
		}
		changed = changed || result.Changed
		files = append(files, result.Files...)
	}

	// 2. Write slash commands (if the agent supports them).
	if adapter.SupportsSlashCommands() {
		commandsDir := adapter.CommandsDir(homeDir)
		if commandsDir != "" {
			commandEntries, err := fs.ReadDir(assets.FS, "opencode/commands")
			if err != nil {
				return InjectionResult{}, fmt.Errorf("read embedded opencode/commands: %w", err)
			}

			for _, entry := range commandEntries {
				if entry.IsDir() {
					continue
				}

				content := assets.MustRead("opencode/commands/" + entry.Name())
				path := filepath.Join(commandsDir, entry.Name())
				writeResult, err := filemerge.WriteFileAtomic(path, []byte(content), 0o644)
				if err != nil {
					return InjectionResult{}, err
				}

				changed = changed || writeResult.Changed
				files = append(files, path)
			}
		}
	}

	// 3. Write SDD skill files (if the agent supports skills).
	if adapter.SupportsSkills() {
		skillDir := adapter.SkillsDir(homeDir)
		if skillDir != "" {
			sddSkills := []string{
				"sdd-init", "sdd-explore", "sdd-propose", "sdd-spec",
				"sdd-design", "sdd-tasks", "sdd-apply", "sdd-verify", "sdd-archive",
			}

			for _, skill := range sddSkills {
				assetPath := "skills/" + skill + "/SKILL.md"
				content, readErr := assets.Read(assetPath)
				if readErr != nil {
					continue
				}

				path := filepath.Join(skillDir, skill, "SKILL.md")
				writeResult, err := filemerge.WriteFileAtomic(path, []byte(content), 0o644)
				if err != nil {
					return InjectionResult{}, err
				}

				changed = changed || writeResult.Changed
				files = append(files, path)
			}
		}
	}

	return InjectionResult{Changed: changed, Files: files}, nil
}

// sddOrchestratorMarker is used to detect if SDD content was already injected
// (e.g., via the persona file or a previous SDD injection).
const sddOrchestratorMarker = "## Spec-Driven Development (SDD) Orchestrator"

func injectFileAppend(homeDir string, adapter agents.Adapter) (InjectionResult, error) {
	promptPath := adapter.SystemPromptFile(homeDir)

	existing, err := readFileOrEmpty(promptPath)
	if err != nil {
		return InjectionResult{}, err
	}

	// If the SDD orchestrator section is already present (e.g., from the
	// gentleman persona asset which includes it), skip to avoid duplication.
	if strings.Contains(existing, sddOrchestratorMarker) {
		return InjectionResult{Files: []string{promptPath}}, nil
	}

	// Use generic SDD orchestrator content suitable for any agent.
	content := assets.MustRead("generic/sdd-orchestrator.md")

	updated := existing
	if len(updated) > 0 && !strings.HasSuffix(updated, "\n") {
		updated += "\n"
	}
	if len(updated) > 0 {
		updated += "\n"
	}
	updated += content

	writeResult, err := filemerge.WriteFileAtomic(promptPath, []byte(updated), 0o644)
	if err != nil {
		return InjectionResult{}, err
	}

	return InjectionResult{Changed: writeResult.Changed, Files: []string{promptPath}}, nil
}

func injectMarkdownSections(homeDir string, adapter agents.Adapter) (InjectionResult, error) {
	promptPath := adapter.SystemPromptFile(homeDir)
	content := assets.MustRead("claude/sdd-orchestrator.md")

	existing, err := readFileOrEmpty(promptPath)
	if err != nil {
		return InjectionResult{}, err
	}

	updated := filemerge.InjectMarkdownSection(existing, "sdd-orchestrator", content)

	writeResult, err := filemerge.WriteFileAtomic(promptPath, []byte(updated), 0o644)
	if err != nil {
		return InjectionResult{}, err
	}

	return InjectionResult{Changed: writeResult.Changed, Files: []string{promptPath}}, nil
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
