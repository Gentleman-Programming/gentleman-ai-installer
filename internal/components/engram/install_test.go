package engram

import (
	"reflect"
	"testing"

	"github.com/gentleman-programming/gentleman-ai-installer/internal/model"
	"github.com/gentleman-programming/gentleman-ai-installer/internal/system"
)

func TestInstallCommandByProfile(t *testing.T) {
	tests := []struct {
		name    string
		profile system.PlatformProfile
		want    []string
		wantErr bool
	}{
		{
			name:    "darwin uses brew",
			profile: system.PlatformProfile{OS: "darwin", PackageManager: "brew"},
			want:    []string{"brew", "install", "engram"},
		},
		{
			name:    "ubuntu uses apt",
			profile: system.PlatformProfile{OS: "linux", LinuxDistro: system.LinuxDistroUbuntu, PackageManager: "apt"},
			want:    []string{"sudo", "apt-get", "install", "-y", "engram"},
		},
		{
			name:    "arch uses pacman",
			profile: system.PlatformProfile{OS: "linux", LinuxDistro: system.LinuxDistroArch, PackageManager: "pacman"},
			want:    []string{"sudo", "pacman", "-S", "--noconfirm", "engram"},
		},
		{
			name:    "unsupported package manager errors",
			profile: system.PlatformProfile{OS: "linux", LinuxDistro: "fedora", PackageManager: "dnf"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			command, err := InstallCommand(tt.profile)
			if (err != nil) != tt.wantErr {
				t.Fatalf("InstallCommand() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			if !reflect.DeepEqual(command, tt.want) {
				t.Fatalf("InstallCommand() = %v, want %v", command, tt.want)
			}
		})
	}
}

func TestValidateAgents(t *testing.T) {
	if err := ValidateAgents([]model.AgentID{model.AgentClaudeCode, model.AgentOpenCode}); err != nil {
		t.Fatalf("ValidateAgents() error = %v", err)
	}

	if err := ValidateAgents([]model.AgentID{model.AgentID("cursor")}); err == nil {
		t.Fatalf("ValidateAgents() expected error for unsupported agent")
	}
}
