package tui

import (
	"context"
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gentleman-programming/gentle-ai/internal/backup"
	"github.com/gentleman-programming/gentle-ai/internal/catalog"
	"github.com/gentleman-programming/gentle-ai/internal/model"
	"github.com/gentleman-programming/gentle-ai/internal/pipeline"
	"github.com/gentleman-programming/gentle-ai/internal/planner"
	"github.com/gentleman-programming/gentle-ai/internal/skillssh"
	"github.com/gentleman-programming/gentle-ai/internal/stackdetect"
	"github.com/gentleman-programming/gentle-ai/internal/system"
	"github.com/gentleman-programming/gentle-ai/internal/tui/screens"
	"github.com/gentleman-programming/gentle-ai/internal/update"
)

// spinnerFrames are the braille characters used for the animated spinner.
var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// TickMsg drives the spinner animation on the installing screen.
type TickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

// StepProgressMsg is sent from the pipeline goroutine when a step changes status.
type StepProgressMsg struct {
	StepID string
	Status pipeline.StepStatus
	Err    error
}

// PipelineDoneMsg is sent when the pipeline finishes execution.
type PipelineDoneMsg struct {
	Result pipeline.ExecutionResult
}

// BackupRestoreMsg is sent when a backup restore completes.
type BackupRestoreMsg struct {
	Err error
}

// UpdateCheckResultMsg is sent when the background update check completes.
type UpdateCheckResultMsg struct {
	Results []update.UpdateResult
}

// skillsLoadedMsg is sent when the skill discovery search completes.
type skillsLoadedMsg struct {
	Skills []screens.SkillDiscoveryItem
	Err    error
}

// skillInstallDoneMsg is sent when a single skill installation finishes.
type skillInstallDoneMsg struct {
	Index int
	Err   error
}

// manualSearchDoneMsg is sent when a manual (user-typed) skill search completes.
type manualSearchDoneMsg struct {
	Results []screens.SkillDiscoveryItem
	Err     error
}

// skillDiscoveryState holds the transient state for the skill discovery screen.
type skillDiscoveryState struct {
	SubState     screens.SkillDiscoverySubState
	Skills       []screens.SkillDiscoveryItem
	CurrentIndex int
	Selected     []screens.SkillDiscoveryItem
	Installing   int
	Error        string
	Failures     map[int]string // index in Selected → error message

	// Manual search sub-state fields.
	SearchQuery        string
	SearchResults      []screens.SkillDiscoveryItem
	SearchSelected     map[int]bool
	SearchResultCursor int
}
// ExecuteFunc builds and runs the installation pipeline. It receives a ProgressFunc
// callback to emit step-level progress events, and returns the ExecutionResult.
type ExecuteFunc func(
	selection model.Selection,
	resolved planner.ResolvedPlan,
	detection system.DetectionResult,
	onProgress pipeline.ProgressFunc,
) pipeline.ExecutionResult

// RestoreFunc restores a backup from a manifest.
type RestoreFunc func(manifest backup.Manifest) error

type Screen int

const (
	ScreenUnknown Screen = iota
	ScreenWelcome
	ScreenDetection
	ScreenAgents
	ScreenPersona
	ScreenPreset
	ScreenDependencyTree
	ScreenReview
	ScreenInstalling
	ScreenComplete
	ScreenBackups
	ScreenSkillDiscovery
)

type Model struct {
	Screen         Screen
	PreviousScreen Screen
	Width          int
	Height         int
	Cursor         int
	Version        string
	SpinnerFrame   int

	Selection      model.Selection
	Detection      system.DetectionResult
	DependencyPlan planner.ResolvedPlan
	Review         planner.ReviewPayload
	Progress       ProgressState
	Execution      pipeline.ExecutionResult
	Backups        []backup.Manifest
	Err            error

	// ExecuteFn is called to run the real pipeline. When nil, the installing
	// screen falls back to manual step-through (useful for tests/development).
	ExecuteFn ExecuteFunc

	// RestoreFn is called to restore a backup. When nil, restore is a no-op.
	RestoreFn RestoreFunc

	// UpdateResults holds the results of the background update check.
	UpdateResults []update.UpdateResult

	// UpdateCheckDone is true once the background update check has completed.
	UpdateCheckDone bool

	// pipelineRunning tracks whether the pipeline goroutine is active.
	pipelineRunning bool

	// skillDiscovery holds the transient state for the skill discovery screen.
	skillDiscovery skillDiscoveryState
}

