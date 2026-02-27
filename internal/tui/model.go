package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gentleman-programming/gentleman-ai-installer/internal/backup"
	"github.com/gentleman-programming/gentleman-ai-installer/internal/catalog"
	"github.com/gentleman-programming/gentleman-ai-installer/internal/model"
	"github.com/gentleman-programming/gentleman-ai-installer/internal/planner"
	"github.com/gentleman-programming/gentleman-ai-installer/internal/system"
	"github.com/gentleman-programming/gentleman-ai-installer/internal/tui/screens"
)

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
)

type Model struct {
	Screen         Screen
	PreviousScreen Screen
	Width          int
	Height         int
	Cursor         int

	Selection      model.Selection
	Detection      system.DetectionResult
	DependencyPlan planner.ResolvedPlan
	Review         planner.ReviewPayload
	Progress       ProgressState
	Backups        []backup.Manifest
	Err            error
}

func NewModel(detection system.DetectionResult) Model {
	selection := model.Selection{
		Agents:     preselectedAgents(detection),
		Persona:    model.PersonaGentleman,
		Preset:     model.PresetFullGentleman,
		Components: componentsForPreset(model.PresetFullGentleman),
	}

	return Model{
		Screen:    ScreenWelcome,
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
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil
	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	}

	return m, nil
}

func (m Model) View() string {
	switch m.Screen {
	case ScreenWelcome:
		return screens.RenderWelcome(m.Cursor, "v0.1.0")
	case ScreenDetection:
		return screens.RenderDetection(m.Detection, m.Cursor)
	case ScreenAgents:
		return screens.RenderAgents(m.Selection.Agents, m.Cursor)
	case ScreenPersona:
		return screens.RenderPersona(m.Selection.Persona, m.Cursor)
	case ScreenPreset:
		return screens.RenderPreset(m.Selection.Preset, m.Cursor)
	case ScreenDependencyTree:
		return screens.RenderDependencyTree(m.DependencyPlan, m.Cursor)
	case ScreenReview:
		return screens.RenderReview(m.Review, m.Cursor)
	case ScreenInstalling:
		return screens.RenderInstalling(m.Progress.ViewModel())
	case ScreenComplete:
		return screens.RenderComplete(len(m.Selection.Agents), len(m.Selection.Components))
	case ScreenBackups:
		return screens.RenderBackups(m.Backups, m.Cursor)
	default:
		return ""
	}
}

func (m Model) handleKeyPress(key tea.KeyMsg) (tea.Model, tea.Cmd) {
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
		return m.goBack(), nil
	case " ":
		if m.Screen == ScreenAgents {
			m.toggleCurrentAgent()
		}
		return m, nil
	case "enter":
		return m.confirmSelection()
	}

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
		if m.Cursor == 0 {
			m.Review = planner.BuildReviewPayload(m.Selection, m.DependencyPlan)
			m.setScreen(ScreenReview)
			return m, nil
		}
		m.setScreen(ScreenPreset)
	case ScreenReview:
		if m.Cursor == 0 {
			m.setScreen(ScreenInstalling)
			m.Progress.Start(0)
			m.Progress.AppendLog("starting installation")
			return m, nil
		}
		m.setScreen(ScreenDependencyTree)
	case ScreenInstalling:
		if m.Progress.Percent() >= 100 {
			m.setScreen(ScreenComplete)
			return m, nil
		}
		m.Progress.Mark(m.Progress.Current, "succeeded")
		if m.Progress.Percent() >= 100 {
			m.setScreen(ScreenComplete)
		}
	case ScreenComplete:
		return m, tea.Quit
	case ScreenBackups:
		m.setScreen(ScreenWelcome)
	}

	return m, nil
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
		return len(screens.DependencyTreeOptions())
	case ScreenReview:
		return len(screens.ReviewOptions())
	case ScreenInstalling:
		return 1
	case ScreenComplete:
		return 1
	case ScreenBackups:
		return len(m.Backups) + 1
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

func componentsForPreset(preset model.PresetID) []model.ComponentID {
	switch preset {
	case model.PresetMinimal:
		return []model.ComponentID{model.ComponentEngram}
	case model.PresetEcosystemOnly:
		return []model.ComponentID{model.ComponentEngram, model.ComponentSDD, model.ComponentSkills, model.ComponentContext7}
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
		}
	}
}
