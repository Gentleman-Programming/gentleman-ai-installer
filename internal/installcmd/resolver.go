package installcmd

import (
	"fmt"

	"github.com/gentleman-programming/gentleman-ai-installer/internal/model"
	"github.com/gentleman-programming/gentleman-ai-installer/internal/system"
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
		return []string{"npm", "install", "-g", "@anthropic-ai/claude-code"}, nil
	case model.AgentOpenCode:
		return profileResolver{}.ResolveDependencyInstall(profile, "opencode")
	default:
		return nil, fmt.Errorf("install command is not supported for agent %q", agent)
	}
}

func (profileResolver) ResolveComponentInstall(profile system.PlatformProfile, component model.ComponentID) ([]string, error) {
	switch component {
	case model.ComponentEngram:
		return profileResolver{}.ResolveDependencyInstall(profile, "engram")
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
