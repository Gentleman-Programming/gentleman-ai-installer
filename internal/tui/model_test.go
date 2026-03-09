package tui

import (
	"fmt"
	"reflect"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gentleman-programming/gentle-ai/internal/model"
	"github.com/gentleman-programming/gentle-ai/internal/pipeline"
	"github.com/gentleman-programming/gentle-ai/internal/planner"
	"github.com/gentleman-programming/gentle-ai/internal/system"
	"github.com/gentleman-programming/gentle-ai/internal/tui/screens"
)

func TestNavigationWelcomeToDetection(t *testing.T) {
	m := NewModel(system.DetectionResult{}, "dev")

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	state := updated.(Model)

	if state.Screen != ScreenDetection {
		t.Fatalf("screen = %v, want %v", state.Screen, ScreenDetection)
	}
}

func TestNavigationBackWithEscape(t *testing.T) {
	m := NewModel(system.DetectionResult{}, "dev")
	m.Screen = ScreenPersona

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	state := updated.(Model)

	if state.Screen != ScreenAgents {
		t.Fatalf("screen = %v, want %v", state.Screen, ScreenAgents)
	}
}

func TestAgentSelectionToggleAndContinue(t *testing.T) {
	m := NewModel(system.DetectionResult{}, "dev")
	m.Screen = ScreenAgents
	m.Selection.Agents = []model.AgentID{model.AgentClaudeCode}
	m.Cursor = 0

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeySpace})
	state := updated.(Model)

	if len(state.Selection.Agents) != 0 {
		t.Fatalf("agents = %v, want empty", state.Selection.Agents)
	}

	state.Cursor = len(screensAgentOptions())
	updated, _ = state.Update(tea.KeyMsg{Type: tea.KeyEnter})
	state = updated.(Model)

	if state.Screen != ScreenAgents {
		t.Fatalf("screen changed with no selected agents: %v", state.Screen)
	}

	state.Selection.Agents = []model.AgentID{model.AgentOpenCode}
	updated, _ = state.Update(tea.KeyMsg{Type: tea.KeyEnter})
	state = updated.(Model)

	if state.Screen != ScreenPersona {
		t.Fatalf("screen = %v, want %v", state.Screen, ScreenPersona)
	}
}

func TestReviewToInstallingInitializesProgress(t *testing.T) {
	m := NewModel(system.DetectionResult{}, "dev")
	m.Screen = ScreenReview

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	state := updated.(Model)

	if state.Screen != ScreenInstalling {
		t.Fatalf("screen = %v, want %v", state.Screen, ScreenInstalling)
	}

	if state.Progress.Current != 0 {
		t.Fatalf("progress current = %d, want 0", state.Progress.Current)
	}
}

func TestStepProgressMsgUpdatesProgressState(t *testing.T) {
	m := NewModel(system.DetectionResult{}, "dev")
	m.Screen = ScreenInstalling
	m.Progress = NewProgressState([]string{"step-a", "step-b"})

	// Send running event for step-a.
	updated, _ := m.Update(StepProgressMsg{StepID: "step-a", Status: pipeline.StepStatusRunning})
	state := updated.(Model)
	if state.Progress.Items[0].Status != ProgressStatusRunning {
		t.Fatalf("step-a status = %q, want running", state.Progress.Items[0].Status)
	}

	// Send succeeded event for step-a.
	updated, _ = state.Update(StepProgressMsg{StepID: "step-a", Status: pipeline.StepStatusSucceeded})
	state = updated.(Model)
	if state.Progress.Items[0].Status != string(pipeline.StepStatusSucceeded) {
		t.Fatalf("step-a status = %q, want succeeded", state.Progress.Items[0].Status)
	}

	// Send failed event for step-b.
	updated, _ = state.Update(StepProgressMsg{StepID: "step-b", Status: pipeline.StepStatusFailed, Err: fmt.Errorf("oops")})
	state = updated.(Model)
	if state.Progress.Items[1].Status != string(pipeline.StepStatusFailed) {
		t.Fatalf("step-b status = %q, want failed", state.Progress.Items[1].Status)
	}

	if !state.Progress.HasFailures() {
		t.Fatalf("expected HasFailures() = true")
	}
}

