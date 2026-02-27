package sdd

import (
	"fmt"
	"path/filepath"

	"github.com/gentleman-programming/gentleman-ai-installer/internal/components/filemerge"
	"github.com/gentleman-programming/gentleman-ai-installer/internal/model"
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
	path := filepath.Join(homeDir, ".claude", "CLAUDE.md")
	content := []byte("# SDD Orchestrator\n\nUse the Spec-Driven Development flow:\n- /sdd-init\n- /sdd-new\n- /sdd-continue\n- /sdd-ff\n- /sdd-apply\n- /sdd-verify\n- /sdd-archive\n")
	writeResult, err := filemerge.WriteFileAtomic(path, content, 0o644)
	if err != nil {
		return InjectionResult{}, err
	}

	return InjectionResult{Changed: writeResult.Changed, Files: []string{path}}, nil
}

func injectOpenCode(homeDir string) (InjectionResult, error) {
	commandsDir := filepath.Join(homeDir, ".config", "opencode", "commands")
	files := make([]string, 0, len(OpenCodeCommands()))
	changed := false

	for _, command := range OpenCodeCommands() {
		path := filepath.Join(commandsDir, command.Name+".md")
		content := []byte("# " + command.Name + "\n\n" + command.Description + "\n\n" + command.Body + "\n")
		writeResult, err := filemerge.WriteFileAtomic(path, content, 0o644)
		if err != nil {
			return InjectionResult{}, err
		}

		changed = changed || writeResult.Changed
		files = append(files, path)
	}

	return InjectionResult{Changed: changed, Files: files}, nil
}
