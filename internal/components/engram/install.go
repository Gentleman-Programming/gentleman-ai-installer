package engram

import (
	"fmt"

	"github.com/gentleman-programming/gentleman-ai-installer/internal/installcmd"
	"github.com/gentleman-programming/gentleman-ai-installer/internal/model"
	"github.com/gentleman-programming/gentleman-ai-installer/internal/system"
)

func InstallCommand(profile system.PlatformProfile) ([]string, error) {
	return installcmd.NewResolver().ResolveDependencyInstall(profile, "engram")
}

func AgentSupportsEngram(agent model.AgentID) bool {
	switch agent {
	case model.AgentClaudeCode, model.AgentOpenCode:
		return true
	default:
		return false
	}
}

func ValidateAgents(agents []model.AgentID) error {
	for _, agent := range agents {
		if !AgentSupportsEngram(agent) {
			return fmt.Errorf("engram is not supported for agent %q in MVP", agent)
		}
	}

	return nil
}
