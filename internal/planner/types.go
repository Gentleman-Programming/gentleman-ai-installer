package planner

import (
	"github.com/gentleman-programming/gentleman-ai-installer/internal/model"
	"github.com/gentleman-programming/gentleman-ai-installer/internal/system"
)

type Resolver interface {
	Resolve(selection model.Selection) (ResolvedPlan, error)
}

type ResolvedPlan struct {
	Agents            []model.AgentID
	UnsupportedAgents []model.AgentID
	OrderedComponents []model.ComponentID
	AddedDependencies []model.ComponentID
	PlatformDecision  PlatformDecision
}

type ReviewPayload struct {
	Agents            []model.AgentID
	UnsupportedAgents []model.AgentID
	Persona           model.PersonaID
	Preset            model.PresetID
	Components        []ComponentAction
	AddedDependencies []model.ComponentID
	PlatformDecision  PlatformDecision
}

type PlatformDecision struct {
	OS             string
	LinuxDistro    string
	PackageManager string
	Supported      bool
}

func PlatformDecisionFromProfile(profile system.PlatformProfile) PlatformDecision {
	return PlatformDecision{
		OS:             profile.OS,
		LinuxDistro:    profile.LinuxDistro,
		PackageManager: profile.PackageManager,
		Supported:      profile.Supported,
	}
}

type ComponentAction struct {
	ID     model.ComponentID
	Action string
}
