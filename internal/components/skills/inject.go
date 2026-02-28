package skills

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/gentleman-programming/gentle-ai/internal/assets"
	"github.com/gentleman-programming/gentle-ai/internal/components/filemerge"
	"github.com/gentleman-programming/gentle-ai/internal/model"
)

type InjectionResult struct {
	Changed bool
	Files   []string
	Skipped []model.SkillID
}

// Inject writes the embedded SKILL.md files for each requested skill
// to the correct directory for the given agent.
//
//   - Claude Code: ~/.claude/skills/{skillID}/SKILL.md
//   - OpenCode:    ~/.config/opencode/skill/{skillID}/SKILL.md
//
// Individual skill failures (e.g., missing embedded asset) are logged
// and skipped rather than aborting the entire operation.
func Inject(homeDir string, agent model.AgentID, skillIDs []model.SkillID) (InjectionResult, error) {
	skillDir, err := skillDirectoryForAgent(homeDir, agent)
	if err != nil {
		return InjectionResult{}, err
	}

	paths := make([]string, 0, len(skillIDs))
	skipped := make([]model.SkillID, 0)
	changed := false

	for _, id := range skillIDs {
		assetPath := "skills/" + string(id) + "/SKILL.md"
		content, readErr := assets.Read(assetPath)
		if readErr != nil {
			log.Printf("skills: skipping %q — embedded asset not found: %v", id, readErr)
			skipped = append(skipped, id)
			continue
		}

		path := filepath.Join(skillDir, string(id), "SKILL.md")
		writeResult, writeErr := filemerge.WriteFileAtomic(path, []byte(content), 0o644)
		if writeErr != nil {
			log.Printf("skills: skipping %q — write error: %v", id, writeErr)
			skipped = append(skipped, id)
			continue
		}

		changed = changed || writeResult.Changed
		paths = append(paths, path)
	}

	return InjectionResult{Changed: changed, Files: paths, Skipped: skipped}, nil
}

// SkillPathForAgent returns the filesystem path where a skill file would be written.
func SkillPathForAgent(homeDir string, agent model.AgentID, id model.SkillID) (string, error) {
	skillDir, err := skillDirectoryForAgent(homeDir, agent)
	if err != nil {
		return "", err
	}
	return filepath.Join(skillDir, string(id), "SKILL.md"), nil
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