func TestPipelineDoneMsgMarksCompletion(t *testing.T) {
	m := NewModel(system.DetectionResult{}, "dev")
	m.Screen = ScreenInstalling
	m.pipelineRunning = true
	m.Progress = NewProgressState([]string{"step-x"})
	m.Progress.Start(0)

	// Simulate pipeline completion with a real step result.
	result := pipeline.ExecutionResult{
		Apply: pipeline.StageResult{
			Success: true,
			Steps: []pipeline.StepResult{
				{StepID: "step-x", Status: pipeline.StepStatusSucceeded},
			},
		},
	}
	updated, _ := m.Update(PipelineDoneMsg{Result: result})
	state := updated.(Model)

	if state.pipelineRunning {
		t.Fatalf("expected pipelineRunning = false")
	}

	if !state.Progress.Done() {
		t.Fatalf("expected progress to be done")
	}
}

func TestPipelineDoneMsgSurfacesFailedSteps(t *testing.T) {
	m := NewModel(system.DetectionResult{}, "dev")
	m.Screen = ScreenInstalling
	m.pipelineRunning = true
	m.Progress = NewProgressState([]string{"step-ok", "step-bad"})

	result := pipeline.ExecutionResult{
		Apply: pipeline.StageResult{
			Success: false,
			Err:     fmt.Errorf("step-bad failed"),
			Steps: []pipeline.StepResult{
				{StepID: "step-ok", Status: pipeline.StepStatusSucceeded},
				{StepID: "step-bad", Status: pipeline.StepStatusFailed, Err: fmt.Errorf("skill inject: write failed")},
			},
		},
		Err: fmt.Errorf("step-bad failed"),
	}
	updated, _ := m.Update(PipelineDoneMsg{Result: result})
	state := updated.(Model)

	if !state.Progress.HasFailures() {
		t.Fatalf("expected HasFailures() = true")
	}

	// Verify that the error message appears in the logs.
	found := false
	for _, log := range state.Progress.Logs {
		if contains(log, "skill inject: write failed") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected error detail in logs, got: %v", state.Progress.Logs)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstring(s, substr)
}

func searchSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestInstallingScreenManualFallbackWithoutExecuteFn(t *testing.T) {
	m := NewModel(system.DetectionResult{}, "dev")
	m.Screen = ScreenInstalling
	m.Progress = NewProgressState([]string{"step-1", "step-2"})
	m.Progress.Start(0)
	// ExecuteFn is nil — manual fallback should work.

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	state := updated.(Model)

	// First enter advances step-1 to succeeded.
	if state.Progress.Items[0].Status != "succeeded" {
		t.Fatalf("step-1 status = %q, want succeeded", state.Progress.Items[0].Status)
	}
}

func TestEscBlockedWhilePipelineRunning(t *testing.T) {
	m := NewModel(system.DetectionResult{}, "dev")
	m.Screen = ScreenInstalling
	m.pipelineRunning = true

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	state := updated.(Model)

	if state.Screen != ScreenInstalling {
		t.Fatalf("screen = %v, want ScreenInstalling (esc should be blocked)", state.Screen)
	}
}

func TestInstallingDoneToComplete(t *testing.T) {
	m := NewModel(system.DetectionResult{}, "dev")
	m.Screen = ScreenInstalling
	m.Progress = NewProgressState([]string{"only-step"})
	m.Progress.Mark(0, string(pipeline.StepStatusSucceeded))

	// Progress is at 100%, enter should go to complete.
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	state := updated.(Model)

	if state.Screen != ScreenComplete {
		t.Fatalf("screen = %v, want ScreenComplete", state.Screen)
	}
}

func TestBuildProgressLabelsFromResolvedPlan(t *testing.T) {
	resolved := planner.ResolvedPlan{
		Agents:            []model.AgentID{model.AgentClaudeCode},
		OrderedComponents: []model.ComponentID{model.ComponentEngram, model.ComponentSDD},
	}

	labels := buildProgressLabels(resolved)

	want := []string{
		"prepare:check-dependencies",
		"prepare:backup-snapshot",
		"apply:rollback-restore",
		"agent:claude-code",
		"component:engram",
		"component:sdd",
	}

	if !reflect.DeepEqual(labels, want) {
		t.Fatalf("labels = %v, want %v", labels, want)
	}
}

func TestBackupRestoreMsgHandledGracefully(t *testing.T) {
	m := NewModel(system.DetectionResult{}, "dev")
	m.Progress = NewProgressState([]string{})

	// Error case.
	updated, _ := m.Update(BackupRestoreMsg{Err: fmt.Errorf("restore-error")})
	state := updated.(Model)
	if state.Err == nil {
		t.Fatalf("expected Err to be set")
	}

	// Success case.
	state.Err = nil
	updated, _ = state.Update(BackupRestoreMsg{})
	state = updated.(Model)
	if state.Err != nil {
		t.Fatalf("unexpected Err: %v", state.Err)
	}
}

func screensAgentOptions() []model.AgentID {
	return []model.AgentID{model.AgentClaudeCode, model.AgentOpenCode}
}

// TestWelcomeOptionsCount verifies that the welcome menu now exposes four
// options (Start installation, Manage backups, Discover skills, Quit).
func TestWelcomeOptionsCount(t *testing.T) {
	m := NewModel(system.DetectionResult{}, "dev")

	count := m.optionCount()

	if count != 4 {
		t.Fatalf("optionCount on ScreenWelcome = %d, want 4", count)
	}
}

// TestConfirmSelection_DiscoverSkills_Index2 verifies that pressing enter
// with cursor=2 on the welcome screen navigates to ScreenSkillDiscovery
// or stays on ScreenWelcome with an error when npx is not available.
// Because npx availability is environment-dependent, we accept either
// ScreenSkillDiscovery (npx present) or ScreenWelcome+Err (npx absent).
func TestConfirmSelection_DiscoverSkills_Index2(t *testing.T) {
	m := NewModel(system.DetectionResult{}, "dev")
	m.Cursor = 2

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	state := updated.(Model)

	switch state.Screen {
	case ScreenSkillDiscovery:
		// npx is present in this environment — correct transition.
	case ScreenWelcome:
		// npx is absent — error must be set.
		if state.Err == nil {
			t.Fatalf("expected Err to be set when npx is missing, got nil")
		}
	default:
		t.Fatalf("unexpected screen %v after selecting Discover skills", state.Screen)
	}
}

// TestStartSkillDiscovery_NpxMissing ensures that startSkillDiscovery sets
// m.Err and stays on ScreenWelcome when npx cannot be found.
// We simulate the missing-npx case by temporarily overriding PATH.
func TestStartSkillDiscovery_NpxMissing(t *testing.T) {
	// Temporarily clear PATH so exec.LookPath("npx") always fails.
	t.Setenv("PATH", "")

	m := NewModel(system.DetectionResult{}, "dev")
	// m.Screen starts as ScreenWelcome.

	updated, cmd := m.startSkillDiscovery()
	state := updated.(Model)

	if state.Screen != ScreenWelcome {
		t.Fatalf("screen = %v, want ScreenWelcome when npx is missing", state.Screen)
	}

	if state.Err == nil {
		t.Fatalf("expected Err to be non-nil when npx is missing")
	}

	if cmd != nil {
		t.Fatalf("expected nil cmd when npx is missing, got %v", cmd)
	}
}

// --- New UX tests for skill discovery browsing key changes and search flow ---

// skillDiscoveryModel returns a model already on ScreenSkillDiscovery with
// Browsing sub-state and a set of skills preloaded.
func skillDiscoveryModel(skills []screens.SkillDiscoveryItem) Model {
	m := NewModel(system.DetectionResult{}, "dev")
	m.Screen = ScreenSkillDiscovery
	m.skillDiscovery = skillDiscoveryState{
		SubState:     screens.SkillDiscoveryBrowsing,
		Skills:       skills,
		CurrentIndex: 0,
	}
	return m
}

func makeSkills(names ...string) []screens.SkillDiscoveryItem {
	items := make([]screens.SkillDiscoveryItem, len(names))
	for i, n := range names {
		items[i] = screens.SkillDiscoveryItem{Name: n, Source: "test-source", Installs: 1000}
	}
	return items
}

// TestBrowsing_UpKeyAddsAndAdvances checks that pressing ↑ adds the current
// skill to Selected and advances the index.
func TestBrowsing_UpKeyAddsAndAdvances(t *testing.T) {
	skills := makeSkills("skill-a", "skill-b")
	m := skillDiscoveryModel(skills)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	state := updated.(Model)

	if len(state.skillDiscovery.Selected) != 1 {
		t.Fatalf("Selected len = %d, want 1", len(state.skillDiscovery.Selected))
	}
	if state.skillDiscovery.Selected[0].Name != "skill-a" {
		t.Fatalf("Selected[0] = %q, want skill-a", state.skillDiscovery.Selected[0].Name)
	}
	if state.skillDiscovery.CurrentIndex != 1 {
		t.Fatalf("CurrentIndex = %d, want 1", state.skillDiscovery.CurrentIndex)
	}
	if state.skillDiscovery.SubState != screens.SkillDiscoveryBrowsing {
		t.Fatalf("SubState = %v, want Browsing", state.skillDiscovery.SubState)
	}
}

// TestBrowsing_RightKeySkipsWithoutAdding checks that pressing → skips without
// adding the skill to Selected.
func TestBrowsing_RightKeySkipsWithoutAdding(t *testing.T) {
	skills := makeSkills("skill-a", "skill-b")
	m := skillDiscoveryModel(skills)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRight})
	state := updated.(Model)

	if len(state.skillDiscovery.Selected) != 0 {
		t.Fatalf("Selected len = %d, want 0 after skip", len(state.skillDiscovery.Selected))
	}
	if state.skillDiscovery.CurrentIndex != 1 {
		t.Fatalf("CurrentIndex = %d, want 1", state.skillDiscovery.CurrentIndex)
	}
}

