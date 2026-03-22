package update

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/gentleman-programming/gentle-ai/internal/system"
)

func TestCommandsForAllBrew(t *testing.T) {
	results := []UpdateResult{
		{Tool: ToolInfo{Name: "gentle-ai"}, Status: UpdateAvailable},
		{Tool: ToolInfo{Name: "engram"}, Status: UpdateAvailable},
		{Tool: ToolInfo{Name: "gga"}, Status: UpToDate},
	}

	got := CommandsForAll(results, system.PlatformProfile{PackageManager: "brew"})
	want := [][]string{{"brew", "update"}, {"brew", "upgrade", "gentle-ai", "engram"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("CommandsForAll() = %v, want %v", got, want)
	}
}

func TestCommandsForAllNonBrew(t *testing.T) {
	results := []UpdateResult{{Tool: ToolInfo{Name: "engram"}, Status: UpdateAvailable}}
	if got := CommandsForAll(results, system.PlatformProfile{PackageManager: "apt"}); got != nil {
		t.Fatalf("CommandsForAll() = %v, want nil", got)
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

	want := [][]string{{"brew", "update"}, {"brew", "upgrade", "engram"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ApplyAll() commands = %v, want %v", got, want)
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
