package planner

import (
	"reflect"
	"testing"

	"github.com/gentleman-programming/gentleman-ai-installer/internal/model"
	"github.com/gentleman-programming/gentleman-ai-installer/internal/system"
)

func TestBuildReviewPayloadIncludesPlatformDecision(t *testing.T) {
	selection := model.Selection{
		Persona: model.PersonaGentleman,
		Preset:  model.PresetFullGentleman,
	}

	resolved := ResolvedPlan{
		Agents:            []model.AgentID{model.AgentClaudeCode},
		OrderedComponents: []model.ComponentID{model.ComponentEngram},
		PlatformDecision: PlatformDecision{
			OS:             "linux",
			LinuxDistro:    "arch",
			PackageManager: "pacman",
			Supported:      true,
		},
	}

	payload := BuildReviewPayload(selection, resolved)

	if !reflect.DeepEqual(payload.PlatformDecision, resolved.PlatformDecision) {
		t.Fatalf("platform decision = %#v, want %#v", payload.PlatformDecision, resolved.PlatformDecision)
	}
}

// --- Batch E: Platform decision propagation matrix ---

func TestPlatformDecisionFromProfileMatrix(t *testing.T) {
	tests := []struct {
		name    string
		profile system.PlatformProfile
		want    PlatformDecision
	}{
		{
			name: "darwin profile maps to brew decision",
			profile: system.PlatformProfile{
				OS: "darwin", PackageManager: "brew", Supported: true,
			},
			want: PlatformDecision{
				OS: "darwin", PackageManager: "brew", Supported: true,
			},
		},
		{
			name: "ubuntu profile maps to apt decision",
			profile: system.PlatformProfile{
				OS: "linux", LinuxDistro: system.LinuxDistroUbuntu, PackageManager: "apt", Supported: true,
			},
			want: PlatformDecision{
				OS: "linux", LinuxDistro: system.LinuxDistroUbuntu, PackageManager: "apt", Supported: true,
			},
		},
		{
			name: "arch profile maps to pacman decision",
			profile: system.PlatformProfile{
				OS: "linux", LinuxDistro: system.LinuxDistroArch, PackageManager: "pacman", Supported: true,
			},
			want: PlatformDecision{
				OS: "linux", LinuxDistro: system.LinuxDistroArch, PackageManager: "pacman", Supported: true,
			},
		},
		{
			name: "unsupported distro propagates unsupported status",
			profile: system.PlatformProfile{
				OS: "linux", LinuxDistro: "fedora", Supported: false,
			},
			want: PlatformDecision{
				OS: "linux", LinuxDistro: "fedora", Supported: false,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := PlatformDecisionFromProfile(tc.profile)
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("PlatformDecisionFromProfile() = %#v, want %#v", got, tc.want)
			}
		})
	}
}

func TestBuildReviewPayloadPlatformDecisionPropagatesPerProfile(t *testing.T) {
	profiles := []PlatformDecision{
		{OS: "darwin", PackageManager: "brew", Supported: true},
		{OS: "linux", LinuxDistro: "ubuntu", PackageManager: "apt", Supported: true},
		{OS: "linux", LinuxDistro: "arch", PackageManager: "pacman", Supported: true},
	}

	for _, decision := range profiles {
		t.Run(decision.OS+"/"+decision.LinuxDistro, func(t *testing.T) {
			selection := model.Selection{
				Persona: model.PersonaGentleman,
				Preset:  model.PresetFullGentleman,
			}
			resolved := ResolvedPlan{
				Agents:           []model.AgentID{model.AgentClaudeCode},
				PlatformDecision: decision,
			}

			payload := BuildReviewPayload(selection, resolved)
			if !reflect.DeepEqual(payload.PlatformDecision, decision) {
				t.Fatalf("review payload platform decision = %#v, want %#v", payload.PlatformDecision, decision)
			}
		})
	}
}

func TestResolverOutputIsPlatformAgnostic(t *testing.T) {
	// Planner resolver does NOT set PlatformDecision — it is set by CLI after resolve.
	// This test confirms resolver output has zero-value PlatformDecision.
	resolver := NewResolver(MVPGraph())
	selection := model.Selection{
		Agents:     []model.AgentID{model.AgentClaudeCode},
		Components: []model.ComponentID{model.ComponentEngram},
	}

	plan, err := resolver.Resolve(selection)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	zero := PlatformDecision{}
	if plan.PlatformDecision != zero {
		t.Fatalf("resolver should not set PlatformDecision, got %#v", plan.PlatformDecision)
	}
}