// TestBrowsing_EscKeyTransitionsToSearchPrompt verifies that pressing ESC during
// browsing moves to SearchPrompt (not directly to Done or Installing).
func TestBrowsing_EscKeyTransitionsToSearchPrompt(t *testing.T) {
	skills := makeSkills("skill-a")
	m := skillDiscoveryModel(skills)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	state := updated.(Model)

	if state.skillDiscovery.SubState != screens.SkillDiscoverySearchPrompt {
		t.Fatalf("SubState = %v, want SearchPrompt", state.skillDiscovery.SubState)
	}
}

// TestBrowsing_LastSkillTransitionsToSearchPrompt verifies that browsing past
// the last skill transitions to SearchPrompt, not directly to Installing.
func TestBrowsing_LastSkillTransitionsToSearchPrompt(t *testing.T) {
	skills := makeSkills("only-skill")
	m := skillDiscoveryModel(skills)
	m.skillDiscovery.Selected = makeSkills("already-selected")

	// Press → to skip the only skill — end of list.
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRight})
	state := updated.(Model)

	if state.skillDiscovery.SubState != screens.SkillDiscoverySearchPrompt {
		t.Fatalf("SubState = %v, want SearchPrompt after browsing last skill", state.skillDiscovery.SubState)
	}
	if state.skillDiscovery.SearchQuery != "" {
		t.Fatalf("SearchQuery = %q, want empty after transition", state.skillDiscovery.SearchQuery)
	}
}

