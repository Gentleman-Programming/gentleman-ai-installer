package installcmd

import (
	"fmt"
	"os"

	"github.com/gentleman-programming/gentle-ai/internal/model"
	"github.com/gentleman-programming/gentle-ai/internal/system"
)

// CommandSequence represents an ordered list of commands to run in sequence.
// Each inner slice is a single command with its arguments (e.g., ["brew", "install", "engram"]).
// Multi-step installs (e.g., tap + install) are expressed as multiple entries.
type CommandSequence = [][]string

type Resolver interface {
	ResolveAgentInstall(profile system.PlatformProfile, agent model.AgentID) (CommandSequence, error)
	ResolveComponentInstall(profile system.PlatformProfile, component model.ComponentID) (CommandSequence, error)
	ResolveDependencyInstall(profile system.PlatformProfile, dependency string) (CommandSequence, error)
}

type profileResolver struct{}

func NewResolver() Resolver {
	return profileResolver{}
}

func (profileResolver) ResolveAgentInstall(profile system.PlatformProfile, agent model.AgentID) (CommandSequence, error) {
	switch agent {
	case model.AgentClaudeCode:
		return resolveClaudeCodeInstall(profile), nil
	case model.AgentOpenCode:
		return resolveOpenCodeInstall(profile)
	default:
		return nil, fmt.Errorf("install command is not supported for agent %q", agent)
	}
}

// resolveClaudeCodeInstall returns the npm install command sequence for Claude Code.
// On Linux with system npm, sudo is required. With nvm/fnm/volta, it is not.
// On Windows and macOS, sudo is never needed.
func resolveClaudeCodeInstall(profile system.PlatformProfile) CommandSequence {
	if profile.OS == "linux" && !profile.NpmWritable {
		return CommandSequence{{"sudo", "npm", "install", "-g", "@anthropic-ai/claude-code"}}
	}
	return CommandSequence{{"npm", "install", "-g", "@anthropic-ai/claude-code"}}
}

func (profileResolver) ResolveComponentInstall(profile system.PlatformProfile, component model.ComponentID) (CommandSequence, error) {
	switch component {
	case model.ComponentEngram:
		return resolveEngramInstall(profile)
	case model.ComponentGGA:
		return resolveGGAInstall(profile)
	default:
		return nil, fmt.Errorf("install command is not supported for component %q", component)
	}
}

func (profileResolver) ResolveDependencyInstall(profile system.PlatformProfile, dependency string) (CommandSequence, error) {
	if dependency == "" {
		return nil, fmt.Errorf("dependency name is required")
	}

	switch profile.PackageManager {
	case "brew":
		return CommandSequence{{"brew", "install", dependency}}, nil
	case "apt":
		return CommandSequence{{"sudo", "apt-get", "install", "-y", dependency}}, nil
	case "pacman":
		return CommandSequence{{"sudo", "pacman", "-S", "--noconfirm", dependency}}, nil
	case "winget":
		return CommandSequence{{"winget", "install", "--id", dependency, "-e", "--accept-source-agreements", "--accept-package-agreements"}}, nil
	default:
		return nil, fmt.Errorf(
			"unsupported package manager %q for os=%q distro=%q",
			profile.PackageManager,
			profile.OS,
			profile.LinuxDistro,
		)
	}
}

// resolveOpenCodeInstall returns the correct install command sequence for OpenCode per platform.
// - darwin: brew install anomalyco/tap/opencode (official OpenCode tap)
// - linux: npm install -g opencode-ai (official npm package)
// See https://opencode.ai/docs for official install methods.
func resolveOpenCodeInstall(profile system.PlatformProfile) (CommandSequence, error) {
	switch profile.PackageManager {
	case "brew":
		return CommandSequence{
			{"brew", "install", "anomalyco/tap/opencode"},
		}, nil
	case "apt", "pacman":
		if profile.NpmWritable {
			return CommandSequence{{"npm", "install", "-g", "opencode-ai"}}, nil
		}
		return CommandSequence{{"sudo", "npm", "install", "-g", "opencode-ai"}}, nil
	case "winget":
		// On Windows, npm global installs do not require sudo.
		return CommandSequence{{"npm", "install", "-g", "opencode-ai"}}, nil
	default:
		return nil, fmt.Errorf(
			"unsupported platform for opencode: os=%q distro=%q pm=%q",
			profile.OS, profile.LinuxDistro, profile.PackageManager,
		)
	}
}

// resolveGGAInstall returns the correct install command sequence for GGA per platform.
// - darwin: brew tap + brew install (via Gentleman-Programming/homebrew-tap)
// - linux: git clone + install.sh (GGA is a pure Bash project, NOT a Go module)
func resolveGGAInstall(profile system.PlatformProfile) (CommandSequence, error) {
	switch profile.PackageManager {
	case "brew":
		return CommandSequence{
			{"brew", "tap", "Gentleman-Programming/homebrew-tap"},
			{"brew", "install", "gga"},
		}, nil
	case "apt", "pacman":
		const tmpDir = "/tmp/gentleman-guardian-angel"
		return CommandSequence{
			{"rm", "-rf", tmpDir},
			{"git", "clone", "https://github.com/Gentleman-Programming/gentleman-guardian-angel.git", tmpDir},
			{"bash", tmpDir + "/install.sh"},
		}, nil
	case "winget":
		// On Windows, use Git Bash (bundled with Git for Windows) to run the install script.
		// Clean up any leftover directory from a previous run before cloning.
		tmpDir := os.TempDir() + "\\gentleman-guardian-angel"
		return CommandSequence{
			{"cmd", "/c", fmt.Sprintf(`if exist "%s" rmdir /s /q "%s"`, tmpDir, tmpDir)},
			{"git", "clone", "https://github.com/Gentleman-Programming/gentleman-guardian-angel.git", tmpDir},
			{"bash", tmpDir + "\\install.sh"},
		}, nil
	default:
		return nil, fmt.Errorf(
			"unsupported platform for gga: os=%q distro=%q pm=%q",
			profile.OS, profile.LinuxDistro, profile.PackageManager,
		)
	}
}

// resolveEngramInstall returns the correct install command sequence for Engram per platform.
// - darwin: brew tap + brew install (via Gentleman-Programming/homebrew-tap)
// - linux: go install (engram is not in any Linux distro's repos)
func resolveEngramInstall(profile system.PlatformProfile) (CommandSequence, error) {
	switch profile.PackageManager {
	case "brew":
		return CommandSequence{
			{"brew", "tap", "Gentleman-Programming/homebrew-tap"},
			{"brew", "install", "engram"},
		}, nil
	case "apt", "pacman":
		return CommandSequence{{"env", "CGO_ENABLED=0", "go", "install", "github.com/Gentleman-Programming/engram/cmd/engram@latest"}}, nil
	case "winget":
		// On Windows, use go install (Engram has no winget package yet).
		return CommandSequence{{"go", "install", "github.com/Gentleman-Programming/engram/cmd/engram@latest"}}, nil
	default:
		return nil, fmt.Errorf(
			"unsupported platform for engram: os=%q distro=%q pm=%q",
			profile.OS, profile.LinuxDistro, profile.PackageManager,
		)
	}
}