func NewModel(detection system.DetectionResult, version string) Model {
	selection := model.Selection{
		Agents:     preselectedAgents(detection),
		Persona:    model.PersonaGentleman,
		Preset:     model.PresetFullGentleman,
		Components: componentsForPreset(model.PresetFullGentleman),
	}

	return Model{
		Screen:    ScreenWelcome,
		Version:   version,
		Selection: selection,
		Detection: detection,
		Progress: NewProgressState([]string{
			"Install dependencies",
			"Configure selected agents",
			"Inject ecosystem components",
		}),
	}
}

func (m Model) Init() tea.Cmd {
	version := m.Version
	profile := m.Detection.System.Profile

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		results := update.CheckAll(ctx, version, profile)
		return UpdateCheckResultMsg{Results: results}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil
	case TickMsg:
		if m.Screen == ScreenInstalling && !m.Progress.Done() {
			m.SpinnerFrame = (m.SpinnerFrame + 1) % len(spinnerFrames)
			return m, tickCmd()
		}
		return m, nil
	case StepProgressMsg:
		return m.handleStepProgress(msg)
	case PipelineDoneMsg:
		return m.handlePipelineDone(msg)
	case BackupRestoreMsg:
		return m.handleBackupRestore(msg)
	case UpdateCheckResultMsg:
		m.UpdateResults = msg.Results
		m.UpdateCheckDone = true
		return m, nil
	case skillsLoadedMsg:
		return m.handleSkillsLoaded(msg)
	case skillInstallDoneMsg:
		return m.handleSkillInstallDone(msg)
	case manualSearchDoneMsg:
		return m.handleManualSearchDone(msg)
	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	}

	return m, nil
}

func (m Model) handleStepProgress(msg StepProgressMsg) (tea.Model, tea.Cmd) {
	if m.Screen != ScreenInstalling {
		return m, nil
	}

	idx := m.findProgressItem(msg.StepID)
	if idx < 0 {
		return m, nil
	}

	switch msg.Status {
	case pipeline.StepStatusRunning:
		m.Progress.Start(idx)
		m.Progress.AppendLog("running: %s", msg.StepID)
	case pipeline.StepStatusSucceeded:
		m.Progress.Mark(idx, string(pipeline.StepStatusSucceeded))
		m.Progress.AppendLog("done: %s", msg.StepID)
	case pipeline.StepStatusFailed:
		m.Progress.Mark(idx, string(pipeline.StepStatusFailed))
		errMsg := "unknown error"
		if msg.Err != nil {
			errMsg = msg.Err.Error()
		}
		m.Progress.AppendLog("FAILED: %s — %s", msg.StepID, errMsg)
	}

	return m, nil
}

func (m Model) handlePipelineDone(msg PipelineDoneMsg) (tea.Model, tea.Cmd) {
	m.Execution = msg.Result
	m.pipelineRunning = false

	// Rebuild progress from real step results so failed steps show ✗ instead
	// of being blindly marked as succeeded.
	m.Progress = ProgressFromExecution(msg.Result)

	// Surface individual error messages so the user knows WHAT failed.
	appendStepErrors := func(steps []pipeline.StepResult) {
		for _, step := range steps {
			if step.Status == pipeline.StepStatusFailed && step.Err != nil {
				m.Progress.AppendLog("FAILED: %s — %s", step.StepID, step.Err.Error())
			}
		}
	}
	appendStepErrors(msg.Result.Prepare.Steps)
	appendStepErrors(msg.Result.Apply.Steps)

	if msg.Result.Err != nil {
		m.Progress.AppendLog("pipeline completed with errors")
	} else {
		m.Progress.AppendLog("pipeline completed successfully")
	}

	return m, nil
}

func (m Model) handleBackupRestore(msg BackupRestoreMsg) (tea.Model, tea.Cmd) {
	if msg.Err != nil {
		m.Err = msg.Err
		m.Progress.AppendLog("restore failed: %s", msg.Err.Error())
	} else {
		m.Progress.AppendLog("backup restored successfully")
	}
	return m, nil
}

