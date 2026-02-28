package skills

import (
	"testing"

	"github.com/gentleman-programming/gentle-ai/internal/model"
)

func TestSkillsForPresetMinimalReturnsSDDOnly(t *testing.T) {
	skills := SkillsForPreset(model.PresetMinimal)
	if len(skills) == 0 {
		t.Fatalf("SkillsForPreset(minimal) returned empty")
	}

	for _, skill := range skills {
		if len(skill) < 4 || skill[:3] != "sdd" {
			t.Fatalf("minimal preset should only contain SDD skills, got %q", skill)
		}
	}
}

func TestSkillsForPresetEcosystemIncludesFrameworks(t *testing.T) {
	skills := SkillsForPreset(model.PresetEcosystemOnly)

	hasTypescript := false
	hasSDDInit := false
	for _, skill := range skills {
		if skill == model.SkillTypeScript {
			hasTypescript = true
		}
		if skill == model.SkillSDDInit {
			hasSDDInit = true
		}
	}

	if !hasTypescript {
		t.Fatalf("ecosystem preset should include typescript")
	}
	if !hasSDDInit {
		t.Fatalf("ecosystem preset should include sdd-init")
	}
}

func TestSkillsForPresetFullIncludesAll(t *testing.T) {
	skills := SkillsForPreset(model.PresetFullGentleman)
	all := AllSkillIDs()

	if len(skills) != len(all) {
		t.Fatalf("full preset skills len = %d, all skills len = %d", len(skills), len(all))
	}
}

func TestSkillsForPresetCustomReturnsNil(t *testing.T) {
	skills := SkillsForPreset(model.PresetCustom)
	if skills != nil {
		t.Fatalf("custom preset should return nil, got %v", skills)
	}
}

func TestAllSkillIDsIncludesEveryKnownSkill(t *testing.T) {
	all := AllSkillIDs()

	required := []model.SkillID{
		model.SkillSDDInit,
		model.SkillTypeScript,
		model.SkillPlaywright,
		model.SkillGoTesting,
	}

	skillSet := make(map[model.SkillID]struct{}, len(all))
	for _, skill := range all {
		skillSet[skill] = struct{}{}
	}

	for _, req := range required {
		if _, ok := skillSet[req]; !ok {
			t.Fatalf("AllSkillIDs() missing %q", req)
		}
	}
}
