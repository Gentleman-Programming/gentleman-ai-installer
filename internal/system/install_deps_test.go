package system

import (
	"strings"
	"testing"
)

func TestInstallHintGitDarwin(t *testing.T) {
	profile := PlatformProfile{OS: "darwin", PackageManager: "brew"}
	hint := installHintGit(profile)
	if hint != "brew install git" {
		t.Fatalf("installHintGit(darwin) = %q", hint)
	}
}

func TestInstallHintGitUbuntu(t *testing.T) {
	profile := PlatformProfile{OS: "linux", PackageManager: "apt", LinuxDistro: "ubuntu"}
	hint := installHintGit(profile)
	if !strings.Contains(hint, "apt-get install") {
		t.Fatalf("installHintGit(ubuntu) = %q", hint)
	}
}

func TestInstallHintGitArch(t *testing.T) {
	profile := PlatformProfile{OS: "linux", PackageManager: "pacman", LinuxDistro: "arch"}
	hint := installHintGit(profile)
	if !strings.Contains(hint, "pacman -S") {
		t.Fatalf("installHintGit(arch) = %q", hint)
	}
}

func TestInstallHintNodeDarwin(t *testing.T) {
	profile := PlatformProfile{OS: "darwin", PackageManager: "brew"}
	hint := installHintNode(profile)
	if hint != "brew install node" {
		t.Fatalf("installHintNode(darwin) = %q", hint)
	}
}

func TestInstallHintNodeUbuntu(t *testing.T) {
	profile := PlatformProfile{OS: "linux", PackageManager: "apt", LinuxDistro: "ubuntu"}
	hint := installHintNode(profile)
	if !strings.Contains(hint, "nodesource") {
		t.Fatalf("installHintNode(ubuntu) = %q, want NodeSource URL", hint)
	}
}

func TestInstallHintNodeArch(t *testing.T) {
	profile := PlatformProfile{OS: "linux", PackageManager: "pacman", LinuxDistro: "arch"}
	hint := installHintNode(profile)
	if !strings.Contains(hint, "pacman") || !strings.Contains(hint, "nodejs") {
		t.Fatalf("installHintNode(arch) = %q", hint)
	}
}

func TestInstallHintBrew(t *testing.T) {
	hint := installHintBrew()
	if !strings.Contains(hint, "Homebrew") {
		t.Fatalf("installHintBrew() = %q, want Homebrew URL", hint)
	}
}

func TestInstallHintGoDarwin(t *testing.T) {
	profile := PlatformProfile{OS: "darwin", PackageManager: "brew"}
	hint := installHintGo(profile)
	if hint != "brew install go" {
		t.Fatalf("installHintGo(darwin) = %q", hint)
	}
}

func TestInstallHintGoUbuntu(t *testing.T) {
	profile := PlatformProfile{OS: "linux", PackageManager: "apt", LinuxDistro: "ubuntu"}
	hint := installHintGo(profile)
	if !strings.Contains(hint, "apt-get install") {
		t.Fatalf("installHintGo(ubuntu) = %q", hint)
	}
}

func TestInstallCommandsForDepGitDarwin(t *testing.T) {
	profile := PlatformProfile{OS: "darwin", PackageManager: "brew"}
	cmds := InstallCommandsForDep("git", profile)
	if len(cmds) != 1 {
		t.Fatalf("git darwin commands = %d, want 1", len(cmds))
	}
	if cmds[0][0] != "brew" || cmds[0][1] != "install" || cmds[0][2] != "git" {
		t.Fatalf("git darwin command = %v", cmds[0])
	}
}

func TestInstallCommandsForDepGitUbuntu(t *testing.T) {
	profile := PlatformProfile{OS: "linux", PackageManager: "apt", LinuxDistro: "ubuntu"}
	cmds := InstallCommandsForDep("git", profile)
	if len(cmds) != 1 {
		t.Fatalf("git ubuntu commands = %d, want 1", len(cmds))
	}
	if cmds[0][0] != "sudo" {
		t.Fatalf("git ubuntu command = %v, want sudo", cmds[0])
	}
}

