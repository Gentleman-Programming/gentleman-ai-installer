package skills

import "testing"

func TestPresetsContainFullAndMinimal(t *testing.T) {
	presets := Presets()
	if len(presets) != 2 {
		t.Fatalf("Presets() len = %d", len(presets))
	}

	if presets[0].ID != PresetFullStack {
		t.Fatalf("first preset ID = %q", presets[0].ID)
	}

	if presets[1].ID != PresetMinimal {
		t.Fatalf("second preset ID = %q", presets[1].ID)
	}
}
