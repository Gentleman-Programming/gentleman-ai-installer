package gga

import (
	"fmt"

	"github.com/gentleman-programming/gentle-ai/internal/installcmd"
	"github.com/gentleman-programming/gentle-ai/internal/model"
	"github.com/gentleman-programming/gentle-ai/internal/system"
)

func InstallCommand(profile system.PlatformProfile) ([]string, error) {
	return installcmd.NewResolver().ResolveDependencyInstall(profile, "gga")
}

func ShouldInstall(enabled bool) bool {
	return enabled
}

func AgentSupportsGGA(agent model.AgentID) bool {
	switch agent {
	case model.AgentClaudeCode, model.AgentOpenCode:
		return true
	default:
		return false
	}
}

func ValidateAgents(agents []model.AgentID) error {
	for _, agent := range agents {
		if !AgentSupportsGGA(agent) {
			return fmt.Errorf("gga is not supported for agent %q in MVP", agent)
		}
	}

	return nil
}
