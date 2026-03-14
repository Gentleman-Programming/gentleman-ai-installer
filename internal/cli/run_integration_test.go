package cli

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gentleman-programming/gentle-ai/internal/agents/opencode"
	"github.com/gentleman-programming/gentle-ai/internal/system"
)

// missingBinaryLookPath simulates all installable binaries (engram, gga) as
// missing while keeping Go available (needed for Linux engram go-install path).
func missingBinaryLookPath(name string) (string, error) {
	if name == "go" {
		return "/usr/local/bin/go", nil
	}
	return "", exec.ErrNotFound
}

func TestRunInstallAppliesFilesystemChanges(t *testing.T) {
	home := t.TempDir()
	restoreHome := osUserHomeDir
	restoreCommand := runCommand
	restoreLookPath := cmdLookPath
	t.Cleanup(func() {
		osUserHomeDir = restoreHome
		runCommand = restoreCommand
		cmdLookPath = restoreLookPath
	})

	osUserHomeDir = func() (string, error) { return home, nil }
	runCommand = func(string, ...string) error { return nil }
	cmdLookPath = missingBinaryLookPath

	result, err := RunInstall([]string{"--agent", "opencode", "--component", "permissions"}, system.DetectionResult{})
	if err != nil {
		t.Fatalf("RunInstall() error = %v", err)
	}

	if !result.Verify.Ready {
		t.Fatalf("verification ready = false, report = %#v", result.Verify)
	}

	configPath := filepath.Join(home, ".config", "opencode", "opencode.json")
	if _, err := os.Stat(configPath); err != nil {
		t.Fatalf("expected config file %q: %v", configPath, err)
	}
}

