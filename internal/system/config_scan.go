package system

import (
	"os"
	"path/filepath"
)

type ConfigState struct {
	Agent       string
	Path        string
	Exists      bool
	IsDirectory bool
}

func ScanConfigs(homeDir string) []ConfigState {
	paths := []ConfigState{
		{Agent: "claude-code", Path: filepath.Join(homeDir, ".claude")},
		{Agent: "opencode", Path: filepath.Join(homeDir, ".config", "opencode")},
	}

	for idx := range paths {
		info, err := os.Stat(paths[idx].Path)
		if err != nil {
			continue
		}

		paths[idx].Exists = true
		paths[idx].IsDirectory = info.IsDir()
	}

	return paths
}
