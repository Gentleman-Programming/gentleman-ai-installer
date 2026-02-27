package verify

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestScenarioReadyWhenAllChecksPass(t *testing.T) {
	checks := []Check{
		{
			ID:          "engram-health",
			Description: "Engram health endpoint responds",
			Run: func(context.Context) error {
				return nil
			},
		},
		{
			ID:          "skills-path",
			Description: "Skills are available in configured path",
			Run: func(context.Context) error {
				return nil
			},
		},
	}

	report := BuildReport(RunChecks(context.Background(), checks))
	if !report.Ready {
		t.Fatalf("Ready = false, want true")
	}

	if report.Failed != 0 {
		t.Fatalf("Failed = %d, want 0", report.Failed)
	}

	if report.FinalNote != readyMessage {
		t.Fatalf("FinalNote = %q", report.FinalNote)
	}
}

func TestScenarioNotReadyWhenAnyCheckFails(t *testing.T) {
	checks := []Check{
		{
			ID: "engram-health",
			Run: func(context.Context) error {
				return errors.New("503 service unavailable")
			},
		},
		{
			ID: "mcp-tools",
			Run: func(context.Context) error {
				return nil
			},
		},
	}

	report := BuildReport(RunChecks(context.Background(), checks))
	if report.Ready {
		t.Fatalf("Ready = true, want false")
	}

	if report.Failed != 1 {
		t.Fatalf("Failed = %d, want 1", report.Failed)
	}

	rendered := RenderReport(report)
	if !strings.Contains(rendered, "[!!] engram-health") {
		t.Fatalf("RenderReport() missing failed check line: %q", rendered)
	}

	if !strings.Contains(rendered, "verification issues") {
		t.Fatalf("RenderReport() missing failure final note: %q", rendered)
	}
}

func TestScenarioSkippedCheckIsReported(t *testing.T) {
	report := BuildReport(RunChecks(context.Background(), []Check{{ID: "future-check"}}))

	if report.Skipped != 1 {
		t.Fatalf("Skipped = %d, want 1", report.Skipped)
	}

	if !report.Ready {
		t.Fatalf("Ready = false, want true when no failures")
	}
}

// --- Batch E: Platform-aware verification checks ---

func TestScenarioMixedPassFailSkipCounts(t *testing.T) {
	checks := []Check{
		{ID: "pass-1", Run: func(context.Context) error { return nil }},
		{ID: "fail-1", Run: func(context.Context) error { return errors.New("broken") }},
		{ID: "skip-1"}, // nil Run → skipped
		{ID: "pass-2", Run: func(context.Context) error { return nil }},
	}

	report := BuildReport(RunChecks(context.Background(), checks))

	if report.Passed != 2 {
		t.Fatalf("Passed = %d, want 2", report.Passed)
	}
	if report.Failed != 1 {
		t.Fatalf("Failed = %d, want 1", report.Failed)
	}
	if report.Skipped != 1 {
		t.Fatalf("Skipped = %d, want 1", report.Skipped)
	}
	if report.Ready {
		t.Fatalf("Ready = true, want false when failures present")
	}
}

func TestRenderReportFormatsAllStatuses(t *testing.T) {
	results := []CheckResult{
		{ID: "check-ok", Description: "passes", Status: CheckStatusPassed},
		{ID: "check-fail", Description: "fails", Status: CheckStatusFailed, Error: "timeout"},
		{ID: "check-skip", Description: "skipped", Status: CheckStatusSkipped, Error: "check not implemented"},
	}

	report := BuildReport(results)
	rendered := RenderReport(report)

	if !strings.Contains(rendered, "[ok] check-ok - passes") {
		t.Fatalf("missing passed line in: %s", rendered)
	}
	if !strings.Contains(rendered, "[!!] check-fail - fails (timeout)") {
		t.Fatalf("missing failed line in: %s", rendered)
	}
	if !strings.Contains(rendered, "[--] check-skip - skipped (check not implemented)") {
		t.Fatalf("missing skipped line in: %s", rendered)
	}
}

func TestEmptyCheckListIsReady(t *testing.T) {
	report := BuildReport(RunChecks(context.Background(), nil))

	if !report.Ready {
		t.Fatalf("Ready = false, want true for empty check list")
	}
	if report.FinalNote != readyMessage {
		t.Fatalf("FinalNote = %q, want ready message", report.FinalNote)
	}
}

func TestRunChecksPreservesCheckOrder(t *testing.T) {
	ids := []string{"z-last", "a-first", "m-middle"}
	checks := make([]Check, 0, len(ids))
	for _, id := range ids {
		checks = append(checks, Check{ID: id, Run: func(context.Context) error { return nil }})
	}

	results := RunChecks(context.Background(), checks)
	for i, r := range results {
		if r.ID != ids[i] {
			t.Fatalf("results[%d].ID = %q, want %q", i, r.ID, ids[i])
		}
	}
}
