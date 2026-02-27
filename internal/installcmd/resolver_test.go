package installcmd

import (
	"reflect"
	"testing"

	"github.com/gentleman-programming/gentleman-ai-installer/internal/model"
	"github.com/gentleman-programming/gentleman-ai-installer/internal/system"
)

func TestResolveDependencyInstall(t *testing.T) {
	r := NewResolver()

	tests := []struct {
		name    string
		profile system.PlatformProfile
		want    []string
		wantErr bool
	}{
		{
			name:    "darwin resolves brew command",
			profile: system.PlatformProfile{OS: "darwin", PackageManager: "brew"},
			want:    []string{"brew", "install", "opencode"},
		},
		{
			name:    "ubuntu resolves apt command",
			profile: system.PlatformProfile{OS: "linux", LinuxDistro: system.LinuxDistroUbuntu, PackageManager: "apt"},
			want:    []string{"sudo", "apt-get", "install", "-y", "opencode"},
		},
		{
			name:    "arch resolves pacman command",
			profile: system.PlatformProfile{OS: "linux", LinuxDistro: system.LinuxDistroArch, PackageManager: "pacman"},
			want:    []string{"sudo", "pacman", "-S", "--noconfirm", "opencode"},
		},
		{
			name:    "unsupported package manager returns error",
			profile: system.PlatformProfile{OS: "linux", LinuxDistro: "fedora", PackageManager: "dnf"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			command, err := r.ResolveDependencyInstall(tt.profile, "opencode")
			if (err != nil) != tt.wantErr {
				t.Fatalf("ResolveDependencyInstall() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			if !reflect.DeepEqual(command, tt.want) {
				t.Fatalf("ResolveDependencyInstall() = %v, want %v", command, tt.want)
			}
		})
	}
}

func TestResolveAgentInstall(t *testing.T) {
	r := NewResolver()

	command, err := r.ResolveAgentInstall(system.PlatformProfile{OS: "darwin", PackageManager: "brew"}, model.AgentClaudeCode)
	if err != nil {
		t.Fatalf("ResolveAgentInstall() error = %v", err)
	}
	if !reflect.DeepEqual(command, []string{"npm", "install", "-g", "@anthropic-ai/claude-code"}) {
		t.Fatalf("ResolveAgentInstall() = %v", command)
	}

	command, err = r.ResolveAgentInstall(
		system.PlatformProfile{OS: "linux", LinuxDistro: system.LinuxDistroUbuntu, PackageManager: "apt"},
		model.AgentOpenCode,
	)
	if err != nil {
		t.Fatalf("ResolveAgentInstall() error = %v", err)
	}
	if !reflect.DeepEqual(command, []string{"sudo", "apt-get", "install", "-y", "opencode"}) {
		t.Fatalf("ResolveAgentInstall() = %v", command)
	}
}

func TestResolveComponentInstall(t *testing.T) {
	r := NewResolver()

	command, err := r.ResolveComponentInstall(
		system.PlatformProfile{OS: "linux", LinuxDistro: system.LinuxDistroArch, PackageManager: "pacman"},
		model.ComponentEngram,
	)
	if err != nil {
		t.Fatalf("ResolveComponentInstall() error = %v", err)
	}
	if !reflect.DeepEqual(command, []string{"sudo", "pacman", "-S", "--noconfirm", "engram"}) {
		t.Fatalf("ResolveComponentInstall() = %v", command)
	}
}
