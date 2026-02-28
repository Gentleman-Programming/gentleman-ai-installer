package agents

import (
	"context"

	"github.com/gentleman-programming/gentle-ai/internal/model"
	"github.com/gentleman-programming/gentle-ai/internal/system"
)

type Capability string

const (
	CapabilityAutoInstall Capability = "auto-install"
)

type Adapter interface {
	Agent() model.AgentID
	SupportsAutoInstall() bool
	Detect(ctx context.Context, homeDir string) (installed bool, binaryPath string, configPath string, configFound bool, err error)
	InstallCommand(profile system.PlatformProfile) ([]string, error)
}
