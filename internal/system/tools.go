package system

import (
	"context"
	"os/exec"
)

type ToolStatus struct {
	Name      string
	Installed bool
	Path      string
}

func DetectTools(ctx context.Context, names []string) map[string]ToolStatus {
	tools := make(map[string]ToolStatus, len(names))

	for _, name := range names {
		path, err := exec.LookPath(name)
		status := ToolStatus{Name: name, Installed: err == nil}
		if err == nil {
			status.Path = path
		}
		tools[name] = status
	}

	_ = ctx

	return tools
}
