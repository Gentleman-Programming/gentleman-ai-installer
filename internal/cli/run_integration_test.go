package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/gentleman-programming/gentleman-ai-installer/internal/system"
)

func TestRunInstallAppliesFilesystemChanges(t *testing.T) {
	home := t.TempDir()
	restoreHome := osUserHomeDir
	restoreCommand := runCommand
	t.Cleanup(func() {
		osUserHomeDir = restoreHome
		runCommand = restoreCommand
	})

	osUserHomeDir = func() (string, error) { return home, nil }
	runCommand = func(string, ...string) error { return nil }

	result, err := RunInstall([]string{"--agent", "opencode", "--component", "permissions"}, system.DetectionResult{})
	if err != nil {
		t.Fatalf("RunInstall() error = %v", err)
	}

	if !result.Verify.Ready {
		t.Fatalf("verification ready = false, report = %#v", result.Verify)
	}

	settingsPath := filepath.Join(home, ".config", "opencode", "settings.json")
	if _, err := os.Stat(settingsPath); err != nil {
		t.Fatalf("expected settings file %q: %v", settingsPath, err)
	}
}

func TestRunInstallRollsBackOnComponentFailure(t *testing.T) {
	home := t.TempDir()
	settingsPath := filepath.Join(home, ".config", "opencode", "settings.json")
	if err := os.MkdirAll(filepath.Dir(settingsPath), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	before := []byte("{\n  \"existing\": true\n}\n")
	if err := os.WriteFile(settingsPath, before, 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	restoreHome := osUserHomeDir
	restoreCommand := runCommand
	t.Cleanup(func() {
		osUserHomeDir = restoreHome
		runCommand = restoreCommand
	})

	osUserHomeDir = func() (string, error) { return home, nil }
	runCommand = func(name string, args ...string) error {
		if name == "brew" && len(args) == 2 && args[0] == "install" && args[1] == "engram" {
			return os.ErrPermission
		}
		return nil
	}

	_, err := RunInstall(
		[]string{"--agent", "opencode", "--component", "context7", "--component", "engram"},
		system.DetectionResult{},
	)
	if err == nil {
		t.Fatalf("RunInstall() expected error")
	}

	if !strings.Contains(err.Error(), "execute install pipeline") {
		t.Fatalf("RunInstall() error = %v", err)
	}

	after, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	if string(after) != string(before) {
		t.Fatalf("settings content changed after rollback\nafter=%s\nbefore=%s", after, before)
	}
}

// --- Batch D: Linux profile runtime wiring integration tests ---

// linuxDetectionResult builds a DetectionResult with a Linux profile for integration tests.
func linuxDetectionResult(distro, pkgMgr string) system.DetectionResult {
	return system.DetectionResult{
		System: system.SystemInfo{
			OS:        "linux",
			Arch:      "amd64",
			Shell:     "/bin/bash",
			Supported: true,
			Profile: system.PlatformProfile{
				OS:             "linux",
				LinuxDistro:    distro,
				PackageManager: pkgMgr,
				Supported:      true,
			},
		},
	}
}

// commandRecorder captures all external commands invoked during a pipeline run.
type commandRecorder struct {
	mu       sync.Mutex
	commands []string
}

func (r *commandRecorder) record(name string, args ...string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.commands = append(r.commands, fmt.Sprintf("%s %s", name, strings.Join(args, " ")))
	return nil
}

func (r *commandRecorder) get() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := make([]string, len(r.commands))
	copy(cp, r.commands)
	return cp
}

func TestRunInstallLinuxUbuntuResolvesAptCommands(t *testing.T) {
	home := t.TempDir()
	restoreHome := osUserHomeDir
	restoreCommand := runCommand
	t.Cleanup(func() {
		osUserHomeDir = restoreHome
		runCommand = restoreCommand
	})

	osUserHomeDir = func() (string, error) { return home, nil }
	recorder := &commandRecorder{}
	runCommand = recorder.record

	detection := linuxDetectionResult(system.LinuxDistroUbuntu, "apt")
	result, err := RunInstall(
		[]string{"--agent", "opencode", "--component", "permissions"},
		detection,
	)
	if err != nil {
		t.Fatalf("RunInstall() error = %v", err)
	}

	if !result.Verify.Ready {
		t.Fatalf("verification ready = false, report = %#v", result.Verify)
	}

	// Verify platform decision was resolved from the Linux profile.
	if result.Resolved.PlatformDecision.OS != "linux" {
		t.Fatalf("platform decision OS = %q, want linux", result.Resolved.PlatformDecision.OS)
	}
	if result.Resolved.PlatformDecision.PackageManager != "apt" {
		t.Fatalf("platform decision package manager = %q, want apt", result.Resolved.PlatformDecision.PackageManager)
	}
}

func TestRunInstallLinuxArchResolvesPacmanCommands(t *testing.T) {
	home := t.TempDir()
	restoreHome := osUserHomeDir
	restoreCommand := runCommand
	t.Cleanup(func() {
		osUserHomeDir = restoreHome
		runCommand = restoreCommand
	})

	osUserHomeDir = func() (string, error) { return home, nil }
	recorder := &commandRecorder{}
	runCommand = recorder.record

	detection := linuxDetectionResult(system.LinuxDistroArch, "pacman")
	result, err := RunInstall(
		[]string{"--agent", "opencode", "--component", "permissions"},
		detection,
	)
	if err != nil {
		t.Fatalf("RunInstall() error = %v", err)
	}

	if !result.Verify.Ready {
		t.Fatalf("verification ready = false, report = %#v", result.Verify)
	}

	if result.Resolved.PlatformDecision.PackageManager != "pacman" {
		t.Fatalf("platform decision package manager = %q, want pacman", result.Resolved.PlatformDecision.PackageManager)
	}
}

func TestRunInstallLinuxUbuntuWithEngramResolvesGoInstallCommand(t *testing.T) {
	home := t.TempDir()
	restoreHome := osUserHomeDir
	restoreCommand := runCommand
	t.Cleanup(func() {
		osUserHomeDir = restoreHome
		runCommand = restoreCommand
	})

	osUserHomeDir = func() (string, error) { return home, nil }
	recorder := &commandRecorder{}
	runCommand = recorder.record

	detection := linuxDetectionResult(system.LinuxDistroUbuntu, "apt")
	result, err := RunInstall(
		[]string{"--agent", "opencode", "--component", "engram"},
		detection,
	)
	if err != nil {
		t.Fatalf("RunInstall() error = %v", err)
	}

	if !result.Verify.Ready {
		t.Fatalf("verification ready = false, report = %#v", result.Verify)
	}

	// Verify that at least one command used go install (the engram install command).
	commands := recorder.get()
	foundGoInstall := false
	for _, cmd := range commands {
		if strings.Contains(cmd, "env CGO_ENABLED=0 go install github.com/Gentleman-Programming/engram/cmd/engram@latest") {
			foundGoInstall = true
			break
		}
	}
	if !foundGoInstall {
		t.Fatalf("expected go install command for engram, got commands: %v", commands)
	}
}

func TestRunInstallLinuxArchWithEngramResolvesGoInstallCommand(t *testing.T) {
	home := t.TempDir()
	restoreHome := osUserHomeDir
	restoreCommand := runCommand
	t.Cleanup(func() {
		osUserHomeDir = restoreHome
		runCommand = restoreCommand
	})

	osUserHomeDir = func() (string, error) { return home, nil }
	recorder := &commandRecorder{}
	runCommand = recorder.record

	detection := linuxDetectionResult(system.LinuxDistroArch, "pacman")
	result, err := RunInstall(
		[]string{"--agent", "opencode", "--component", "engram"},
		detection,
	)
	if err != nil {
		t.Fatalf("RunInstall() error = %v", err)
	}

	if !result.Verify.Ready {
		t.Fatalf("verification ready = false, report = %#v", result.Verify)
	}

	commands := recorder.get()
	foundGoInstall := false
	for _, cmd := range commands {
		if strings.Contains(cmd, "env CGO_ENABLED=0 go install github.com/Gentleman-Programming/engram/cmd/engram@latest") {
			foundGoInstall = true
			break
		}
	}
	if !foundGoInstall {
		t.Fatalf("expected go install command for engram, got commands: %v", commands)
	}
}

func TestRunInstallLinuxRollsBackOnComponentFailure(t *testing.T) {
	home := t.TempDir()
	settingsPath := filepath.Join(home, ".config", "opencode", "settings.json")
	if err := os.MkdirAll(filepath.Dir(settingsPath), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	before := []byte("{\n  \"linux-original\": true\n}\n")
	if err := os.WriteFile(settingsPath, before, 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	restoreHome := osUserHomeDir
	restoreCommand := runCommand
	t.Cleanup(func() {
		osUserHomeDir = restoreHome
		runCommand = restoreCommand
	})

	osUserHomeDir = func() (string, error) { return home, nil }
	runCommand = func(name string, args ...string) error {
		// Fail the engram install command to trigger rollback.
		// Command is now: env CGO_ENABLED=0 go install .../engram@latest
		if name == "env" && strings.Contains(strings.Join(args, " "), "engram") {
			return os.ErrPermission
		}
		return nil
	}

	detection := linuxDetectionResult(system.LinuxDistroUbuntu, "apt")
	_, err := RunInstall(
		[]string{"--agent", "opencode", "--component", "context7", "--component", "engram"},
		detection,
	)
	if err == nil {
		t.Fatalf("RunInstall() expected error")
	}

	if !strings.Contains(err.Error(), "execute install pipeline") {
		t.Fatalf("RunInstall() error = %v", err)
	}

	// Verify rollback restored the original file.
	after, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	if string(after) != string(before) {
		t.Fatalf("settings content changed after rollback on Linux\nafter=%s\nbefore=%s", after, before)
	}
}

func TestRunInstallLinuxAgentInstallResolvesGoInstallCommand(t *testing.T) {
	home := t.TempDir()
	restoreHome := osUserHomeDir
	restoreCommand := runCommand
	t.Cleanup(func() {
		osUserHomeDir = restoreHome
		runCommand = restoreCommand
	})

	osUserHomeDir = func() (string, error) { return home, nil }
	recorder := &commandRecorder{}
	runCommand = recorder.record

	detection := linuxDetectionResult(system.LinuxDistroUbuntu, "apt")
	_, err := RunInstall(
		[]string{"--agent", "opencode", "--component", "permissions"},
		detection,
	)
	if err != nil {
		t.Fatalf("RunInstall() error = %v", err)
	}

	// OpenCode on Ubuntu should resolve via go install with CGO_ENABLED=0.
	commands := recorder.get()
	foundGoInstall := false
	for _, cmd := range commands {
		if strings.Contains(cmd, "env CGO_ENABLED=0 go install github.com/opencode-ai/opencode@latest") {
			foundGoInstall = true
			break
		}
	}
	if !foundGoInstall {
		t.Fatalf("expected go install command for opencode agent, got commands: %v", commands)
	}
}

// --- Batch E: Linux verification and macOS parity matrix ---

func TestRunInstallLinuxVerificationReportsReadyOnSuccess(t *testing.T) {
	home := t.TempDir()
	restoreHome := osUserHomeDir
	restoreCommand := runCommand
	t.Cleanup(func() {
		osUserHomeDir = restoreHome
		runCommand = restoreCommand
	})

	osUserHomeDir = func() (string, error) { return home, nil }
	runCommand = func(string, ...string) error { return nil }

	detection := linuxDetectionResult(system.LinuxDistroUbuntu, "apt")
	result, err := RunInstall(
		[]string{"--agent", "opencode", "--component", "permissions"},
		detection,
	)
	if err != nil {
		t.Fatalf("RunInstall() error = %v", err)
	}

	if !result.Verify.Ready {
		t.Fatalf("Verify.Ready = false, want true for successful Linux install")
	}
	if result.Verify.Failed != 0 {
		t.Fatalf("Verify.Failed = %d, want 0", result.Verify.Failed)
	}
}

func TestRunInstallLinuxArchVerificationReportsReadyOnSuccess(t *testing.T) {
	home := t.TempDir()
	restoreHome := osUserHomeDir
	restoreCommand := runCommand
	t.Cleanup(func() {
		osUserHomeDir = restoreHome
		runCommand = restoreCommand
	})

	osUserHomeDir = func() (string, error) { return home, nil }
	runCommand = func(string, ...string) error { return nil }

	detection := linuxDetectionResult(system.LinuxDistroArch, "pacman")
	result, err := RunInstall(
		[]string{"--agent", "opencode", "--component", "permissions"},
		detection,
	)
	if err != nil {
		t.Fatalf("RunInstall() error = %v", err)
	}

	if !result.Verify.Ready {
		t.Fatalf("Verify.Ready = false, want true for successful Arch install")
	}
}

func TestRunInstallLinuxDryRunSkipsVerification(t *testing.T) {
	detection := linuxDetectionResult(system.LinuxDistroUbuntu, "apt")
	result, err := RunInstall([]string{"--dry-run"}, detection)
	if err != nil {
		t.Fatalf("RunInstall() error = %v", err)
	}

	if !result.DryRun {
		t.Fatalf("DryRun = false, want true")
	}
	// Verify report should be zero-value (no checks run in dry-run)
	if result.Verify.Passed != 0 || result.Verify.Failed != 0 {
		t.Fatalf("expected zero verify counters in dry-run, got passed=%d failed=%d", result.Verify.Passed, result.Verify.Failed)
	}
}

func TestRunInstallLinuxDryRunPlatformDecisionRendersCorrectly(t *testing.T) {
	detection := linuxDetectionResult(system.LinuxDistroArch, "pacman")
	result, err := RunInstall([]string{"--dry-run"}, detection)
	if err != nil {
		t.Fatalf("RunInstall() error = %v", err)
	}

	output := RenderDryRun(result)
	want := "os=linux distro=arch package-manager=pacman status=supported"
	if !strings.Contains(output, want) {
		t.Fatalf("RenderDryRun() missing platform decision\noutput=%s\nwant contains=%s", output, want)
	}
}

// --- macOS parity regression checks ---

func macOSDetectionResult() system.DetectionResult {
	return system.DetectionResult{
		System: system.SystemInfo{
			OS:        "darwin",
			Arch:      "arm64",
			Shell:     "/bin/zsh",
			Supported: true,
			Profile: system.PlatformProfile{
				OS:             "darwin",
				PackageManager: "brew",
				Supported:      true,
			},
		},
	}
}

func TestRunInstallMacOSStillResolvesBrewCommands(t *testing.T) {
	home := t.TempDir()
	restoreHome := osUserHomeDir
	restoreCommand := runCommand
	t.Cleanup(func() {
		osUserHomeDir = restoreHome
		runCommand = restoreCommand
	})

	osUserHomeDir = func() (string, error) { return home, nil }
	recorder := &commandRecorder{}
	runCommand = recorder.record

	detection := macOSDetectionResult()
	result, err := RunInstall(
		[]string{"--agent", "opencode", "--component", "engram"},
		detection,
	)
	if err != nil {
		t.Fatalf("RunInstall() error = %v", err)
	}

	if !result.Verify.Ready {
		t.Fatalf("macOS verification ready = false")
	}

	// Verify brew install command was used, not apt or pacman.
	commands := recorder.get()
	foundBrew := false
	for _, cmd := range commands {
		if strings.Contains(cmd, "brew install engram") {
			foundBrew = true
			break
		}
	}
	if !foundBrew {
		t.Fatalf("expected brew install for macOS engram, got commands: %v", commands)
	}
}

func TestRunInstallMacOSDryRunPlatformDecision(t *testing.T) {
	detection := macOSDetectionResult()
	result, err := RunInstall([]string{"--dry-run"}, detection)
	if err != nil {
		t.Fatalf("RunInstall() error = %v", err)
	}

	if result.Resolved.PlatformDecision.OS != "darwin" {
		t.Fatalf("macOS platform decision OS = %q, want darwin", result.Resolved.PlatformDecision.OS)
	}
	if result.Resolved.PlatformDecision.PackageManager != "brew" {
		t.Fatalf("macOS platform decision PM = %q, want brew", result.Resolved.PlatformDecision.PackageManager)
	}
	if !result.Resolved.PlatformDecision.Supported {
		t.Fatalf("macOS platform decision Supported = false, want true")
	}
}

func TestRunInstallMacOSVerificationMatchesPreLinuxBehavior(t *testing.T) {
	home := t.TempDir()
	restoreHome := osUserHomeDir
	restoreCommand := runCommand
	t.Cleanup(func() {
		osUserHomeDir = restoreHome
		runCommand = restoreCommand
	})

	osUserHomeDir = func() (string, error) { return home, nil }
	runCommand = func(string, ...string) error { return nil }

	detection := macOSDetectionResult()
	result, err := RunInstall(
		[]string{"--agent", "opencode", "--component", "permissions"},
		detection,
	)
	if err != nil {
		t.Fatalf("RunInstall() error = %v", err)
	}

	if !result.Verify.Ready {
		t.Fatalf("macOS verify ready = false, want true")
	}
	if result.Verify.Failed != 0 {
		t.Fatalf("macOS verify failed = %d, want 0", result.Verify.Failed)
	}
}

func TestRunInstallMacOSRollbackStillWorks(t *testing.T) {
	home := t.TempDir()
	settingsPath := filepath.Join(home, ".config", "opencode", "settings.json")
	if err := os.MkdirAll(filepath.Dir(settingsPath), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	before := []byte("{\n  \"macos-original\": true\n}\n")
	if err := os.WriteFile(settingsPath, before, 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	restoreHome := osUserHomeDir
	restoreCommand := runCommand
	t.Cleanup(func() {
		osUserHomeDir = restoreHome
		runCommand = restoreCommand
	})

	osUserHomeDir = func() (string, error) { return home, nil }
	runCommand = func(name string, args ...string) error {
		if name == "brew" && len(args) == 2 && args[0] == "install" && args[1] == "engram" {
			return os.ErrPermission
		}
		return nil
	}

	detection := macOSDetectionResult()
	_, err := RunInstall(
		[]string{"--agent", "opencode", "--component", "context7", "--component", "engram"},
		detection,
	)
	if err == nil {
		t.Fatalf("RunInstall() expected error")
	}

	if !strings.Contains(err.Error(), "execute install pipeline") {
		t.Fatalf("RunInstall() error = %v", err)
	}

	after, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	if string(after) != string(before) {
		t.Fatalf("macOS settings changed after rollback\nafter=%s\nbefore=%s", after, before)
	}
}
