package installcmd

import (
	"fmt"

	"github.com/gentleman-programming/gentle-ai/internal/model"
	"github.com/gentleman-programming/gentle-ai/internal/system"
)

type Resolver interface {
	ResolveAgentInstall(profile system.PlatformProfile, agent model.AgentID) ([]string, error)
	ResolveComponentInstall(profile system.PlatformProfile, component model.ComponentID) ([]string, error)
	ResolveDependencyInstall(profile system.PlatformProfile, dependency string) ([]string, error)
}

type profileResolver struct{}

func NewResolver() Resolver {
	return profileResolver{}
}

func (profileResolver) ResolveAgentInstall(profile system.PlatformProfile, agent model.AgentID) ([]string, error) {
	switch agent {
	case model.AgentClaudeCode:
		return resolveClaudeCodeInstall(profile), nil
	case model.AgentOpenCode:
		return resolveOpenCodeInstall(profile)
	default:
		return nil, fmt.Errorf("install command is not supported for agent %q", agent)
	}
}

// resolveClaudeCodeInstall returns the npm install command for Claude Code.
// On Linux, sudo is required because npm global installs write to system directories.
func resolveClaudeCodeInstall(profile system.PlatformProfile) []string {
	if profile.OS == "linux" {
		return []string{"sudo", "npm", "install", "-g", "@anthropic-ai/claude-code"}
	}
	return []string{"npm", "install", "-g", "@anthropic-ai/claude-code"}
}

func (profileResolver) ResolveComponentInstall(profile system.PlatformProfile, component model.ComponentID) ([]string, error) {
	switch component {
	case model.ComponentEngram:
		return resolveEngramInstall(profile)
	default:
		return nil, fmt.Errorf("install command is not supported for component %q", component)
	}
}

func (profileResolver) ResolveDependencyInstall(profile system.PlatformProfile, dependency string) ([]string, error) {
	if dependency == "" {
		return nil, fmt.Errorf("dependency name is required")
	}

	switch profile.PackageManager {
	case "brew":
		return []string{"brew", "install", dependency}, nil
	case "apt":
		return []string{"sudo", "apt-get", "install", "-y", dependency}, nil
	case "pacman":
		return []string{"sudo", "pacman", "-S", "--noconfirm", dependency}, nil
	default:
		return nil, fmt.Errorf(
			"unsupported package manager %q for os=%q distro=%q",
			profile.PackageManager,
			profile.OS,
			profile.LinuxDistro,
		)
	}
}

// resolveOpenCodeInstall returns the correct install command for OpenCode per platform.
// - darwin: brew install opencode (available in Homebrew)
// - arch: pacman -S opencode (available in extra/)
// - ubuntu/debian: go install (not in apt repos)
func resolveOpenCodeInstall(profile system.PlatformProfile) ([]string, error) {
	switch profile.PackageManager {
	case "brew":
		return []string{"brew", "install", "opencode"}, nil
	case "pacman":
		return []string{"sudo", "pacman", "-S", "--noconfirm", "opencode"}, nil
	case "apt":
		return []string{"env", "CGO_ENABLED=0", "go", "install", "github.com/opencode-ai/opencode@latest"}, nil
	default:
		return nil, fmt.Errorf(
			"unsupported platform for opencode: os=%q distro=%q pm=%q",
			profile.OS, profile.LinuxDistro, profile.PackageManager,
		)
	}
}

// resolveEngramInstall returns the correct install command for Engram per platform.
// - darwin: brew install (via Gentleman-Programming/homebrew-tap)
// - linux: go install (engram is not in any Linux distro's repos)
func resolveEngramInstall(profile system.PlatformProfile) ([]string, error) {
	switch profile.PackageManager {
	case "brew":
		return []string{"brew", "install", "engram"}, nil
	case "apt", "pacman":
		return []string{"env", "CGO_ENABLED=0", "go", "install", "github.com/Gentleman-Programming/engram/cmd/engram@latest"}, nil
	default:
		return nil, fmt.Errorf(
			"unsupported platform for engram: os=%q distro=%q pm=%q",
			profile.OS, profile.LinuxDistro, profile.PackageManager,
		)
	}
}