func TestInstallCommandsForDepNodeUbuntuHasTwoSteps(t *testing.T) {
	profile := PlatformProfile{OS: "linux", PackageManager: "apt", LinuxDistro: "ubuntu"}
	cmds := InstallCommandsForDep("node", profile)
	if len(cmds) != 2 {
		t.Fatalf("node ubuntu commands = %d, want 2 (nodesource setup + install)", len(cmds))
	}
}

func TestInstallCommandsForDepNpmReturnsNil(t *testing.T) {
	profile := PlatformProfile{OS: "darwin", PackageManager: "brew"}
	cmds := InstallCommandsForDep("npm", profile)
	if cmds != nil {
		t.Fatalf("npm commands = %v, want nil (comes with node)", cmds)
	}
}

func TestInstallCommandsForDepBrewOnLinuxReturnsNil(t *testing.T) {
	profile := PlatformProfile{OS: "linux", PackageManager: "apt"}
	cmds := InstallCommandsForDep("brew", profile)
	if cmds != nil {
		t.Fatalf("brew on linux = %v, want nil", cmds)
	}
}

func TestInstallCommandsForDepBrewOnDarwin(t *testing.T) {
	profile := PlatformProfile{OS: "darwin", PackageManager: "brew"}
	cmds := InstallCommandsForDep("brew", profile)
	if len(cmds) != 1 {
		t.Fatalf("brew darwin commands = %d, want 1", len(cmds))
	}
}

func TestInstallCommandsForDepUnknownReturnsNil(t *testing.T) {
	profile := PlatformProfile{OS: "darwin", PackageManager: "brew"}
	cmds := InstallCommandsForDep("unknown_tool", profile)
	if cmds != nil {
		t.Fatalf("unknown tool commands = %v, want nil", cmds)
	}
}

func TestInstallCommandsForDepGitArchUsesPacman(t *testing.T) {
	profile := PlatformProfile{OS: "linux", PackageManager: "pacman", LinuxDistro: "arch"}
	cmds := InstallCommandsForDep("git", profile)
	if len(cmds) != 1 {
		t.Fatalf("git arch commands = %d, want 1", len(cmds))
	}
	if cmds[0][0] != "sudo" || cmds[0][1] != "pacman" {
		t.Fatalf("git arch command = %v, want sudo pacman", cmds[0])
	}
}

func TestFormatMissingDepsMessageAllPresent(t *testing.T) {
	report := DependencyReport{AllPresent: true}
	msg := FormatMissingDepsMessage(report)
	if !strings.Contains(msg, "All required dependencies are present") {
		t.Fatalf("unexpected message for all present: %q", msg)
	}
}

func TestFormatMissingDepsMessageWithMissing(t *testing.T) {
	report := DependencyReport{
		Dependencies: []Dependency{
			{Name: "node", Required: true, Installed: false, InstallHint: "brew install node"},
		},
		AllPresent:      false,
		MissingRequired: []string{"node"},
	}

	msg := FormatMissingDepsMessage(report)
	if !strings.Contains(msg, "node") || !strings.Contains(msg, "brew install node") {
		t.Fatalf("missing deps message = %q", msg)
	}
}

func TestInstallCommandsFullMatrix(t *testing.T) {
	profiles := []PlatformProfile{
		{OS: "darwin", PackageManager: "brew"},
		{OS: "linux", PackageManager: "apt", LinuxDistro: "ubuntu"},
		{OS: "linux", PackageManager: "pacman", LinuxDistro: "arch"},
	}

	deps := []string{"git", "curl", "node", "go"}

	for _, profile := range profiles {
		for _, dep := range deps {
			t.Run(profile.OS+"/"+profile.PackageManager+"/"+dep, func(t *testing.T) {
				cmds := InstallCommandsForDep(dep, profile)
				if cmds == nil {
					t.Fatalf("InstallCommandsForDep(%q, %s/%s) = nil", dep, profile.OS, profile.PackageManager)
				}
				if len(cmds) == 0 {
					t.Fatalf("InstallCommandsForDep(%q) returned empty slice", dep)
				}
				for _, cmd := range cmds {
					if len(cmd) == 0 {
						t.Fatalf("empty command in sequence for %q", dep)
					}
				}
			})
		}
	}
}
