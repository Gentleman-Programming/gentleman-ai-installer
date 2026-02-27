package agents

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/gentleman-programming/gentleman-ai-installer/internal/model"
	"github.com/gentleman-programming/gentleman-ai-installer/internal/system"
)

type mockAdapter struct {
	agent model.AgentID
}

func (m mockAdapter) Agent() model.AgentID {
	return m.agent
}

func (m mockAdapter) SupportsAutoInstall() bool {
	return true
}

func (m mockAdapter) Detect(_ context.Context, _ string) (bool, string, string, bool, error) {
	return false, "", "", false, nil
}

func (m mockAdapter) InstallCommand(system.PlatformProfile) ([]string, error) {
	return nil, nil
}

func TestRegistrySupportedAgentsSorted(t *testing.T) {
	r, err := NewRegistry(
		mockAdapter{agent: model.AgentOpenCode},
		mockAdapter{agent: model.AgentClaudeCode},
	)
	if err != nil {
		t.Fatalf("NewRegistry() returned error: %v", err)
	}

	if !reflect.DeepEqual(r.SupportedAgents(), []model.AgentID{model.AgentClaudeCode, model.AgentOpenCode}) {
		t.Fatalf("SupportedAgents() = %v", r.SupportedAgents())
	}
}

func TestRegistryRejectsDuplicateAgent(t *testing.T) {
	_, err := NewRegistry(
		mockAdapter{agent: model.AgentClaudeCode},
		mockAdapter{agent: model.AgentClaudeCode},
	)
	if err == nil {
		t.Fatalf("NewRegistry() expected duplicate error")
	}

	if !errors.Is(err, ErrDuplicateAdapter) {
		t.Fatalf("NewRegistry() error = %v, want ErrDuplicateAdapter", err)
	}
}

func TestFactoryReturnsMVPAdapters(t *testing.T) {
	registry, err := NewMVPRegistry()
	if err != nil {
		t.Fatalf("NewMVPRegistry() returned error: %v", err)
	}

	if _, ok := registry.Get(model.AgentClaudeCode); !ok {
		t.Fatalf("registry missing claude adapter")
	}

	if _, ok := registry.Get(model.AgentOpenCode); !ok {
		t.Fatalf("registry missing opencode adapter")
	}
}

func TestFactoryRejectsUnsupportedAgent(t *testing.T) {
	_, err := NewAdapter(model.AgentID("cursor"))
	if err == nil {
		t.Fatalf("NewAdapter() expected unsupported agent error")
	}

	if !errors.Is(err, ErrAgentNotSupported) {
		t.Fatalf("NewAdapter() error = %v, want ErrAgentNotSupported", err)
	}
}
