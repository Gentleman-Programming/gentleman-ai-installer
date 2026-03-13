package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gentleman-programming/gentle-ai/internal/system"
)

func TestGGAAvailableDetectsPs1Wrapper(t *testing.T) {
	tmpDir := t.TempDir()

	origHomeDir := osUserHomeDir
	osUserHomeDir = func() (string, error) { return tmpDir, nil }
	t.Cleanup(func() { osUserHomeDir = origHomeDir })

	origStat := osStat
	osStat = func(name string) (os.FileInfo, error) {
		// gga (no extension) does not exist — only gga.ps1 does
		if name == filepath.Join(tmpDir, ".local", "bin", "gga.ps1") {
			return os.Stat(t.TempDir()) // any valid FileInfo
		}
		return nil, os.ErrNotExist
	}
	t.Cleanup(func() { osStat = origStat })

	origLookPath := cmdLookPath
	cmdLookPath = func(file string) (string, error) { return "", os.ErrNotExist }
	t.Cleanup(func() { cmdLookPath = origLookPath })

	if !ggaAvailable(system.PlatformProfile{OS: "windows"}) {
		t.Fatal("ggaAvailable() = false, want true when gga.ps1 exists")
	}
}
