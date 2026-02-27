package gga

import (
	"path/filepath"

	"github.com/gentleman-programming/gentleman-ai-installer/internal/components/filemerge"
)

type ConfigResult struct {
	Changed bool
	File    string
}

var defaultConfigJSON = []byte("{\n  \"enabled\": true,\n  \"providers\": [\n    \"claude-code\",\n    \"opencode\"\n  ],\n  \"defaultProvider\": \"claude-code\"\n}\n")

func WriteDefaultConfig(homeDir string) (ConfigResult, error) {
	path := filepath.Join(homeDir, ".config", "gga", "config.json")
	writeResult, err := filemerge.WriteFileAtomic(path, defaultConfigJSON, 0o644)
	if err != nil {
		return ConfigResult{}, err
	}

	return ConfigResult{Changed: writeResult.Changed, File: path}, nil
}
