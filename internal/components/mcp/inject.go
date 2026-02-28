package mcp

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

func Inject(homeDir string, agent model.AgentID) (InjectionResult, error) {
	switch agent {
	case model.AgentClaudeCode:
		return injectClaude(homeDir)
	case model.AgentOpenCode:
		return injectOpenCode(homeDir)
	default:
		return InjectionResult{}, fmt.Errorf("mcp injector does not support agent %q", agent)
	}
}

func injectClaude(homeDir string) (InjectionResult, error) {
	path := filepath.Join(homeDir, ".claude", "mcp", "context7.json")
	writeResult, err := filemerge.WriteFileAtomic(path, DefaultContext7ServerJSON(), 0o644)
	if err != nil {
		return InjectionResult{}, err
	}

	return InjectionResult{Changed: writeResult.Changed, Files: []string{path}}, nil
}

func injectOpenCode(homeDir string) (InjectionResult, error) {
	settingsPath := filepath.Join(homeDir, ".config", "opencode", "settings.json")
	settingsWrite, err := mergeJSONFile(settingsPath, DefaultContext7OverlayJSON())
	if err != nil {
		return InjectionResult{}, err
	}

	return InjectionResult{Changed: settingsWrite.Changed, Files: []string{settingsPath}}, nil
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
