package skills

import (
	"fmt"
	"path/filepath"

	"github.com/gentleman-programming/gentleman-ai-installer/internal/components/filemerge"
	"github.com/gentleman-programming/gentleman-ai-installer/internal/model"
)

type SkillFile struct {
	Name    string
	Content []byte
}

type InjectionResult struct {
	Changed bool
	Files   []string
}

func Inject(homeDir string, agent model.AgentID, files []SkillFile) (InjectionResult, error) {
	skillDir, err := skillDirectoryForAgent(homeDir, agent)
	if err != nil {
		return InjectionResult{}, err
	}

	paths := make([]string, 0, len(files))
	changed := false
	for _, file := range files {
		path := filepath.Join(skillDir, file.Name, "SKILL.md")
		writeResult, err := filemerge.WriteFileAtomic(path, file.Content, 0o644)
		if err != nil {
			return InjectionResult{}, err
		}

		changed = changed || writeResult.Changed
		paths = append(paths, path)
	}

	return InjectionResult{Changed: changed, Files: paths}, nil
}

func skillDirectoryForAgent(homeDir string, agent model.AgentID) (string, error) {
	switch agent {
	case model.AgentClaudeCode:
		return filepath.Join(homeDir, ".claude", "skills"), nil
	case model.AgentOpenCode:
		return filepath.Join(homeDir, ".config", "opencode", "skill"), nil
	default:
		return "", fmt.Errorf("skills injector does not support agent %q", agent)
	}
}
