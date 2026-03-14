package screens

import (
	"strings"
	"testing"
)

func TestRenderCompleteSuccessShowsGGANotesWhenInstalled(t *testing.T) {
	out := RenderComplete(CompletePayload{
		ConfiguredAgents:    1,
		InstalledComponents: 1,
		GGAInstalled:        true,
	})

	if !strings.Contains(out, "GGA (per project)") {
		t.Fatalf("missing GGA section: %q", out)
	}
	if !strings.Contains(out, "gga init") || !strings.Contains(out, "gga install") {
		t.Fatalf("missing GGA repo commands: %q", out)
	}
}

func TestRenderCompleteSuccessHidesGGANotesWhenNotInstalled(t *testing.T) {
	out := RenderComplete(CompletePayload{
		ConfiguredAgents:    1,
		InstalledComponents: 1,
		GGAInstalled:        false,
	})

	if strings.Contains(out, "GGA (per project)") {
		t.Fatalf("unexpected GGA section: %q", out)
	}
}