// TestSearchPrompt_TypeCharsBuildsQuery verifies that typing rune keys appends
// to SearchQuery.
func TestSearchPrompt_TypeCharsBuildsQuery(t *testing.T) {
	m := NewModel(system.DetectionResult{}, "dev")
	m.Screen = ScreenSkillDiscovery
	m.skillDiscovery.SubState = screens.SkillDiscoverySearchPrompt

	for _, r := range "go" {
		updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = updated.(Model)
	}

	if m.skillDiscovery.SearchQuery != "go" {
		t.Fatalf("SearchQuery = %q, want go", m.skillDiscovery.SearchQuery)
	}
}

// TestSearchPrompt_BackspaceDeletesLastChar verifies that backspace removes the
// last character from SearchQuery.
func TestSearchPrompt_BackspaceDeletesLastChar(t *testing.T) {
	m := NewModel(system.DetectionResult{}, "dev")
	m.Screen = ScreenSkillDiscovery
	m.skillDiscovery.SubState = screens.SkillDiscoverySearchPrompt
	m.skillDiscovery.SearchQuery = "golang"

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	state := updated.(Model)

	if state.skillDiscovery.SearchQuery != "golan" {
		t.Fatalf("SearchQuery = %q, want golan", state.skillDiscovery.SearchQuery)
	}
}

// TestSearchPrompt_BackspaceOnEmptyQueryIsNoOp verifies that backspace on an
// empty query does not panic or change state.
func TestSearchPrompt_BackspaceOnEmptyQueryIsNoOp(t *testing.T) {
	m := NewModel(system.DetectionResult{}, "dev")
	m.Screen = ScreenSkillDiscovery
	m.skillDiscovery.SubState = screens.SkillDiscoverySearchPrompt
	m.skillDiscovery.SearchQuery = ""

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	state := updated.(Model)

	if state.skillDiscovery.SearchQuery != "" {
		t.Fatalf("SearchQuery = %q, want empty", state.skillDiscovery.SearchQuery)
	}
}

