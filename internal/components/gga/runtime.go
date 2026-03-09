package gga

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gentleman-programming/gentle-ai/internal/assets"
	"github.com/gentleman-programming/gentle-ai/internal/components/filemerge"
)

// RuntimeLibDir returns the runtime lib path used by gga.
func RuntimeLibDir(homeDir string) string {
	return filepath.Join(homeDir, ".local", "share", "gga", "lib")
}

// RuntimePRModePath returns the expected pr_mode.sh runtime path.
func RuntimePRModePath(homeDir string) string {
	return filepath.Join(RuntimeLibDir(homeDir), "pr_mode.sh")
}

// EnsureRuntimeAssets makes sure critical gga runtime files exist.
// Today this guards against upstream installer drift where pr_mode.sh may be missing.
func EnsureRuntimeAssets(homeDir string) error {
	prModePath := RuntimePRModePath(homeDir)
	if _, err := os.Stat(prModePath); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("stat gga runtime file %q: %w", prModePath, err)
	}

	content, err := assets.Read("gga/pr_mode.sh")
	if err != nil {
		return fmt.Errorf("read embedded gga runtime asset pr_mode.sh: %w", err)
	}

	if _, err := filemerge.WriteFileAtomic(prModePath, []byte(content), 0o755); err != nil {
		return fmt.Errorf("write gga runtime file %q: %w", prModePath, err)
	}

	return nil
}
