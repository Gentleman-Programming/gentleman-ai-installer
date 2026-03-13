package skillssh

import "testing"

func TestFilter_TrustedSourcesPass(t *testing.T) {
	skills := []SearchSkill{
		{ID: "1", SkillID: "react-hooks", Name: "React Hooks", Installs: 1000, Source: "vercel-labs/agent-skills"},
		{ID: "2", SkillID: "stitch-skill", Name: "Stitch Skill", Installs: 800, Source: "google-labs-code/stitch-skills"},
		{ID: "3", SkillID: "unknown", Name: "Unknown", Installs: 9999, Source: "random-org/random-skills"},
	}

	got := Filter(skills)

	if len(got) != 2 {
		t.Fatalf("expected 2 trusted skills, got %d", len(got))
	}

	for _, s := range got {
		if !isTrusted(s.Source) {
			t.Errorf("untrusted source %q slipped through filter", s.Source)
		}
	}
}

func TestFilter_SortOrderDescendingInstalls(t *testing.T) {
	skills := []SearchSkill{
		{ID: "1", SkillID: "low", Name: "Low", Installs: 600, Source: "vercel-labs/agent-skills"},
		{ID: "2", SkillID: "high", Name: "High", Installs: 5000, Source: "vercel-labs/agent-skills"},
		{ID: "3", SkillID: "mid", Name: "Mid", Installs: 2000, Source: "google-labs-code/stitch-skills"},
	}

	got := Filter(skills)

	if len(got) != 3 {
		t.Fatalf("expected 3 skills, got %d", len(got))
	}

	for i := 1; i < len(got); i++ {
		if got[i].Installs > got[i-1].Installs {
			t.Errorf("sort order wrong at index %d: %d > %d", i, got[i].Installs, got[i-1].Installs)
		}
	}

	if got[0].Name != "High" {
		t.Errorf("first skill = %q, want %q", got[0].Name, "High")
	}
}

func TestFilter_BelowMinInstallsExcluded(t *testing.T) {
	skills := []SearchSkill{
		{ID: "1", SkillID: "popular", Name: "Popular", Installs: 1000, Source: "vercel-labs/agent-skills"},
		{ID: "2", SkillID: "too-few", Name: "TooFew", Installs: 499, Source: "vercel-labs/agent-skills"},
	}

	got := Filter(skills)

	if len(got) != 1 {
		t.Fatalf("expected 1 skill (above threshold), got %d", len(got))
	}
	if got[0].Name != "Popular" {
		t.Errorf("got %q, want Popular", got[0].Name)
	}
}

func TestFilter_EmptyInput(t *testing.T) {
	got := Filter([]SearchSkill{})

	if got == nil {
		t.Fatal("Filter returned nil, want non-nil empty slice")
	}
	if len(got) != 0 {
		t.Fatalf("expected empty slice, got %d", len(got))
	}
}

func TestFilter_AllUntrusted(t *testing.T) {
	skills := []SearchSkill{
		{ID: "1", SkillID: "shady", Name: "Shady", Installs: 9999, Source: "shady-org/shady-skills"},
		{ID: "2", SkillID: "unknown", Name: "Unknown", Installs: 5000, Source: "no-trust/whatever"},
	}

	got := Filter(skills)

	if got == nil {
		t.Fatal("Filter returned nil, want non-nil empty slice")
	}
	if len(got) != 0 {
		t.Fatalf("expected empty slice for all-untrusted input, got %d", len(got))
	}
}

func TestFilter_NilInput(t *testing.T) {
	got := Filter(nil)

	if got == nil {
		t.Fatal("Filter returned nil for nil input, want non-nil empty slice")
	}
	if len(got) != 0 {
		t.Fatalf("expected empty slice, got %d", len(got))
	}
}

func TestFilter_ColonSkillIDExcluded(t *testing.T) {
	skills := []SearchSkill{
		{ID: "1", SkillID: "react:components", Name: "react:components", Installs: 13000, Source: "google-labs-code/stitch-skills"},
		{ID: "2", SkillID: "shadcn-ui", Name: "shadcn-ui", Installs: 5000, Source: "google-labs-code/stitch-skills"},
	}

	got := Filter(skills)

	if len(got) != 1 {
		t.Fatalf("expected 1 skill (colon skillId excluded), got %d", len(got))
	}
	if got[0].SkillID != "shadcn-ui" {
		t.Errorf("got %q, want shadcn-ui", got[0].SkillID)
	}
}