// TestSearchPrompt_EnterWithEmptyQueryStartsInstall verifies that pressing
// enter with an empty query transitions to Installing (or Done if nothing selected).
func TestSearchPrompt_EnterWithEmptyQueryStartsInstall(t *testing.T) {
	m := NewModel(system.DetectionResult{}, "dev")
	m.Screen = ScreenSkillDiscovery
	m.skillDiscovery.SubState = screens.SkillDiscoverySearchPrompt
	m.skillDiscovery.SearchQuery = ""
	// No selected skills → should go to Done.

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	state := updated.(Model)

	if state.skillDiscovery.SubState != screens.SkillDiscoveryDone {
		t.Fatalf("SubState = %v, want Done when no skills selected", state.skillDiscovery.SubState)
	}
}

// TestSearchPrompt_EnterWithEmptyQueryAndSelectedSkillsStartsInstall verifies
// that pressing enter with empty query and selected skills goes to Installing.
func TestSearchPrompt_EnterWithEmptyQueryAndSelectedSkillsStartsInstall(t *testing.T) {
	m := NewModel(system.DetectionResult{}, "dev")
	m.Screen = ScreenSkillDiscovery
	m.skillDiscovery.SubState = screens.SkillDiscoverySearchPrompt
	m.skillDiscovery.SearchQuery = ""
	m.skillDiscovery.Selected = makeSkills("my-skill")

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	state := updated.(Model)

	if state.skillDiscovery.SubState != screens.SkillDiscoveryInstalling {
		t.Fatalf("SubState = %v, want Installing", state.skillDiscovery.SubState)
	}
	if cmd == nil {
		t.Fatalf("expected a non-nil install cmd")
	}
}

// TestSearchPrompt_EscKeyStartsInstall verifies that pressing ESC in SearchPrompt
// goes directly to Installing/Done without waiting for search.
func TestSearchPrompt_EscKeyStartsInstall(t *testing.T) {
	m := NewModel(system.DetectionResult{}, "dev")
	m.Screen = ScreenSkillDiscovery
	m.skillDiscovery.SubState = screens.SkillDiscoverySearchPrompt
	m.skillDiscovery.Selected = makeSkills("skill-a")

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	state := updated.(Model)

	if state.skillDiscovery.SubState != screens.SkillDiscoveryInstalling {
		t.Fatalf("SubState = %v, want Installing", state.skillDiscovery.SubState)
	}
}

// TestManualSearchDoneMsg_SetsResultsAndTransitions verifies that receiving a
// manualSearchDoneMsg populates SearchResults and transitions to SearchResults sub-state.
func TestManualSearchDoneMsg_SetsResultsAndTransitions(t *testing.T) {
	m := NewModel(system.DetectionResult{}, "dev")
	m.Screen = ScreenSkillDiscovery
	m.skillDiscovery.SubState = screens.SkillDiscoverySearchPrompt

	results := makeSkills("result-1", "result-2", "result-3")
	updated, _ := m.Update(manualSearchDoneMsg{Results: results})
	state := updated.(Model)

	if state.skillDiscovery.SubState != screens.SkillDiscoverySearchResults {
		t.Fatalf("SubState = %v, want SearchResults", state.skillDiscovery.SubState)
	}
	if len(state.skillDiscovery.SearchResults) != 3 {
		t.Fatalf("SearchResults len = %d, want 3", len(state.skillDiscovery.SearchResults))
	}
	if state.skillDiscovery.SearchSelected == nil {
		t.Fatalf("SearchSelected should be initialised (not nil)")
	}
}

// TestManualSearchDoneMsg_EmptyResultsShowsEmpty verifies that an empty result
// set still transitions to SearchResults sub-state (so "No results found." is shown).
func TestManualSearchDoneMsg_EmptyResultsShowsEmpty(t *testing.T) {
	m := NewModel(system.DetectionResult{}, "dev")
	m.Screen = ScreenSkillDiscovery
	m.skillDiscovery.SubState = screens.SkillDiscoverySearchPrompt

	updated, _ := m.Update(manualSearchDoneMsg{Results: nil})
	state := updated.(Model)

	if state.skillDiscovery.SubState != screens.SkillDiscoverySearchResults {
		t.Fatalf("SubState = %v, want SearchResults", state.skillDiscovery.SubState)
	}
	if len(state.skillDiscovery.SearchResults) != 0 {
		t.Fatalf("SearchResults len = %d, want 0", len(state.skillDiscovery.SearchResults))
	}
}

