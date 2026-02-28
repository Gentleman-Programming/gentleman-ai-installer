package theme

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gentleman-programming/gentle-ai/internal/components/filemerge"
	"github.com/gentleman-programming/gentle-ai/internal/model"
)

type InjectionResult struct {
	Changed bool
	Files   []string
}

var themeOverlayJSON = []byte("{\n  \"theme\": \"gentleman-kanagawa\"\n}\n")

func Inject(homeDir string, agent model.AgentID) (InjectionResult, error) {
	settingsPath, err := settingsPathForAgent(homeDir, agent)
	if err != nil {
		return InjectionResult{}, err
	}

	writeResult, err := mergeJSONFile(settingsPath, themeOverlayJSON)
	if err != nil {
		return InjectionResult{}, err
	}

	return InjectionResult{Changed: writeResult.Changed, Files: []string{settingsPath}}, nil
}

func settingsPathForAgent(homeDir string, agent model.AgentID) (string, error) {
	switch agent {
	case model.AgentClaudeCode:
		return filepath.Join(homeDir, ".claude", "settings.json"), nil
	case model.AgentOpenCode:
		return filepath.Join(homeDir, ".config", "opencode", "settings.json"), nil
	default:
		return "", fmt.Errorf("theme injector does not support agent %q", agent)
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