func (m Model) handleSkillsLoaded(msg skillsLoadedMsg) (tea.Model, tea.Cmd) {
	if m.Screen != ScreenSkillDiscovery {
		return m, nil
	}

	if msg.Err != nil {
		m.skillDiscovery.Error = msg.Err.Error()
		// Stay in loading sub-state so the user sees the error banner; a future
		// Batch C key-handler will allow the user to quit the screen.
		return m, nil
	}

	if len(msg.Skills) == 0 {
		// No skills found — jump straight to browsing so the empty-state UI is
		// shown instead of leaving the user stuck on the spinner.
		m.skillDiscovery.SubState = screens.SkillDiscoveryBrowsing
		m.skillDiscovery.Skills = nil
		return m, nil
	}

	m.skillDiscovery.SubState = screens.SkillDiscoveryBrowsing
	m.skillDiscovery.Skills = msg.Skills
	m.skillDiscovery.CurrentIndex = 0
	return m, nil
}

func (m Model) handleSkillInstallDone(msg skillInstallDoneMsg) (tea.Model, tea.Cmd) {
	if m.Screen != ScreenSkillDiscovery {
		return m, nil
	}

	if msg.Err != nil {
		// Non-fatal: record per-skill failure and continue to the next skill.
		if m.skillDiscovery.Failures == nil {
			m.skillDiscovery.Failures = make(map[int]string)
		}
		m.skillDiscovery.Failures[msg.Index] = msg.Err.Error()
	}

	next := msg.Index + 1
	if next >= len(m.skillDiscovery.Selected) {
		// All installs done.
		m.skillDiscovery.SubState = screens.SkillDiscoveryDone
		return m, nil
	}

	m.skillDiscovery.Installing = next
	return m, InstallSkillCmd(m.skillDiscovery.Selected[next], next)
}

func (m Model) handleManualSearchDone(msg manualSearchDoneMsg) (tea.Model, tea.Cmd) {
	if m.Screen != ScreenSkillDiscovery {
		return m, nil
	}

	if msg.Err != nil {
		m.skillDiscovery.Error = msg.Err.Error()
	}

	m.skillDiscovery.SearchResults = msg.Results
	m.skillDiscovery.SearchSelected = make(map[int]bool)
	m.skillDiscovery.SubState = screens.SkillDiscoverySearchResults
	return m, nil
}

func (m Model) findProgressItem(stepID string) int {
	for i, item := range m.Progress.Items {
		if item.Label == stepID {
			return i
		}
	}
	return -1
}

func (m Model) View() string {
	switch m.Screen {
	case ScreenWelcome:
		var banner string
		if m.UpdateCheckDone && update.HasUpdates(m.UpdateResults) {
			banner = "Updates available: " + update.UpdateSummaryLine(m.UpdateResults)
		}
		return screens.RenderWelcome(m.Cursor, m.Version, banner)
	case ScreenDetection:
		return screens.RenderDetection(m.Detection, m.Cursor)
	case ScreenAgents:
		return screens.RenderAgents(m.Selection.Agents, m.Cursor)
	case ScreenPersona:
		return screens.RenderPersona(m.Selection.Persona, m.Cursor)
	case ScreenPreset:
		return screens.RenderPreset(m.Selection.Preset, m.Cursor)
	case ScreenDependencyTree:
		return screens.RenderDependencyTree(m.DependencyPlan, m.Selection, m.Cursor)
	case ScreenReview:
		return screens.RenderReview(m.Review, m.Cursor)
	case ScreenInstalling:
		return screens.RenderInstalling(m.Progress.ViewModel(), spinnerFrames[m.SpinnerFrame])
	case ScreenComplete:
		return screens.RenderComplete(screens.CompletePayload{
			ConfiguredAgents:    len(m.Selection.Agents),
			InstalledComponents: len(m.Selection.Components),
			FailedSteps:         extractFailedSteps(m.Execution),
			RollbackPerformed:   len(m.Execution.Rollback.Steps) > 0,
			MissingDeps:         extractMissingDeps(m.Detection),
			AvailableUpdates:    extractAvailableUpdates(m.UpdateResults),
		})
	case ScreenBackups:
		return screens.RenderBackups(m.Backups, m.Cursor)
	case ScreenSkillDiscovery:
		return screens.RenderSkillDiscovery(screens.SkillDiscoveryViewState{
			SubState:       m.skillDiscovery.SubState,
			Skills:         m.skillDiscovery.Skills,
			CurrentIndex:   m.skillDiscovery.CurrentIndex,
			Selected:       m.skillDiscovery.Selected,
			Installing:     m.skillDiscovery.Installing,
			Error:          m.skillDiscovery.Error,
			Failures:       m.skillDiscovery.Failures,
			SearchQuery:        m.skillDiscovery.SearchQuery,
			SearchResults:      m.skillDiscovery.SearchResults,
			SearchSelected:     m.skillDiscovery.SearchSelected,
			SearchResultCursor: m.skillDiscovery.SearchResultCursor,
		})
	default:
		return ""
	}
}

