package engram

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

// defaultEngramServerJSON is the MCP server config for Claude Code.
// Placed at ~/.claude/mcp/engram.json — same pattern as context7.json.
var defaultEngramServerJSON = []byte("{\n  \"command\": \"engram\",\n  \"args\": []\n}\n")

// defaultEngramOverlayJSON is the settings.json overlay for OpenCode.
var defaultEngramOverlayJSON = []byte("{\n  \"mcpServers\": {\n    \"engram\": {\n      \"command\": \"engram\",\n      \"args\": []\n    }\n  }\n}\n")

func Inject(homeDir string, agent model.AgentID) (InjectionResult, error) {
	switch agent {
	case model.AgentClaudeCode:
		return injectClaude(homeDir)
	case model.AgentOpenCode:
		return injectOpenCode(homeDir)
	default:
		return InjectionResult{}, fmt.Errorf("engram injector does not support agent %q", agent)
	}
}

func injectClaude(homeDir string) (InjectionResult, error) {
	files := make([]string, 0, 2)
	changed := false

	// 1. Write MCP server config at ~/.claude/mcp/engram.json.
	mcpPath := filepath.Join(homeDir, ".claude", "mcp", "engram.json")
	mcpWrite, err := filemerge.WriteFileAtomic(mcpPath, defaultEngramServerJSON, 0o644)
	if err != nil {
		return InjectionResult{}, err
	}
	changed = changed || mcpWrite.Changed
	files = append(files, mcpPath)

	// 2. Inject Engram memory protocol into CLAUDE.md.
	claudeMDPath := filepath.Join(homeDir, ".claude", "CLAUDE.md")
	protocolContent := assets.MustRead("claude/engram-protocol.md")

	existing, err := readFileOrEmpty(claudeMDPath)
	if err != nil {
		return InjectionResult{}, err
	}

	updated := filemerge.InjectMarkdownSection(existing, "engram-protocol", protocolContent)

	mdWrite, err := filemerge.WriteFileAtomic(claudeMDPath, []byte(updated), 0o644)
	if err != nil {
		return InjectionResult{}, err
	}
	changed = changed || mdWrite.Changed
	files = append(files, claudeMDPath)

	return InjectionResult{Changed: changed, Files: files}, nil
}

func injectOpenCode(homeDir string) (InjectionResult, error) {
	// Merge engram into mcpServers in settings.json — same pattern as context7.
	settingsPath := filepath.Join(homeDir, ".config", "opencode", "settings.json")
	settingsWrite, err := mergeJSONFile(settingsPath, defaultEngramOverlayJSON)
	if err != nil {
		return InjectionResult{}, err
	}

	return InjectionResult{
		Changed: settingsWrite.Changed,
		Files:   []string{settingsPath},
	}, nil
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
