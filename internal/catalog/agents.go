package catalog

import "github.com/gentleman-programming/gentleman-ai-installer/internal/model"

type Agent struct {
	ID         model.AgentID
	Name       string
	ConfigPath string
}

var mvpAgents = []Agent{
	{ID: model.AgentClaudeCode, Name: "Claude Code", ConfigPath: "~/.claude"},
	{ID: model.AgentOpenCode, Name: "OpenCode", ConfigPath: "~/.config/opencode"},
}

func MVPAgents() []Agent {
	agents := make([]Agent, len(mvpAgents))
	copy(agents, mvpAgents)
	return agents
}

func IsMVPAgent(agent model.AgentID) bool {
	for _, current := range mvpAgents {
		if current.ID == agent {
			return true
		}
	}

	return false
}