func (m Model) handleKeyPress(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Screen-specific handlers that take priority over global bindings.
	if m.Screen == ScreenSkillDiscovery {
		return m.handleSkillDiscoveryKey(key)
	}

	switch key.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "up", "k":
		if m.Cursor > 0 {
			m.Cursor--
		}
		return m, nil
	case "down", "j":
		if m.Cursor+1 < m.optionCount() {
			m.Cursor++
		}
		return m, nil
	case "esc":
		// Don't allow going back while pipeline is running.
		if m.Screen == ScreenInstalling && m.pipelineRunning {
			return m, nil
		}
		return m.goBack(), nil
	case " ":
		switch m.Screen {
		case ScreenAgents:
			m.toggleCurrentAgent()
		case ScreenDependencyTree:
			if m.Selection.Preset == model.PresetCustom {
				m.toggleCurrentComponent()
			}
		}
		return m, nil
	case "enter":
		return m.confirmSelection()
	}

	return m, nil
}

// isSkillAlreadySelected returns true if the skill is already in the Selected list.
func (m Model) isSkillAlreadySelected(skill screens.SkillDiscoveryItem) bool {
	for _, s := range m.skillDiscovery.Selected {
		if s.Name == skill.Name && s.Source == skill.Source {
			return true
		}
	}
	return false
}

// handleSkillDiscoveryKey routes key presses while on ScreenSkillDiscovery.
func (m Model) handleSkillDiscoveryKey(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.skillDiscovery.SubState {
	case screens.SkillDiscoveryLoading:
		// Allow cancelling while the search is in flight.
		switch key.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyEsc:
			m.setScreen(ScreenWelcome)
		}
		return m, nil

	case screens.SkillDiscoveryInstalling:
		// Block all keys while install is running; only ctrl+c is honoured.
		if key.String() == "ctrl+c" {
			return m, tea.Quit
		}
		return m, nil

	case screens.SkillDiscoveryBrowsing:
		switch key.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyUp:
			if len(m.skillDiscovery.Skills) > 0 {
				skill := m.skillDiscovery.Skills[m.skillDiscovery.CurrentIndex]
				if !m.isSkillAlreadySelected(skill) {
					m.skillDiscovery.Selected = append(m.skillDiscovery.Selected, skill)
				} else {
					m.skillDiscovery.Error = "\"" + skill.Name + "\" ya está en la lista"
				}
			}
			return m.advanceSkillIndex()
		case tea.KeyRight:
			return m.advanceSkillIndex()
		case tea.KeyEsc:
			// Finish browsing → go to search prompt.
			m.skillDiscovery.SubState = screens.SkillDiscoverySearchPrompt
			m.skillDiscovery.SearchQuery = ""
		}

	case screens.SkillDiscoverySearchPrompt:
		switch key.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyEsc:
			// Esc = done searching, install everything selected so far.
			return m.startSkillInstall()
		case tea.KeyEnter:
			if m.skillDiscovery.SearchQuery == "" {
				return m.startSkillInstall()
			}
			query := m.skillDiscovery.SearchQuery
			return m, SearchManualCmd(query)
		case tea.KeyBackspace:
			q := m.skillDiscovery.SearchQuery
			if len(q) > 0 {
				m.skillDiscovery.SearchQuery = q[:len(q)-1]
			}
		case tea.KeyRunes:
			m.skillDiscovery.SearchQuery += string(key.Runes)
		}

	case screens.SkillDiscoverySearchResults:
		switch key.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyEsc:
			// Back to search prompt without adding anything.
			m.skillDiscovery.SearchResults = nil
			m.skillDiscovery.SearchSelected = nil
			m.skillDiscovery.SearchResultCursor = 0
			m.skillDiscovery.SearchQuery = ""
			m.skillDiscovery.SubState = screens.SkillDiscoverySearchPrompt
		case tea.KeyUp:
			if m.skillDiscovery.SearchResultCursor > 0 {
				m.skillDiscovery.SearchResultCursor--
			}
		case tea.KeyDown:
			max := len(m.skillDiscovery.SearchResults) - 1
			if m.skillDiscovery.SearchResultCursor < max {
				m.skillDiscovery.SearchResultCursor++
			}
		case tea.KeyEnter:
			// Add the currently highlighted skill and go back to search prompt.
			cursor := m.skillDiscovery.SearchResultCursor
			if cursor < len(m.skillDiscovery.SearchResults) {
				skill := m.skillDiscovery.SearchResults[cursor]
				if !m.isSkillAlreadySelected(skill) {
					m.skillDiscovery.Selected = append(m.skillDiscovery.Selected, skill)
					m.skillDiscovery.Error = ""
				} else {
					m.skillDiscovery.Error = "\"" + skill.Name + "\" ya está en la lista"
				}
			}
			m.skillDiscovery.SearchResults = nil
			m.skillDiscovery.SearchSelected = nil
			m.skillDiscovery.SearchResultCursor = 0
			m.skillDiscovery.SearchQuery = ""
			m.skillDiscovery.SubState = screens.SkillDiscoverySearchPrompt
		}

	case screens.SkillDiscoveryDone:
		switch key.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyEnter, tea.KeyEsc:
			m.setScreen(ScreenWelcome)
		}

	default:
		// Loading sub-state or unknown — only allow ctrl+c.
		if key.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}

	return m, nil
}

