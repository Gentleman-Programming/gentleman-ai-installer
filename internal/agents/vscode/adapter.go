package vscode

import (
	"context"
	"os/exec"
	"path/filepath"

	"github.com/gentleman-programming/gentle-ai/internal/model"
	"github.com/gentleman-programming/gentle-ai/internal/system"
)

type Adapter struct {
	lookPath func(string) (string, error)
}

func NewAdapter() *Adapter {
	return &Adapter{
		lookPath: exec.LookPath,
	}
}

// --- Identity ---

func (a *Adapter) Agent() model.AgentID {
	return model.AgentVSCodeCopilot
}

func (a *Adapter) Tier() model.SupportTier {
	return model.TierFull
}

// --- Detection ---

func (a *Adapter) Detect(_ context.Context, _ string) (bool, string, string, bool, error) {
	// VS Code is detected by its binary on PATH.
	binaryPath, err := a.lookPath("code")
	if err != nil {
		return false, "", "", false, nil
	}

	return true, binaryPath, "", true, nil
}

// --- Installation ---

func (a *Adapter) SupportsAutoInstall() bool {
	return false // VS Code is a desktop app installed via package managers.
}

func (a *Adapter) InstallCommand(_ system.PlatformProfile) ([][]string, error) {
	return nil, AgentNotInstallableError{Agent: model.AgentVSCodeCopilot}
}

// --- Config paths ---
// VS Code ecosystem: Copilot reads ~/.github/copilot-instructions.md for system
// instructions. Skills and MCP config go under ~/.vscode/ so that any AI extension
// (Copilot, Claude extension, Codex, Roo, etc.) can benefit from them.

func (a *Adapter) GlobalConfigDir(homeDir string) string {
	return filepath.Join(homeDir, ".vscode")
}

func (a *Adapter) SystemPromptFile(homeDir string) string {
	// Global instructions path — VS Code Copilot reads ~/.github/copilot-instructions.md
	return filepath.Join(homeDir, ".github", "copilot-instructions.md")
}

func (a *Adapter) SkillsDir(homeDir string) string {
	// Skills under ~/.vscode/skills/ — available to any VS Code AI extension.
	return filepath.Join(homeDir, ".vscode", "skills")
}

func (a *Adapter) SettingsPath(homeDir string) string {
	// MCP settings for VS Code AI extensions.
	return filepath.Join(homeDir, ".vscode", "settings.json")
}

// --- Config strategies ---

func (a *Adapter) SystemPromptStrategy() model.SystemPromptStrategy {
	return model.StrategyFileReplace
}

func (a *Adapter) MCPStrategy() model.MCPStrategy {
	return model.StrategyMergeIntoSettings
}

// --- MCP ---

func (a *Adapter) MCPConfigPath(homeDir string, _ string) string {
	return filepath.Join(homeDir, ".vscode", "settings.json")
}

// --- Optional capabilities ---

func (a *Adapter) SupportsOutputStyles() bool {
	return false
}

func (a *Adapter) OutputStyleDir(_ string) string {
	return ""
}

func (a *Adapter) SupportsSlashCommands() bool {
	return false
}

func (a *Adapter) CommandsDir(_ string) string {
	return ""
}

func (a *Adapter) SupportsSkills() bool {
	return true
}

func (a *Adapter) SupportsSystemPrompt() bool {
	return true
}

func (a *Adapter) SupportsMCP() bool {
	return true
}

// AgentNotInstallableError is returned when InstallCommand is called on a desktop-only agent.
type AgentNotInstallableError struct {
	Agent model.AgentID
}

func (e AgentNotInstallableError) Error() string {
	return "agent " + string(e.Agent) + " is a desktop app and cannot be installed via CLI"
}
