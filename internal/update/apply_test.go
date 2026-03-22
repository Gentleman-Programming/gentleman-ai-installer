package update

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/gentleman-programming/gentle-ai/internal/installcmd"
	"github.com/gentleman-programming/gentle-ai/internal/system"
)

func TestCommandsForAllBrew(t *testing.T) {
	results := []UpdateResult{
		{Tool: ToolInfo{Name: "gentle-ai"}, Status: UpdateAvailable},
		{Tool: ToolInfo{Name: "engram"}, Status: UpdateAvailable},
		{Tool: ToolInfo{Name: "gga"}, Status: UpToDate},
	}

	got := CommandsForAll(results, system.PlatformProfile{PackageManager: "brew"})
	want := [][]string{{"brew", "update-if-needed"}, {"brew", "upgrade", "gentle-ai", "engram"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("CommandsForAll() = %v, want %v", got, want)
	}
}

func TestCommandsForAllBrewIncludesUnknownPackages(t *testing.T) {
	results := []UpdateResult{
		{Tool: ToolInfo{Name: "gentle-ai"}, Status: UpdateAvailable},
		{Tool: ToolInfo{Name: "mise"}, Status: UpdateAvailable},
		{Tool: ToolInfo{Name: "mise"}, Status: UpdateAvailable},
		{Tool: ToolInfo{Name: ""}, Status: UpdateAvailable},
		{Tool: ToolInfo{Name: "ignored"}, Status: UpToDate},
	}

	got := CommandsForAll(results, system.PlatformProfile{PackageManager: "brew"})
	want := [][]string{{"brew", "update-if-needed"}, {"brew", "upgrade", "gentle-ai", "mise"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("CommandsForAll() = %v, want %v", got, want)
	}
}

func TestCommandsForAllLinux(t *testing.T) {
	results := []UpdateResult{
		{Tool: ToolInfo{Name: "gentle-ai"}, Status: UpdateAvailable},
		{Tool: ToolInfo{Name: "engram"}, Status: UpdateAvailable},
		{Tool: ToolInfo{Name: "gga"}, Status: UpdateAvailable},
	}

	got := CommandsForAll(results, system.PlatformProfile{OS: "linux", LinuxDistro: system.LinuxDistroUbuntu, PackageManager: "apt"})
	want := [][]string{
		{"bash", "-c", "curl -fsSL https://raw.githubusercontent.com/Gentleman-Programming/gentle-ai/main/scripts/install.sh | bash"},
		{"env", "CGO_ENABLED=0", "go", "install", "github.com/Gentleman-Programming/engram/cmd/engram@latest"},
		{"rm", "-rf", "/tmp/gentleman-guardian-angel"},
		{"git", "clone", "https://github.com/Gentleman-Programming/gentleman-guardian-angel.git", "/tmp/gentleman-guardian-angel"},
		{"bash", "/tmp/gentleman-guardian-angel/install.sh"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("CommandsForAll() = %v, want %v", got, want)
	}
}

func TestCommandsForAllWindowsSkipsGentleAISelfUpdate(t *testing.T) {
	results := []UpdateResult{
		{Tool: ToolInfo{Name: "gentle-ai"}, Status: UpdateAvailable},
		{Tool: ToolInfo{Name: "engram"}, Status: UpdateAvailable},
		{Tool: ToolInfo{Name: "gga"}, Status: UpdateAvailable},
	}

	cloneDst := filepath.Join(os.TempDir(), "gentleman-guardian-angel")
	got := CommandsForAll(results, system.PlatformProfile{OS: "windows", PackageManager: "winget"})
	want := [][]string{
		{"go", "install", "github.com/Gentleman-Programming/engram/cmd/engram@latest"},
		{"powershell", "-NoProfile", "-Command", fmt.Sprintf("Remove-Item -Recurse -Force -ErrorAction SilentlyContinue '%s'; exit 0", cloneDst)},
		{"git", "clone", "https://github.com/Gentleman-Programming/gentleman-guardian-angel.git", cloneDst},
		{installcmd.GitBashPath(), strings.ReplaceAll(filepath.Join(cloneDst, "install.sh"), `\`, "/")},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("CommandsForAll() = %v, want %v", got, want)
	}
}

func TestApplyAllRunsCommandsSequentially(t *testing.T) {
	originalRunCommand := runCommand
	defer func() { runCommand = originalRunCommand }()

	var got [][]string
	runCommand = func(_ context.Context, name string, args ...string) error {
		got = append(got, append([]string{name}, args...))
		return nil
	}

	err := ApplyAll(context.Background(), []UpdateResult{{Tool: ToolInfo{Name: "engram"}, Status: UpdateAvailable}}, system.PlatformProfile{PackageManager: "brew"})
	if err != nil {
		t.Fatalf("ApplyAll() error = %v", err)
	}

	want := [][]string{{"brew", "update-if-needed"}, {"brew", "upgrade", "engram"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ApplyAll() commands = %v, want %v", got, want)
	}
}

func TestApplyAllReturnsPlatformUnavailableError(t *testing.T) {
	err := ApplyAll(context.Background(), []UpdateResult{{Tool: ToolInfo{Name: "gentle-ai"}, Status: UpdateAvailable}}, system.PlatformProfile{OS: "windows", PackageManager: "winget"})
	if err == nil {
		t.Fatal("ApplyAll() error = nil, want non-nil")
	}
}

func TestApplyAllReturnsCommandError(t *testing.T) {
	originalRunCommand := runCommand
	defer func() { runCommand = originalRunCommand }()

	runCommand = func(_ context.Context, name string, args ...string) error {
		return fmt.Errorf("boom")
	}

	err := ApplyAll(context.Background(), []UpdateResult{{Tool: ToolInfo{Name: "engram"}, Status: UpdateAvailable}}, system.PlatformProfile{PackageManager: "brew"})
	if err == nil {
		t.Fatal("ApplyAll() error = nil, want non-nil")
	}
}