// toggleSearchResult toggles selection for the given zero-based index in SearchResults.
func (m *Model) toggleSearchResult(idx int) {
	if idx >= len(m.skillDiscovery.SearchResults) {
		return
	}
	if m.skillDiscovery.SearchSelected == nil {
		m.skillDiscovery.SearchSelected = make(map[int]bool)
	}
	m.skillDiscovery.SearchSelected[idx] = !m.skillDiscovery.SearchSelected[idx]
}

// startSkillInstall transitions to the Installing sub-state (or Done if nothing selected).
func (m Model) startSkillInstall() (tea.Model, tea.Cmd) {
	if len(m.skillDiscovery.Selected) == 0 {
		m.skillDiscovery.SubState = screens.SkillDiscoveryDone
		return m, nil
	}
	m.skillDiscovery.SubState = screens.SkillDiscoveryInstalling
	m.skillDiscovery.Installing = 0
	first := m.skillDiscovery.Selected[0]
	return m, InstallSkillCmd(first, 0)
}

// advanceSkillIndex moves CurrentIndex forward after an add/skip decision.
// When the last skill has been browsed it transitions to SkillDiscoverySearchPrompt
// so the user can optionally search for more skills before installing.
func (m Model) advanceSkillIndex() (tea.Model, tea.Cmd) {
	next := m.skillDiscovery.CurrentIndex + 1
	if next < len(m.skillDiscovery.Skills) {
		m.skillDiscovery.CurrentIndex = next
		return m, nil
	}

	// Reached the end of the recommended list — offer manual search.
	m.skillDiscovery.SubState = screens.SkillDiscoverySearchPrompt
	m.skillDiscovery.SearchQuery = ""
	return m, nil
}

func (m Model) confirmSelection() (tea.Model, tea.Cmd) {
	switch m.Screen {
	case ScreenWelcome:
		switch m.Cursor {
		case 0:
			m.setScreen(ScreenDetection)
		case 1:
			m.setScreen(ScreenBackups)
		case 2:
			return m.startSkillDiscovery()
		default:
			return m, tea.Quit
		}
	case ScreenDetection:
		if m.Cursor == 0 {
			m.setScreen(ScreenAgents)
			return m, nil
		}
		m.setScreen(ScreenWelcome)
	case ScreenAgents:
		agentCount := len(screens.AgentOptions())
		switch {
		case m.Cursor < agentCount:
			m.toggleCurrentAgent()
		case m.Cursor == agentCount && len(m.Selection.Agents) > 0:
			m.setScreen(ScreenPersona)
		case m.Cursor == agentCount+1:
			m.setScreen(ScreenDetection)
		}
	case ScreenPersona:
		options := screens.PersonaOptions()
		if m.Cursor < len(options) {
			m.Selection.Persona = options[m.Cursor]
			m.setScreen(ScreenPreset)
			return m, nil
		}
		m.setScreen(ScreenAgents)
	case ScreenPreset:
		options := screens.PresetOptions()
		if m.Cursor < len(options) {
			m.Selection.Preset = options[m.Cursor]
			m.Selection.Components = componentsForPreset(options[m.Cursor])
			m.buildDependencyPlan()
			m.setScreen(ScreenDependencyTree)
			return m, nil
		}
		m.setScreen(ScreenPersona)
	case ScreenDependencyTree:
		if m.Selection.Preset == model.PresetCustom {
			allComps := screens.AllComponents()
			switch {
			case m.Cursor < len(allComps):
				m.toggleCurrentComponent()
			case m.Cursor == len(allComps):
				m.buildDependencyPlan()
				m.Review = planner.BuildReviewPayload(m.Selection, m.DependencyPlan)
				m.setScreen(ScreenReview)
			default:
				m.setScreen(ScreenPreset)
			}
			return m, nil
		}
		if m.Cursor == 0 {
			m.Review = planner.BuildReviewPayload(m.Selection, m.DependencyPlan)
			m.setScreen(ScreenReview)
			return m, nil
		}
		m.setScreen(ScreenPreset)
	case ScreenReview:
		if m.Cursor == 0 {
			return m.startInstalling()
		}
		m.setScreen(ScreenDependencyTree)
	case ScreenInstalling:
		if m.Progress.Done() {
			m.setScreen(ScreenComplete)
			return m, nil
		}
		// If no ExecuteFn, fall back to manual step-through for dev/tests.
		if m.ExecuteFn == nil && !m.pipelineRunning {
			m.Progress.Mark(m.Progress.Current, "succeeded")
			if m.Progress.Done() {
				m.setScreen(ScreenComplete)
			}
		}
	case ScreenComplete:
		return m, tea.Quit
	case ScreenBackups:
		if m.Cursor < len(m.Backups) {
			return m.restoreBackup(m.Backups[m.Cursor])
		}
		m.setScreen(ScreenWelcome)
	}

	return m, nil
}