// TestSearchResults_ArrowKeysNavigateCursor verifies that ↑/↓ move the
// highlighted cursor through search results.
func TestSearchResults_ArrowKeysNavigateCursor(t *testing.T) {
	m := NewModel(system.DetectionResult{}, "dev")
	m.Screen = ScreenSkillDiscovery
	m.skillDiscovery.SubState = screens.SkillDiscoverySearchResults
	m.skillDiscovery.SearchResults = makeSkills("r1", "r2", "r3")
	m.skillDiscovery.SearchResultCursor = 0

	// Press ↓ — cursor moves to 1.
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	state := updated.(Model)
	if state.skillDiscovery.SearchResultCursor != 1 {
		t.Fatalf("cursor = %d, want 1", state.skillDiscovery.SearchResultCursor)
	}

	// Press ↓ again — cursor moves to 2.
	updated, _ = state.Update(tea.KeyMsg{Type: tea.KeyDown})
	state = updated.(Model)
	if state.skillDiscovery.SearchResultCursor != 2 {
		t.Fatalf("cursor = %d, want 2", state.skillDiscovery.SearchResultCursor)
	}

	// Press ↑ — cursor moves back to 1.
	updated, _ = state.Update(tea.KeyMsg{Type: tea.KeyUp})
	state = updated.(Model)
	if state.skillDiscovery.SearchResultCursor != 1 {
		t.Fatalf("cursor = %d, want 1 after ↑", state.skillDiscovery.SearchResultCursor)
	}
}

// TestSearchResults_EnterAddsCursorItemAndReturnsToPrompt verifies that pressing
// Enter adds only the currently highlighted item to Selected and returns to SearchPrompt.
func TestSearchResults_EnterAddsCursorItemAndReturnsToPrompt(t *testing.T) {
	m := NewModel(system.DetectionResult{}, "dev")
	m.Screen = ScreenSkillDiscovery
	m.skillDiscovery.SubState = screens.SkillDiscoverySearchResults
	m.skillDiscovery.SearchResults = makeSkills("r1", "r2", "r3")
	m.skillDiscovery.SearchResultCursor = 1 // r2 is highlighted

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	state := updated.(Model)

	if state.skillDiscovery.SubState != screens.SkillDiscoverySearchPrompt {
		t.Fatalf("SubState = %v, want SearchPrompt", state.skillDiscovery.SubState)
	}
	if len(state.skillDiscovery.Selected) != 1 {
		t.Fatalf("Selected len = %d, want 1 (r2 only)", len(state.skillDiscovery.Selected))
	}
	if state.skillDiscovery.Selected[0].Name != "r2" {
		t.Fatalf("Selected[0].Name = %q, want r2", state.skillDiscovery.Selected[0].Name)
	}
	if state.skillDiscovery.SearchQuery != "" {
		t.Fatalf("SearchQuery = %q, want empty after confirm", state.skillDiscovery.SearchQuery)
	}
	if state.skillDiscovery.SearchResults != nil {
		t.Fatalf("SearchResults should be nil after confirm")
	}
}

// TestSearchResults_EscReturnsToSearchPromptWithoutAdding verifies that pressing
// ESC in SearchResults discards results and goes back to SearchPrompt.
func TestSearchResults_EscReturnsToSearchPromptWithoutAdding(t *testing.T) {
	m := NewModel(system.DetectionResult{}, "dev")
	m.Screen = ScreenSkillDiscovery
	m.skillDiscovery.SubState = screens.SkillDiscoverySearchResults
	m.skillDiscovery.SearchResults = makeSkills("r1")
	m.skillDiscovery.Selected = makeSkills("already-selected")

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	state := updated.(Model)

	if state.skillDiscovery.SubState != screens.SkillDiscoverySearchPrompt {
		t.Fatalf("SubState = %v, want SearchPrompt", state.skillDiscovery.SubState)
	}
	// previously selected item should still be there, search results discarded
	if len(state.skillDiscovery.Selected) != 1 {
		t.Fatalf("Selected len = %d, want 1 (unchanged)", len(state.skillDiscovery.Selected))
	}
	if state.skillDiscovery.SearchResults != nil {
		t.Fatalf("SearchResults should be nil after ESC")
	}
}

