package sdd

import (
	"fmt"
	"io/fs"
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

func Inject(homeDir string, agent model.AgentID) (InjectionResult, error) {
	switch agent {
	case model.AgentClaudeCode:
		return injectClaude(homeDir)
	case model.AgentOpenCode:
		return injectOpenCode(homeDir)
	default:
		return InjectionResult{}, fmt.Errorf("sdd injector does not support agent %q", agent)
	}
}

func injectClaude(homeDir string) (InjectionResult, error) {
	claudeMDPath := filepath.Join(homeDir, ".claude", "CLAUDE.md")

	content := assets.MustRead("claude/sdd-orchestrator.md")

	existing, err := readFileOrEmpty(claudeMDPath)
	if err != nil {
		return InjectionResult{}, err
	}

	updated := filemerge.InjectMarkdownSection(existing, "sdd-orchestrator", content)

	writeResult, err := filemerge.WriteFileAtomic(claudeMDPath, []byte(updated), 0o644)
	if err != nil {
		return InjectionResult{}, err
	}

	return InjectionResult{Changed: writeResult.Changed, Files: []string{claudeMDPath}}, nil
}

func injectOpenCode(homeDir string) (InjectionResult, error) {
	commandsDir := filepath.Join(homeDir, ".config", "opencode", "commands")
	skillDir := filepath.Join(homeDir, ".config", "opencode", "skill")

	files := make([]string, 0)
	changed := false

	// Write real command files from embedded assets.
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

	// Write SDD skill files from embedded assets.
	sddSkills := []string{
		"sdd-init", "sdd-explore", "sdd-propose", "sdd-spec",
		"sdd-design", "sdd-tasks", "sdd-apply", "sdd-verify", "sdd-archive",
	}

	for _, skill := range sddSkills {
		assetPath := "skills/" + skill + "/SKILL.md"
		content, readErr := assets.Read(assetPath)
		if readErr != nil {
			// Skip skills that don't have embedded content.
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

	return InjectionResult{Changed: changed, Files: files}, nil
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
