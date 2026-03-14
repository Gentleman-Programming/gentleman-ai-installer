package sdd

import (
	"io/fs"

	"github.com/gentleman-programming/gentle-ai/internal/assets"
	"github.com/gentleman-programming/gentle-ai/internal/model"
)

// SDDCommandNamesForAgent returns the base names of SDD command files for the given
// agent, derived from the embedded asset directory for that agent. Returns nil if the
// agent has no embedded command assets.
func SDDCommandNamesForAgent(agentID model.AgentID) []string {
	assetDir := commandsAssetDir(agentID)
	if assetDir == "" {
		return nil
	}
	entries, err := fs.ReadDir(assets.FS, assetDir)
	if err != nil {
		return nil
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if len(name) > 3 && name[len(name)-3:] == ".md" {
			name = name[:len(name)-3]
		}
		names = append(names, name)
	}
	return names
}

type OpenCodeCommand struct {
	Name        string
	Description string
	Body        string
}

func OpenCodeCommands() []OpenCodeCommand {
	return []OpenCodeCommand{
		{Name: "sdd-init", Description: "Initialize SDD context", Body: "/sdd-init"},
		{Name: "sdd-new", Description: "Start a new SDD change", Body: "/sdd-new ${change-name}"},
		{Name: "sdd-continue", Description: "Continue next pending artifact", Body: "/sdd-continue ${change-name}"},
		{Name: "sdd-ff", Description: "Generate all planning artifacts", Body: "/sdd-ff ${change-name}"},
		{Name: "sdd-apply", Description: "Implement tasks", Body: "/sdd-apply ${change-name}"},
		{Name: "sdd-verify", Description: "Verify implementation", Body: "/sdd-verify ${change-name}"},
		{Name: "sdd-archive", Description: "Archive completed change", Body: "/sdd-archive ${change-name}"},
	}
}

// SDDCommandNames returns the base names of all SDD Claude command files (without extension).
// Used by both injection and cleanup logic for Claude Code.
// Only the 3 user-facing meta-commands are included here; the remaining SDD phases
// are available as skills and do not need separate command files.
func SDDCommandNames() []string {
	return []string{
		"sdd-new",
		"sdd-continue",
		"sdd-ff",
	}
}
