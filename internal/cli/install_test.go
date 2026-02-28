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
		Agents:  []model.AgentID{model.AgentClaudeCode, model.AgentOpenCode},
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
