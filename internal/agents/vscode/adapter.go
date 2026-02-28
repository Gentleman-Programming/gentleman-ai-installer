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
	return model.TierPartial
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
// VS Code Copilot uses project-level .github/ directory for system instructions.

func (a *Adapter) GlobalConfigDir(_ string) string {
	// VS Code doesn't have a global agent config dir in the same sense.
	// Copilot instructions live at project level (.github/).
	return ""
}

func (a *Adapter) SystemPromptFile(homeDir string) string {
	// Global instructions path — VS Code supports ~/.github/copilot-instructions.md
	return filepath.Join(homeDir, ".github", "copilot-instructions.md")
}

func (a *Adapter) SkillsDir(_ string) string {
	// VS Code Copilot doesn't support standalone skill files.
	return ""
}

func (a *Adapter) SettingsPath(_ string) string {
	// VS Code settings.json lives in platform-specific app data dirs.
	// We don't modify it for MCP — Copilot uses its own extension settings.
	return ""
}

// --- Config strategies ---

func (a *Adapter) SystemPromptStrategy() model.SystemPromptStrategy {
	return model.StrategyFileReplace
}

func (a *Adapter) MCPStrategy() model.MCPStrategy {
	// VS Code Copilot MCP is configured via VS Code settings which we don't manage.
	return model.StrategyMergeIntoSettings
}

// --- MCP ---

func (a *Adapter) MCPConfigPath(_ string, _ string) string {
	return ""
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
	return false
}

func (a *Adapter) SupportsSystemPrompt() bool {
	return true
}

func (a *Adapter) SupportsMCP() bool {
	return false // VS Code Copilot MCP config is managed through VS Code UI, not files.
}

// AgentNotInstallableError is returned when InstallCommand is called on a desktop-only agent.
type AgentNotInstallableError struct {
	Agent model.AgentID
}

func (e AgentNotInstallableError) Error() string {
	return "agent " + string(e.Agent) + " is a desktop app and cannot be installed via CLI"
}