// startInstalling initializes the progress state from the resolved plan and
// starts the pipeline execution in a goroutine if ExecuteFn is provided.
func (m Model) startInstalling() (tea.Model, tea.Cmd) {
	m.setScreen(ScreenInstalling)
	m.SpinnerFrame = 0

	// Build progress labels from the resolved plan.
	labels := buildProgressLabels(m.DependencyPlan)
	if len(labels) == 0 {
		// Fallback labels when the plan is empty (dev/test).
		labels = []string{
			"Install dependencies",
			"Configure selected agents",
			"Inject ecosystem components",
		}
	}

	m.Progress = NewProgressState(labels)
	m.Progress.Start(0)
	m.Progress.AppendLog("starting installation")

	if m.ExecuteFn == nil {
		// No real executor; fall back to manual step-through.
		return m, tickCmd()
	}

	m.pipelineRunning = true

	// Capture values for the goroutine closure.
	executeFn := m.ExecuteFn
	selection := m.Selection
	resolved := m.DependencyPlan
	detection := m.Detection

	return m, tea.Batch(tickCmd(), func() tea.Msg {
		onProgress := func(event pipeline.ProgressEvent) {
			// NOTE: ProgressFunc is called synchronously from the pipeline goroutine.
			// We cannot use p.Send() here because we don't have a reference to the
			// tea.Program. Instead, these events are collected in the ExecutionResult
			// and the PipelineDoneMsg handles the final state. For real-time updates,
			// we rely on the pipeline calling this synchronously from each step.
		}

		result := executeFn(selection, resolved, detection, onProgress)
		return PipelineDoneMsg{Result: result}
	})
}

// restoreBackup triggers a backup restore in a goroutine.
func (m Model) restoreBackup(manifest backup.Manifest) (tea.Model, tea.Cmd) {
	if m.RestoreFn == nil {
		m.Err = fmt.Errorf("restore not available")
		return m, nil
	}

	restoreFn := m.RestoreFn
	return m, func() tea.Msg {
		err := restoreFn(manifest)
		return BackupRestoreMsg{Err: err}
	}
}

// buildProgressLabels creates step labels from the resolved plan that match
// the step IDs the pipeline will produce.
func buildProgressLabels(resolved planner.ResolvedPlan) []string {
	labels := make([]string, 0, 2+len(resolved.Agents)+len(resolved.OrderedComponents)+1)

	labels = append(labels, "prepare:check-dependencies")
	labels = append(labels, "prepare:backup-snapshot")
	labels = append(labels, "apply:rollback-restore")

	for _, agent := range resolved.Agents {
		labels = append(labels, "agent:"+string(agent))
	}

	for _, component := range resolved.OrderedComponents {
		labels = append(labels, "component:"+string(component))
	}

	return labels
}

func (m Model) goBack() Model {
	previous, ok := PreviousScreen(m.Screen)
	if !ok {
		return m
	}

	m.setScreen(previous)
	return m
}

func (m *Model) setScreen(next Screen) {
	m.PreviousScreen = m.Screen
	m.Screen = next
	m.Cursor = 0
}

