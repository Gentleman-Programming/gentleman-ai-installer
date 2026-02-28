package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gentleman-programming/gentleman-ai-installer/internal/model"
	"github.com/gentleman-programming/gentleman-ai-installer/internal/system"
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

func screensAgentOptions() []model.AgentID {
	return []model.AgentID{model.AgentClaudeCode, model.AgentOpenCode}
}
