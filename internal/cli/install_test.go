package cli

import (
	"reflect"
	"testing"

	"github.com/gentleman-programming/gentle-ai/internal/model"
	"github.com/gentleman-programming/gentle-ai/internal/system"
)

func TestParseInstallFlagsSupportsCSVAndRepeated(t *testing.T) {
	flags, err := ParseInstallFlags([]string{
		"--agent", "claude-code,opencode",
		"--agent", "cursor",
		"--component", "engram,sdd",
		"--component", "skills",
		"--skill", "sdd-apply",
		"--persona", "neutral",
		"--preset", "minimal",
		"--dry-run",
	})
	if err != nil {
		t.Fatalf("ParseInstallFlags() error = %v", err)
	}

	if !reflect.DeepEqual(flags.Agents, []string{"claude-code", "opencode", "cursor"}) {
		t.Fatalf("agents = %v", flags.Agents)
	}

	if !reflect.DeepEqual(flags.Components, []string{"engram", "sdd", "skills"}) {
		t.Fatalf("components = %v", flags.Components)
	}

	if !flags.DryRun {
		t.Fatalf("DryRun = false, want true")
	}
}

func TestNormalizeInstallFlagsDefaults(t *testing.T) {
	input, err := NormalizeInstallFlags(InstallFlags{}, system.DetectionResult{})
	if err != nil {
		t.Fatalf("NormalizeInstallFlags() error = %v", err)
	}

	want := model.Selection{
		Agents:  []model.AgentID{model.AgentClaudeCode, model.AgentOpenCode, model.AgentGeminiCLI, model.AgentCursor, model.AgentVSCodeCopilot},
		Persona: model.PersonaGentleman,
		Preset:  model.PresetFullGentleman,
		Components: []model.ComponentID{
			model.ComponentEngram,
			model.ComponentSDD,
			model.ComponentSkills,
			model.ComponentContext7,
			model.ComponentPersona,
			model.ComponentPermission,
			model.ComponentGGA,
		},
	}

	if !reflect.DeepEqual(input.Selection, want) {
		t.Fatalf("selection = %#v, want %#v", input.Selection, want)
	}
}

func TestNormalizeInstallFlagsRejectsUnknownPersona(t *testing.T) {
	_, err := NormalizeInstallFlags(InstallFlags{Persona: "wizard"}, system.DetectionResult{})
	if err == nil {
		t.Fatalf("NormalizeInstallFlags() expected error")
	}
}

func TestNormalizeSDDMode(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    model.SDDModeID
		wantErr bool
	}{
		{name: "empty returns zero value", input: "", want: ""},
		{name: "whitespace returns zero value", input: "   ", want: ""},
		{name: "single is valid", input: "single", want: model.SDDModeSingle},
		{name: "multi is valid", input: "multi", want: model.SDDModeMulti},
		{name: "invalid rejected", input: "turbo", wantErr: true},
		{name: "partial invalid", input: "mult", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeSDDMode(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("normalizeSDDMode(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("normalizeSDDMode(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseInstallFlagsSDDMode(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    string
		wantErr bool
	}{
		{
			name: "flag absent defaults to empty",
			args: []string{"--agent", "opencode"},
			want: "",
		},
		{
			name: "flag set to multi",
			args: []string{"--agent", "opencode", "--sdd-mode", "multi"},
			want: "multi",
		},
		{
			name: "flag set to single",
			args: []string{"--agent", "opencode", "--sdd-mode", "single"},
			want: "single",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags, err := ParseInstallFlags(tt.args)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseInstallFlags() error = %v, wantErr %v", err, tt.wantErr)
			}
			if flags.SDDMode != tt.want {
				t.Fatalf("flags.SDDMode = %q, want %q", flags.SDDMode, tt.want)
			}
		})
	}
}

func TestNormalizeInstallFlagsSDDModeMulti(t *testing.T) {
	input, err := NormalizeInstallFlags(
		InstallFlags{SDDMode: "multi"},
		system.DetectionResult{},
	)
	if err != nil {
		t.Fatalf("NormalizeInstallFlags() error = %v", err)
	}
	if input.Selection.SDDMode != model.SDDModeMulti {
		t.Fatalf("SDDMode = %q, want %q", input.Selection.SDDMode, model.SDDModeMulti)
	}
}

func TestNormalizeInstallFlagsSDDModeInvalid(t *testing.T) {
	_, err := NormalizeInstallFlags(
		InstallFlags{SDDMode: "turbo"},
		system.DetectionResult{},
	)
	if err == nil {
		t.Fatal("expected error for invalid sdd-mode")
	}
}

func TestRunInstallDryRunSkipsExecution(t *testing.T) {
	result, err := RunInstall([]string{"--dry-run"}, system.DetectionResult{})
	if err != nil {
		t.Fatalf("RunInstall() error = %v", err)
	}

	if !result.DryRun {
		t.Fatalf("DryRun = false, want true")
	}

	if len(result.Plan.Apply) == 0 {
		t.Fatalf("apply steps = 0, want > 0")
	}

	if len(result.Execution.Apply.Steps) != 0 || len(result.Execution.Prepare.Steps) != 0 {
		t.Fatalf("execution should be empty in dry-run")
	}
}