// TestSearchResults_EnterWithNoResultsGoesBackToPrompt verifies that pressing
// enter when SearchResults is empty returns to SearchPrompt.
func TestSearchResults_EnterWithNoResultsGoesBackToPrompt(t *testing.T) {
	m := NewModel(system.DetectionResult{}, "dev")
	m.Screen = ScreenSkillDiscovery
	m.skillDiscovery.SubState = screens.SkillDiscoverySearchResults
	m.skillDiscovery.SearchResults = nil

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	state := updated.(Model)

	if state.skillDiscovery.SubState != screens.SkillDiscoverySearchPrompt {
		t.Fatalf("SubState = %v, want SearchPrompt when no results", state.skillDiscovery.SubState)
	}
}

// TestInstalling_KeysBlockedExceptCtrlC verifies that keys are blocked during
// the Installing sub-state except for ctrl+c.
func TestSkillDiscoveryInstalling_KeysBlocked(t *testing.T) {
	m := NewModel(system.DetectionResult{}, "dev")
	m.Screen = ScreenSkillDiscovery
	m.skillDiscovery.SubState = screens.SkillDiscoveryInstalling
	m.skillDiscovery.Selected = makeSkills("s1")

	for _, k := range []tea.KeyMsg{
		{Type: tea.KeyEnter},
		{Type: tea.KeyUp},
		{Type: tea.KeyRunes, Runes: []rune("q")},
	} {
		updated, cmd := m.Update(k)
		state := updated.(Model)
		if state.skillDiscovery.SubState != screens.SkillDiscoveryInstalling {
			t.Fatalf("key %q changed sub-state during Installing — should be blocked", k.String())
		}
		if cmd != nil {
			t.Fatalf("key %q returned non-nil cmd during Installing — should be blocked", k.String())
		}
	}
}

// TestLoading_EscReturnsToWelcome verifies that pressing ESC during the Loading
// sub-state returns to the Welcome screen.
func TestLoading_EscReturnsToWelcome(t *testing.T) {
	m := NewModel(system.DetectionResult{}, "dev")
	m.Screen = ScreenSkillDiscovery
	m.skillDiscovery.SubState = screens.SkillDiscoveryLoading

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	state := updated.(Model)

	if state.Screen != ScreenWelcome {
		t.Fatalf("Screen = %v, want ScreenWelcome after ESC on Loading", state.Screen)
	}
}

// TestBuildInstallArg_UsesSkillIDNotName verifies that buildInstallArg
// produces source@skillId and does NOT use the display Name.
func TestBuildInstallArg_UsesSkillIDNotName(t *testing.T) {
	skill := screens.SkillDiscoveryItem{
		Name:    "react:components",        // display name — has colon, should NOT appear
		SkillID: "react-components",        // actual API skillId
		Source:  "google-labs-code/stitch-skills",
	}

	got := buildInstallArg(skill)
	want := "google-labs-code/stitch-skills@react-components"

	if got != want {
		t.Errorf("buildInstallArg = %q, want %q", got, want)
	}
}

// TestBuildInstallArg_FallsBackToNameWhenSkillIDEmpty verifies that when SkillID
// is empty (shouldn't happen in practice), Name is used as fallback.
func TestBuildInstallArg_FallsBackToNameWhenSkillIDEmpty(t *testing.T) {
	skill := screens.SkillDiscoveryItem{
		Name:   "my-skill",
		Source: "vercel-labs/agent-skills",
	}

	got := buildInstallArg(skill)
	want := "vercel-labs/agent-skills@my-skill"

	if got != want {
		t.Errorf("buildInstallArg = %q, want %q", got, want)
	}
}

// TestBuildInstallArg_NeverContainsDisplayNameColon verifies that a skill whose
// Name contains ':' (broken git ref) is never passed as the install argument
// when SkillID is present.
func TestBuildInstallArg_NeverContainsDisplayNameColon(t *testing.T) {
	skill := screens.SkillDiscoveryItem{
		Name:    "namespace:thing",
		SkillID: "namespace-thing",
		Source:  "some-org/some-repo",
	}

	got := buildInstallArg(skill)
	if got != "some-org/some-repo@namespace-thing" {
		t.Errorf("buildInstallArg = %q — should use SkillID, not Name with colon", got)
	}
}