func (m Model) optionCount() int {
	switch m.Screen {
	case ScreenWelcome:
		return len(screens.WelcomeOptions())
	case ScreenDetection:
		return len(screens.DetectionOptions())
	case ScreenAgents:
		return len(screens.AgentOptions()) + 2
	case ScreenPersona:
		return len(screens.PersonaOptions()) + 1
	case ScreenPreset:
		return len(screens.PresetOptions()) + 1
	case ScreenDependencyTree:
		if m.Selection.Preset == model.PresetCustom {
			return len(screens.AllComponents()) + len(screens.DependencyTreeOptions())
		}
		return len(screens.DependencyTreeOptions())
	case ScreenReview:
		return len(screens.ReviewOptions())
	case ScreenInstalling:
		return 1
	case ScreenComplete:
		return 1
	case ScreenBackups:
		return len(m.Backups) + 1
	case ScreenSkillDiscovery:
		return 1
	default:
		return 0
	}
}

func (m *Model) toggleCurrentAgent() {
	options := screens.AgentOptions()
	if m.Cursor >= len(options) {
		return
	}

	agent := options[m.Cursor]
	for idx, selected := range m.Selection.Agents {
		if selected == agent {
			m.Selection.Agents = append(m.Selection.Agents[:idx], m.Selection.Agents[idx+1:]...)
			return
		}
	}

	m.Selection.Agents = append(m.Selection.Agents, agent)
}

func (m *Model) toggleCurrentComponent() {
	allComps := screens.AllComponents()
	if m.Cursor >= len(allComps) {
		return
	}

	compID := allComps[m.Cursor].ID
	for idx, selected := range m.Selection.Components {
		if selected == compID {
			m.Selection.Components = append(m.Selection.Components[:idx], m.Selection.Components[idx+1:]...)
			return
		}
	}

	m.Selection.Components = append(m.Selection.Components, compID)
}

func (m *Model) buildDependencyPlan() {
	resolved, err := planner.NewResolver(planner.MVPGraph()).Resolve(m.Selection)
	if err != nil {
		m.Err = err
		m.DependencyPlan = planner.ResolvedPlan{}
		return
	}

	m.DependencyPlan = resolved
}

func preselectedAgents(detection system.DetectionResult) []model.AgentID {
	selected := []model.AgentID{}
	for _, state := range detection.Configs {
		if !state.Exists {
			continue
		}

		switch strings.TrimSpace(state.Agent) {
		case string(model.AgentClaudeCode):
			selected = append(selected, model.AgentClaudeCode)
		case string(model.AgentOpenCode):
			selected = append(selected, model.AgentOpenCode)
		}
	}

	if len(selected) > 0 {
		return selected
	}

	agents := catalog.MVPAgents()
	selected = make([]model.AgentID, 0, len(agents))
	for _, agent := range agents {
		selected = append(selected, agent.ID)
	}

	return selected
}

func extractMissingDeps(detection system.DetectionResult) []screens.MissingDep {
	if detection.Dependencies.AllPresent {
		return nil
	}

	var deps []screens.MissingDep
	for _, dep := range detection.Dependencies.Dependencies {
		if !dep.Installed && dep.Required {
			deps = append(deps, screens.MissingDep{Name: dep.Name, InstallHint: dep.InstallHint})
		}
	}
	return deps
}

func extractFailedSteps(result pipeline.ExecutionResult) []screens.FailedStep {
	var failed []screens.FailedStep
	collect := func(steps []pipeline.StepResult) {
		for _, step := range steps {
			if step.Status == pipeline.StepStatusFailed {
				errMsg := "unknown error"
				if step.Err != nil {
					errMsg = step.Err.Error()
				}
				failed = append(failed, screens.FailedStep{ID: step.StepID, Error: errMsg})
			}
		}
	}
	collect(result.Prepare.Steps)
	collect(result.Apply.Steps)
	return failed
}

// startSkillDiscovery initialises the skill discovery flow.
// It guards against npx not being installed, detects the project stack,
// transitions to ScreenSkillDiscovery, and fires the async search command.
func (m Model) startSkillDiscovery() (tea.Model, tea.Cmd) {
	if _, err := exec.LookPath("npx"); err != nil {
		m.Err = fmt.Errorf("npx not found: install Node.js to use skill discovery")
		return m, nil
	}

	query := stackdetect.Detect(".")
	if query == "" {
		query = "developer tools"
	}

	m.skillDiscovery = skillDiscoveryState{
		SubState: screens.SkillDiscoveryLoading,
	}
	m.setScreen(ScreenSkillDiscovery)

	return m, SearchSkillsCmd(query)
}

