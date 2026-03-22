package update

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/gentleman-programming/gentle-ai/internal/system"
)

var runCommand = func(ctx context.Context, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if len(output) > 0 {
			return fmt.Errorf("%w\noutput:\n%s", err, strings.TrimSpace(string(output)))
		}
		return err
	}
	return nil
}

func CanUpdateAll(results []UpdateResult, profile system.PlatformProfile) bool {
	return len(ActionableResults(results, profile)) > 0
}

func ActionableResults(results []UpdateResult, profile system.PlatformProfile) []UpdateResult {
	if profile.PackageManager != "brew" {
		return nil
	}

	actionable := make([]UpdateResult, 0, len(results))
	for _, result := range results {
		if result.Status != UpdateAvailable {
			continue
		}
		if result.Tool.Name == "" {
			continue
		}
		actionable = append(actionable, result)
	}

	return actionable
}

func CommandsForAll(results []UpdateResult, profile system.PlatformProfile) [][]string {
	actionable := ActionableResults(results, profile)
	if len(actionable) == 0 {
		return nil
	}

	packages := make([]string, 0, len(actionable))
	seen := make(map[string]struct{}, len(actionable))
	for _, result := range actionable {
		if _, ok := seen[result.Tool.Name]; ok {
			continue
		}
		seen[result.Tool.Name] = struct{}{}
		packages = append(packages, result.Tool.Name)
	}

	return [][]string{{"brew", "update-if-needed"}, append([]string{"brew", "upgrade"}, packages...)}
}

func ApplyAll(ctx context.Context, results []UpdateResult, profile system.PlatformProfile) error {
	commands := CommandsForAll(results, profile)
	if len(commands) == 0 {
		return fmt.Errorf("update all is only available for Homebrew-managed updates")
	}

	for _, command := range commands {
		if err := runCommand(ctx, command[0], command[1:]...); err != nil {
			return fmt.Errorf("run %q: %w", strings.Join(command, " "), err)
		}
	}

	return nil
}
