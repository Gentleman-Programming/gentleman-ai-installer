package update

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gentleman-programming/gentle-ai/internal/installcmd"
	"github.com/gentleman-programming/gentle-ai/internal/system"
)

const (
	gentleAIInstallScript = "https://raw.githubusercontent.com/Gentleman-Programming/gentle-ai/main/scripts/install.sh"
	engramModulePath      = "github.com/Gentleman-Programming/engram/cmd/engram@latest"
	ggaRepoURL            = "https://github.com/Gentleman-Programming/gentleman-guardian-angel.git"
)

func canUpdateTool(tool ToolInfo, profile system.PlatformProfile) bool {
	_, err := updateCommandsForTool(tool, profile)
	return err == nil
}

func updateCommandsForTool(tool ToolInfo, profile system.PlatformProfile) ([][]string, error) {
	switch tool.Name {
	case "gentle-ai":
		return gentleAICommands(profile)
	case "engram":
		return engramCommands(profile)
	case "gga":
		return ggaCommands(profile)
	default:
		return nil, fmt.Errorf("unsupported tool %q", tool.Name)
	}
}

func gentleAICommands(profile system.PlatformProfile) ([][]string, error) {
	switch {
	case profile.PackageManager == "brew":
		return [][]string{{"brew", "upgrade", "gentle-ai"}}, nil
	case profile.OS == "linux":
		return [][]string{{"bash", "-c", "curl -fsSL " + gentleAIInstallScript + " | bash"}}, nil
	case profile.OS == "windows":
		// Updating the currently running Windows executable in-place is not
		// reliable, so keep self-update as a manual action there for now.
		return nil, fmt.Errorf("gentle-ai self-update is not supported on windows")
	default:
		return nil, fmt.Errorf("unsupported platform for gentle-ai update: os=%q pm=%q", profile.OS, profile.PackageManager)
	}
}

func engramCommands(profile system.PlatformProfile) ([][]string, error) {
	switch profile.PackageManager {
	case "brew":
		return [][]string{{"brew", "upgrade", "engram"}}, nil
	case "apt", "pacman":
		return [][]string{{"env", "CGO_ENABLED=0", "go", "install", engramModulePath}}, nil
	case "winget":
		return [][]string{{"go", "install", engramModulePath}}, nil
	default:
		return nil, fmt.Errorf("unsupported platform for engram update: os=%q distro=%q pm=%q", profile.OS, profile.LinuxDistro, profile.PackageManager)
	}
}

func ggaCommands(profile system.PlatformProfile) ([][]string, error) {
	switch profile.PackageManager {
	case "brew":
		return [][]string{{"brew", "upgrade", "gga"}}, nil
	case "apt", "pacman":
		const tmpDir = "/tmp/gentleman-guardian-angel"
		return [][]string{
			{"rm", "-rf", tmpDir},
			{"git", "clone", ggaRepoURL, tmpDir},
			{"bash", filepath.ToSlash(filepath.Join(tmpDir, "install.sh"))},
		}, nil
	case "winget":
		cloneDst := filepath.Join(os.TempDir(), "gentleman-guardian-angel")
		return [][]string{
			{"powershell", "-NoProfile", "-Command", fmt.Sprintf("Remove-Item -Recurse -Force -ErrorAction SilentlyContinue '%s'; exit 0", cloneDst)},
			{"git", "clone", ggaRepoURL, cloneDst},
			{installcmd.GitBashPath(), strings.ReplaceAll(filepath.Join(cloneDst, "install.sh"), `\`, "/")},
		}, nil
	default:
		return nil, fmt.Errorf("unsupported platform for gga update: os=%q distro=%q pm=%q", profile.OS, profile.LinuxDistro, profile.PackageManager)
	}
}