func TestRunInstallRollsBackOnComponentFailure(t *testing.T) {
	home := t.TempDir()
	settingsPath := filepath.Join(home, ".config", "opencode", "opencode.json")
	if err := os.MkdirAll(filepath.Dir(settingsPath), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	before := []byte("{\n  \"existing\": true\n}\n")
	if err := os.WriteFile(settingsPath, before, 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	restoreHome := osUserHomeDir
	restoreCommand := runCommand
	restoreLookPath := cmdLookPath
	t.Cleanup(func() {
		osUserHomeDir = restoreHome
		runCommand = restoreCommand
		cmdLookPath = restoreLookPath
	})
	cmdLookPath = missingBinaryLookPath

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
	restoreLookPath := cmdLookPath
	t.Cleanup(func() {
		osUserHomeDir = restoreHome
		runCommand = restoreCommand
		cmdLookPath = restoreLookPath
	})

	osUserHomeDir = func() (string, error) { return home, nil }
	cmdLookPath = missingBinaryLookPath
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
	restoreLookPath := cmdLookPath
	t.Cleanup(func() {
		osUserHomeDir = restoreHome
		runCommand = restoreCommand
		cmdLookPath = restoreLookPath
	})

	osUserHomeDir = func() (string, error) { return home, nil }
	cmdLookPath = missingBinaryLookPath
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
	restoreLookPath := cmdLookPath
	t.Cleanup(func() {
		osUserHomeDir = restoreHome
		runCommand = restoreCommand
		cmdLookPath = restoreLookPath
	})

	osUserHomeDir = func() (string, error) { return home, nil }
	cmdLookPath = missingBinaryLookPath
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
	restoreLookPath := cmdLookPath
	t.Cleanup(func() {
		osUserHomeDir = restoreHome
		runCommand = restoreCommand
		cmdLookPath = restoreLookPath
	})

	osUserHomeDir = func() (string, error) { return home, nil }
	cmdLookPath = missingBinaryLookPath
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
	settingsPath := filepath.Join(home, ".config", "opencode", "opencode.json")
	if err := os.MkdirAll(filepath.Dir(settingsPath), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	before := []byte("{\n  \"linux-original\": true\n}\n")
	if err := os.WriteFile(settingsPath, before, 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	restoreHome := osUserHomeDir
	restoreCommand := runCommand
	restoreLookPath := cmdLookPath
	t.Cleanup(func() {
		osUserHomeDir = restoreHome
		runCommand = restoreCommand
		cmdLookPath = restoreLookPath
	})
	cmdLookPath = missingBinaryLookPath

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
	restoreLookPath := cmdLookPath
	t.Cleanup(func() {
		osUserHomeDir = restoreHome
		runCommand = restoreCommand
		cmdLookPath = restoreLookPath
	})

	osUserHomeDir = func() (string, error) { return home, nil }
	cmdLookPath = missingBinaryLookPath
	recorder := &commandRecorder{}
	runCommand = recorder.record

	// Set the agent adapter's lookPath to simulate missing opencode
	opencodeAdapterLookPath := opencode.LookPathOverride
	opencode.LookPathOverride = missingBinaryLookPath
	t.Cleanup(func() {
		opencode.LookPathOverride = opencodeAdapterLookPath
	})

	detection := linuxDetectionResult(system.LinuxDistroUbuntu, "apt")
	_, err := RunInstall(
		[]string{"--agent", "opencode", "--component", "permissions"},
		detection,
	)
	if err != nil {
		t.Fatalf("RunInstall() error = %v", err)
	}

	// OpenCode on Ubuntu should resolve via npm install (official method from opencode.ai).
	commands := recorder.get()
	foundNpmInstall := false
	for _, cmd := range commands {
		if strings.Contains(cmd, "sudo npm install -g opencode-ai") {
			foundNpmInstall = true
			break
		}
	}
	if !foundNpmInstall {
		t.Fatalf("expected npm install command for opencode agent, got commands: %v", commands)
	}
}

// --- Batch E: Linux verification and macOS parity matrix ---

func TestRunInstallLinuxVerificationReportsReadyOnSuccess(t *testing.T) {
	home := t.TempDir()
	restoreHome := osUserHomeDir
	restoreCommand := runCommand
	restoreLookPath := cmdLookPath
	t.Cleanup(func() {
		osUserHomeDir = restoreHome
		runCommand = restoreCommand
		cmdLookPath = restoreLookPath
	})

	osUserHomeDir = func() (string, error) { return home, nil }
	runCommand = func(string, ...string) error { return nil }
	cmdLookPath = missingBinaryLookPath

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
	restoreLookPath := cmdLookPath
	t.Cleanup(func() {
		osUserHomeDir = restoreHome
		runCommand = restoreCommand
		cmdLookPath = restoreLookPath
	})

	osUserHomeDir = func() (string, error) { return home, nil }
	runCommand = func(string, ...string) error { return nil }
	cmdLookPath = missingBinaryLookPath

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
	restoreLookPath := cmdLookPath
	t.Cleanup(func() {
		osUserHomeDir = restoreHome
		runCommand = restoreCommand
		cmdLookPath = restoreLookPath
	})

	osUserHomeDir = func() (string, error) { return home, nil }
	cmdLookPath = missingBinaryLookPath
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
	restoreLookPath := cmdLookPath
	t.Cleanup(func() {
		osUserHomeDir = restoreHome
		runCommand = restoreCommand
		cmdLookPath = restoreLookPath
	})

	osUserHomeDir = func() (string, error) { return home, nil }
	runCommand = func(string, ...string) error { return nil }
	cmdLookPath = missingBinaryLookPath

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
	settingsPath := filepath.Join(home, ".config", "opencode", "opencode.json")
	if err := os.MkdirAll(filepath.Dir(settingsPath), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	before := []byte("{\n  \"macos-original\": true\n}\n")
	if err := os.WriteFile(settingsPath, before, 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	restoreHome := osUserHomeDir
	restoreCommand := runCommand
	restoreLookPath := cmdLookPath
	t.Cleanup(func() {
		osUserHomeDir = restoreHome
		runCommand = restoreCommand
		cmdLookPath = restoreLookPath
	})
	cmdLookPath = missingBinaryLookPath

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

// --- Skip-when-installed and Go auto-install tests ---

func TestRunInstallEngramSkipsInstallWhenAlreadyOnPath(t *testing.T) {
	home := t.TempDir()
	restoreHome := osUserHomeDir
	restoreCommand := runCommand
	restoreLookPath := cmdLookPath
	t.Cleanup(func() {
		osUserHomeDir = restoreHome
		runCommand = restoreCommand
		cmdLookPath = restoreLookPath
	})

	osUserHomeDir = func() (string, error) { return home, nil }
	// Simulate engram already installed on PATH.
	cmdLookPath = func(name string) (string, error) {
		return "/usr/local/bin/" + name, nil
	}
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
		t.Fatalf("verification ready = false")
	}

	// No brew/go install commands should have been recorded — only agent install.
	for _, cmd := range recorder.get() {
		if strings.Contains(cmd, "brew install engram") || (strings.Contains(cmd, "go install") && strings.Contains(cmd, "engram")) {
			t.Fatalf("expected engram install to be skipped, but got command: %s", cmd)
		}
	}
}

func TestRunInstallEngramAttemptsOpenCodeSetupWhenBinaryPresent(t *testing.T) {
	home := t.TempDir()
	restoreHome := osUserHomeDir
	restoreCommand := runCommand
	restoreLookPath := cmdLookPath
	t.Cleanup(func() {
		osUserHomeDir = restoreHome
		runCommand = restoreCommand
		cmdLookPath = restoreLookPath
	})

	osUserHomeDir = func() (string, error) { return home, nil }
	cmdLookPath = func(name string) (string, error) {
		return "/usr/local/bin/" + name, nil
	}
	recorder := &commandRecorder{}
	runCommand = recorder.record

	result, err := RunInstall(
		[]string{"--agent", "opencode", "--component", "engram"},
		macOSDetectionResult(),
	)
	if err != nil {
		t.Fatalf("RunInstall() error = %v", err)
	}
	if !result.Verify.Ready {
		t.Fatalf("verification ready = false")
	}

	commands := recorder.get()
	foundSetup := false
	for _, cmd := range commands {
		if strings.Contains(cmd, "engram setup opencode") {
			foundSetup = true
			break
		}
	}
	if !foundSetup {
		t.Fatalf("expected engram setup command, got commands: %v", commands)
	}
}

func TestRunInstallEngramFallsBackToInjectWhenSetupFails(t *testing.T) {
	home := t.TempDir()
	restoreHome := osUserHomeDir
	restoreCommand := runCommand
	restoreLookPath := cmdLookPath
	t.Cleanup(func() {
		osUserHomeDir = restoreHome
		runCommand = restoreCommand
		cmdLookPath = restoreLookPath
	})

	osUserHomeDir = func() (string, error) { return home, nil }
	cmdLookPath = func(name string) (string, error) {
		return "/usr/local/bin/" + name, nil
	}
	runCommand = func(name string, args ...string) error {
		if name == "engram" && len(args) == 2 && args[0] == "setup" && args[1] == "opencode" {
			return errors.New("setup failed")
		}
		return nil
	}

	result, err := RunInstall(
		[]string{"--agent", "opencode", "--component", "engram"},
		macOSDetectionResult(),
	)
	if err != nil {
		t.Fatalf("RunInstall() error = %v", err)
	}
	if !result.Verify.Ready {
		t.Fatalf("verification ready = false")
	}

	configPath := filepath.Join(home, ".config", "opencode", "opencode.json")
	if _, err := os.Stat(configPath); err != nil {
		t.Fatalf("expected fallback inject to create %q: %v", configPath, err)
	}
}

func TestRunInstallEngramSetupStrictFailsWhenSetupFails(t *testing.T) {
	t.Setenv("GENTLE_AI_ENGRAM_SETUP_STRICT", "1")

	home := t.TempDir()
	restoreHome := osUserHomeDir
	restoreCommand := runCommand
	restoreLookPath := cmdLookPath
	t.Cleanup(func() {
		osUserHomeDir = restoreHome
		runCommand = restoreCommand
		cmdLookPath = restoreLookPath
	})

	osUserHomeDir = func() (string, error) { return home, nil }
	cmdLookPath = func(name string) (string, error) {
		return "/usr/local/bin/" + name, nil
	}
	runCommand = func(name string, args ...string) error {
		if name == "engram" && len(args) == 2 && args[0] == "setup" && args[1] == "opencode" {
			return errors.New("setup failed")
		}
		return nil
	}

	_, err := RunInstall(
		[]string{"--agent", "opencode", "--component", "engram"},
		macOSDetectionResult(),
	)
	if err == nil {
		t.Fatalf("RunInstall() expected error in strict setup mode")
	}
	if !strings.Contains(err.Error(), "engram setup for \"opencode\"") {
		t.Fatalf("RunInstall() error = %v, want setup error", err)
	}
}

func TestRunInstallEngramDefaultModeAttemptsClaudeSetup(t *testing.T) {
	home := t.TempDir()
	restoreHome := osUserHomeDir
	restoreCommand := runCommand
	restoreLookPath := cmdLookPath
	t.Cleanup(func() {
		osUserHomeDir = restoreHome
		runCommand = restoreCommand
		cmdLookPath = restoreLookPath
	})

	osUserHomeDir = func() (string, error) { return home, nil }
	cmdLookPath = func(name string) (string, error) {
		return "/usr/local/bin/" + name, nil
	}
	recorder := &commandRecorder{}
	runCommand = recorder.record

	result, err := RunInstall(
		[]string{"--agent", "claude-code", "--component", "engram"},
		macOSDetectionResult(),
	)
	if err != nil {
		t.Fatalf("RunInstall() error = %v", err)
	}
	if !result.Verify.Ready {
		t.Fatalf("verification ready = false")
	}

	commands := recorder.get()
	foundSetup := false
	for _, cmd := range commands {
		if strings.Contains(cmd, "engram setup claude-code") {
			foundSetup = true
			break
		}
	}
	if !foundSetup {
		t.Fatalf("expected default setup mode to attempt claude-code setup, got commands: %v", commands)
	}
}

func TestRunInstallGGASkipsInstallWhenAlreadyOnPath(t *testing.T) {
	home := t.TempDir()
	restoreHome := osUserHomeDir
	restoreCommand := runCommand
	restoreLookPath := cmdLookPath
	t.Cleanup(func() {
		osUserHomeDir = restoreHome
		runCommand = restoreCommand
		cmdLookPath = restoreLookPath
	})

	osUserHomeDir = func() (string, error) { return home, nil }
	cmdLookPath = func(name string) (string, error) {
		return "/usr/local/bin/" + name, nil
	}
	recorder := &commandRecorder{}
	runCommand = recorder.record

	detection := macOSDetectionResult()
	result, err := RunInstall(
		[]string{"--agent", "opencode", "--component", "gga"},
		detection,
	)
	if err != nil {
		t.Fatalf("RunInstall() error = %v", err)
	}

	if !result.Verify.Ready {
		t.Fatalf("verification ready = false")
	}

	// No brew/git clone commands for GGA should have been recorded.
	for _, cmd := range recorder.get() {
		if strings.Contains(cmd, "gga") || strings.Contains(cmd, "gentleman-guardian-angel") {
			t.Fatalf("expected gga install to be skipped, but got command: %s", cmd)
		}
	}

	prModePath := filepath.Join(home, ".local", "share", "gga", "lib", "pr_mode.sh")
	content, err := os.ReadFile(prModePath)
	if err != nil {
		t.Fatalf("expected gga runtime asset at %q: %v", prModePath, err)
	}
	if !strings.Contains(string(content), "detect_base_branch") {
		t.Fatalf("expected pr_mode.sh to contain detect_base_branch")
	}
}

func TestRunInstallGGALinuxIncludesTempCleanupBeforeClone(t *testing.T) {
	home := t.TempDir()
	restoreHome := osUserHomeDir
	restoreCommand := runCommand
	restoreLookPath := cmdLookPath
	t.Cleanup(func() {
		osUserHomeDir = restoreHome
		runCommand = restoreCommand
		cmdLookPath = restoreLookPath
	})

	osUserHomeDir = func() (string, error) { return home, nil }
	cmdLookPath = func(name string) (string, error) {
		if name == "gga" {
			return "", exec.ErrNotFound
		}
		return "/usr/local/bin/" + name, nil
	}
	recorder := &commandRecorder{}
	runCommand = recorder.record

	result, err := RunInstall(
		[]string{"--agent", "opencode", "--component", "gga"},
		linuxDetectionResult(system.LinuxDistroUbuntu, "apt"),
	)
	if err != nil {
		t.Fatalf("RunInstall() error = %v", err)
	}
	if !result.Verify.Ready {
		t.Fatalf("verification ready = false")
	}

	commands := recorder.get()
	cleanupIdx := -1
	cloneIdx := -1
	for i, cmd := range commands {
		if strings.Contains(cmd, "rm -rf /tmp/gentleman-guardian-angel") {
			cleanupIdx = i
		}
		if strings.Contains(cmd, "git clone https://github.com/Gentleman-Programming/gentleman-guardian-angel.git /tmp/gentleman-guardian-angel") {
			cloneIdx = i
		}
	}

	for _, cmd := range commands {
		if strings.Contains(cmd, "gga install") || strings.Contains(cmd, "gga init") {
			t.Fatalf("expected global gga provisioning only, got repo-level command: %s", cmd)
		}
	}

	if cleanupIdx == -1 {
		t.Fatalf("expected cleanup command before clone, got commands: %v", commands)
	}
	if cloneIdx == -1 {
		t.Fatalf("expected clone command, got commands: %v", commands)
	}
	if cleanupIdx >= cloneIdx {
		t.Fatalf("cleanup should run before clone (cleanup=%d clone=%d)", cleanupIdx, cloneIdx)
	}
}

func TestRunInstallEngramAutoInstallsGoWhenMissing(t *testing.T) {
	home := t.TempDir()
	restoreHome := osUserHomeDir
	restoreCommand := runCommand
	restoreLookPath := cmdLookPath
	t.Cleanup(func() {
		osUserHomeDir = restoreHome
		runCommand = restoreCommand
		cmdLookPath = restoreLookPath
	})

	osUserHomeDir = func() (string, error) { return home, nil }
	// Simulate: engram missing, Go missing.
	cmdLookPath = func(string) (string, error) {
		return "", exec.ErrNotFound
	}
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
		t.Fatalf("verification ready = false")
	}

	commands := recorder.get()
	foundGoInstall := false
	foundEngramInstall := false
	goInstallIdx := -1
	engramInstallIdx := -1
	for i, cmd := range commands {
		if strings.Contains(cmd, "apt-get install -y golang") {
			foundGoInstall = true
			goInstallIdx = i
		}
		if strings.Contains(cmd, "go install") && strings.Contains(cmd, "engram") {
			foundEngramInstall = true
			engramInstallIdx = i
		}
	}

	if !foundGoInstall {
		t.Fatalf("expected Go auto-install command, got commands: %v", commands)
	}
	if !foundEngramInstall {
		t.Fatalf("expected engram install command, got commands: %v", commands)
	}
	if goInstallIdx >= engramInstallIdx {
		t.Fatalf("Go install (idx=%d) should run before engram install (idx=%d)", goInstallIdx, engramInstallIdx)
	}
}

func TestRunInstallEngramSkipsGoInstallWhenGoPresent(t *testing.T) {
	home := t.TempDir()
	restoreHome := osUserHomeDir
	restoreCommand := runCommand
	restoreLookPath := cmdLookPath
	t.Cleanup(func() {
		osUserHomeDir = restoreHome
		runCommand = restoreCommand
		cmdLookPath = restoreLookPath
	})

	osUserHomeDir = func() (string, error) { return home, nil }
	// Simulate: engram missing, Go available.
	cmdLookPath = missingBinaryLookPath
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
		t.Fatalf("verification ready = false")
	}

	// Go install should NOT be in the recorded commands.
	for _, cmd := range recorder.get() {
		if strings.Contains(cmd, "apt-get install -y golang") {
			t.Fatalf("Go should not be installed when already on PATH, got command: %s", cmd)
		}
	}
}

func TestRunInstallEngramBrewSkipsGoCheck(t *testing.T) {
	home := t.TempDir()
	restoreHome := osUserHomeDir
	restoreCommand := runCommand
	restoreLookPath := cmdLookPath
	t.Cleanup(func() {
		osUserHomeDir = restoreHome
		runCommand = restoreCommand
		cmdLookPath = restoreLookPath
	})

	osUserHomeDir = func() (string, error) { return home, nil }
	// Simulate: engram missing, Go missing — but brew platform, so no Go needed.
	cmdLookPath = func(string) (string, error) {
		return "", exec.ErrNotFound
	}
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
		t.Fatalf("verification ready = false")
	}

	// Should use brew install, NOT go install, and no Go auto-install.
	commands := recorder.get()
	for _, cmd := range commands {
		if strings.Contains(cmd, "golang") || strings.Contains(cmd, "apt-get") {
			t.Fatalf("brew platform should not install Go, got command: %s", cmd)
		}
	}

	foundBrew := false
	for _, cmd := range commands {
		if strings.Contains(cmd, "brew install engram") {
			foundBrew = true
		}
	}
	if !foundBrew {
		t.Fatalf("expected brew install engram, got commands: %v", commands)
	}
}

func TestEnsureGoAvailableAfterInstallWindowsRefreshesPath(t *testing.T) {
	restoreLookPath := cmdLookPath
	restoreStat := osStat
	restoreSetenv := osSetenv
	oldPath := os.Getenv("PATH")
	oldProgramFiles := os.Getenv("ProgramFiles")
	t.Cleanup(func() {
		cmdLookPath = restoreLookPath
		osStat = restoreStat
		osSetenv = restoreSetenv
		_ = os.Setenv("PATH", oldPath)
		_ = os.Setenv("ProgramFiles", oldProgramFiles)
	})

	programFiles := `C:\Program Files`
	if err := os.Setenv("ProgramFiles", programFiles); err != nil {
		t.Fatalf("Setenv(ProgramFiles) error = %v", err)
	}
	if err := os.Setenv("PATH", `C:\Windows\System32`); err != nil {
		t.Fatalf("Setenv(PATH) error = %v", err)
	}

	cmdLookPath = func(name string) (string, error) {
		if name == "go" {
			return "", exec.ErrNotFound
		}
		return name, nil
	}
	osStat = func(name string) (os.FileInfo, error) {
		want := filepath.Join(programFiles, "Go", "bin", "go.exe")
		if name == want {
			return fakeFileInfo{name: "go.exe"}, nil
		}
		return nil, os.ErrNotExist
	}
	osSetenv = os.Setenv

	if err := ensureGoAvailableAfterInstall(system.PlatformProfile{OS: "windows", PackageManager: "winget"}); err != nil {
		t.Fatalf("ensureGoAvailableAfterInstall() error = %v", err)
	}

	updatedPath := os.Getenv("PATH")
	expectedPrefix := filepath.Join(programFiles, "Go", "bin") + string(os.PathListSeparator)
	if !strings.HasPrefix(updatedPath, expectedPrefix) {
		t.Fatalf("PATH = %q, want prefix %q", updatedPath, expectedPrefix)
	}
}

type fakeFileInfo struct{ name string }

func (f fakeFileInfo) Name() string     { return f.name }
func (fakeFileInfo) Size() int64        { return 0 }
func (fakeFileInfo) Mode() os.FileMode  { return 0 }
func (fakeFileInfo) ModTime() time.Time { return time.Time{} }
func (fakeFileInfo) IsDir() bool        { return false }
func (fakeFileInfo) Sys() any           { return nil }
