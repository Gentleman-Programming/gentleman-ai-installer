package gga

import (
	"os"
	"strings"
	"testing"
)

func TestEnsureRuntimeAssetsCreatesPRModeWhenMissing(t *testing.T) {
	home := t.TempDir()
	path := RuntimePRModePath(home)

	if err := EnsureRuntimeAssets(home); err != nil {
		t.Fatalf("EnsureRuntimeAssets() error = %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}

	text := string(content)
	if !strings.Contains(text, "detect_base_branch") {
		t.Fatalf("runtime pr_mode.sh missing expected content")
	}
}

func TestEnsureRuntimeAssetsDoesNotOverwriteExistingPRMode(t *testing.T) {
	home := t.TempDir()
	path := RuntimePRModePath(home)
	if err := os.MkdirAll(RuntimeLibDir(home), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	const existing = "#!/usr/bin/env bash\n# custom\n"
	if err := os.WriteFile(path, []byte(existing), 0o755); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if err := EnsureRuntimeAssets(home); err != nil {
		t.Fatalf("EnsureRuntimeAssets() error = %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}

	if string(content) != existing {
		t.Fatalf("existing pr_mode.sh was modified")
	}
}