// SearchSkillsCmd searches skills.sh for each keyword individually, merges and
// deduplicates results by skillId, and returns them sorted by installs desc.
// Searching per keyword instead of a combined query gives much better results.
func SearchSkillsCmd(query string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		keywords := strings.Fields(query)
		if len(keywords) == 0 {
			keywords = []string{"developer tools"}
		}

		seen := make(map[string]struct{})
		var merged []skillssh.SearchSkill

		for _, kw := range keywords {
			raw, err := skillssh.Search(ctx, kw)
			if err != nil {
				continue // skip failed keyword, try the rest
			}
			filtered := skillssh.Filter(raw)
			for _, s := range filtered {
				if _, exists := seen[s.ID]; !exists {
					seen[s.ID] = struct{}{}
					merged = append(merged, s)
				}
			}
		}

		// re-sort merged results by installs desc
		sort.Slice(merged, func(i, j int) bool {
			return merged[i].Installs > merged[j].Installs
		})

		const maxRecommended = 5
		if len(merged) > maxRecommended {
			merged = merged[:maxRecommended]
		}

		items := make([]screens.SkillDiscoveryItem, 0, len(merged))
		for _, s := range merged {
			items = append(items, screens.SkillDiscoveryItem{
				Name:     s.Name,
				SkillID:  s.SkillID,
				Source:   s.Source,
				Installs: s.Installs,
			})
		}

		return skillsLoadedMsg{Skills: items}
	}
}

// SearchManualCmd searches skills.sh for a single user-typed query, filters
// results, and returns at most 3 items sorted by installs descending.
func SearchManualCmd(query string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		raw, err := skillssh.Search(ctx, query)
		if err != nil {
			return manualSearchDoneMsg{Err: err}
		}

		filtered := skillssh.Filter(raw)

		const maxResults = 10
		if len(filtered) > maxResults {
			filtered = filtered[:maxResults]
		}

		items := make([]screens.SkillDiscoveryItem, 0, len(filtered))
		for _, s := range filtered {
			items = append(items, screens.SkillDiscoveryItem{
				Name:     s.Name,
				SkillID:  s.SkillID,
				Source:   s.Source,
				Installs: s.Installs,
			})
		}

		return manualSearchDoneMsg{Results: items}
	}
}

// buildInstallArg returns the `source@skillId` argument for `npx skills add`.
// Falls back to Name when SkillID is empty (should not happen in practice).
func buildInstallArg(skill screens.SkillDiscoveryItem) string {
	skillID := skill.SkillID
	if skillID == "" {
		skillID = skill.Name
	}
	return skill.Source + "@" + skillID
}

// InstallSkillCmd runs `npx skills add <source>@<skillId> -y` and returns a
// skillInstallDoneMsg when it finishes.
func InstallSkillCmd(skill screens.SkillDiscoveryItem, index int) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		cmd := exec.CommandContext(ctx, "npx", "skills", "add", buildInstallArg(skill), "-y")
		out, err := cmd.CombinedOutput()
		if err != nil {
			err = fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
		}
		return skillInstallDoneMsg{Index: index, Err: err}
	}
}

func extractAvailableUpdates(results []update.UpdateResult) []screens.UpdateInfo {
	var updates []screens.UpdateInfo
	for _, r := range results {
		if r.Status == update.UpdateAvailable {
			updates = append(updates, screens.UpdateInfo{
				Name:             r.Tool.Name,
				InstalledVersion: r.InstalledVersion,
				LatestVersion:    r.LatestVersion,
				UpdateHint:       r.UpdateHint,
			})
		}
	}
	return updates
}

func componentsForPreset(preset model.PresetID) []model.ComponentID {
	switch preset {
	case model.PresetMinimal:
		return []model.ComponentID{model.ComponentEngram}
	case model.PresetEcosystemOnly:
		return []model.ComponentID{model.ComponentEngram, model.ComponentSDD, model.ComponentSkills, model.ComponentContext7, model.ComponentGGA}
	case model.PresetCustom:
		return nil
	default:
		return []model.ComponentID{
			model.ComponentEngram,
			model.ComponentSDD,
			model.ComponentSkills,
			model.ComponentContext7,
			model.ComponentPersona,
			model.ComponentPermission,
			model.ComponentGGA,
		}
	}
}
