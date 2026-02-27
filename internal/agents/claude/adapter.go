package claude

import (
	"context"
	"os"
	"os/exec"

	"github.com/gentleman-programming/gentleman-ai-installer/internal/installcmd"
	"github.com/gentleman-programming/gentleman-ai-installer/internal/model"
	"github.com/gentleman-programming/gentleman-ai-installer/internal/system"
)

type statResult struct {
	isDir bool
	err   error
}

type Adapter struct {
	lookPath func(string) (string, error)
	statPath func(string) statResult
	resolver installcmd.Resolver
}

func NewAdapter() *Adapter {
	return &Adapter{
		lookPath: exec.LookPath,
		statPath: defaultStat,
		resolver: installcmd.NewResolver(),
	}
}

func (a *Adapter) Agent() model.AgentID {
	return model.AgentClaudeCode
}

func (a *Adapter) SupportsAutoInstall() bool {
	return true
}

func (a *Adapter) Detect(_ context.Context, homeDir string) (bool, string, string, bool, error) {
	configPath := ConfigPath(homeDir)

	binaryPath, err := a.lookPath("claude")
	installed := err == nil

	stat := a.statPath(configPath)
	if stat.err != nil {
		if os.IsNotExist(stat.err) {
			return installed, binaryPath, configPath, false, nil
		}
		return false, "", "", false, stat.err
	}

	return installed, binaryPath, configPath, stat.isDir, nil
}

func (a *Adapter) InstallCommand(profile system.PlatformProfile) ([]string, error) {
	resolver := a.resolver
	if resolver == nil {
		resolver = installcmd.NewResolver()
	}

	return resolver.ResolveAgentInstall(profile, a.Agent())
}

func defaultStat(path string) statResult {
	info, err := os.Stat(path)
	if err != nil {
		return statResult{err: err}
	}

	return statResult{isDir: info.IsDir()}
}
