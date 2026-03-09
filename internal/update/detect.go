package update

import (
	"context"
	"os/exec"
	"regexp"
	"strings"
)

// Package-level vars for testability (swap in tests via t.Cleanup).
var (
	execCommand = exec.Command
	lookPath    = exec.LookPath
)

// versionRegexp extracts a semver-like version from command output.
// Same pattern as internal/system/deps.go for consistency.
var versionRegexp = regexp.MustCompile(`(\d+\.\d+(?:\.\d+)?)`)

// detectInstalledVersion determines the installed version of a tool.
// For tools with nil DetectCmd (gentle-ai), returns currentBuildVersion.
// For other tools, checks LookPath then runs the detect command.
func detectInstalledVersion(ctx context.Context, tool ToolInfo, currentBuildVersion string) string {
	if tool.DetectCmd == nil {
		return currentBuildVersion
	}

	if len(tool.DetectCmd) == 0 {
		return ""
	}

	binary := tool.DetectCmd[0]
	if _, err := lookPath(binary); err != nil {
		return "" // binary not found
	}

	cmd := execCommand(tool.DetectCmd[0], tool.DetectCmd[1:]...)
	out, err := cmd.Output()
	if err != nil {
		return "" // command failed — binary exists but version unknown
	}

	return parseVersionFromOutput(strings.TrimSpace(string(out)))
}

// parseVersionFromOutput extracts the first semver-like pattern from raw output.
func parseVersionFromOutput(output string) string {
	if output == "" {
		return ""
	}

	match := versionRegexp.FindStringSubmatch(output)
	if len(match) >= 2 {
		return match[1]
	}

	return ""
}
