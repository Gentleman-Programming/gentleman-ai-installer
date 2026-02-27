package planner

import (
	"reflect"
	"testing"

	"github.com/gentleman-programming/gentleman-ai-installer/internal/model"
)

func TestResolverAddsMissingDependenciesInOrder(t *testing.T) {
	resolver := NewResolver(MVPGraph())

	selection := model.Selection{
		Agents:     []model.AgentID{model.AgentClaudeCode, model.AgentOpenCode},
		Components: []model.ComponentID{model.ComponentSkills},
	}

	plan, err := resolver.Resolve(selection)
	if err != nil {
		t.Fatalf("Resolve() returned error: %v", err)
	}

	if !reflect.DeepEqual(plan.Agents, []model.AgentID{model.AgentClaudeCode, model.AgentOpenCode}) {
		t.Fatalf("Resolve() agents = %v", plan.Agents)
	}

	if !reflect.DeepEqual(plan.OrderedComponents, []model.ComponentID{model.ComponentEngram, model.ComponentSDD, model.ComponentSkills}) {
		t.Fatalf("Resolve() ordered components = %v", plan.OrderedComponents)
	}

	if !reflect.DeepEqual(plan.AddedDependencies, []model.ComponentID{model.ComponentEngram, model.ComponentSDD}) {
		t.Fatalf("Resolve() added dependencies = %v", plan.AddedDependencies)
	}
}

func TestResolverExcludesUnsupportedAgents(t *testing.T) {
	resolver := NewResolver(MVPGraph())

	selection := model.Selection{
		Agents: []model.AgentID{model.AgentClaudeCode, model.AgentID("cursor")},
	}

	plan, err := resolver.Resolve(selection)
	if err != nil {
		t.Fatalf("Resolve() returned error: %v", err)
	}

	if !reflect.DeepEqual(plan.Agents, []model.AgentID{model.AgentClaudeCode}) {
		t.Fatalf("Resolve() agents = %v", plan.Agents)
	}

	if !reflect.DeepEqual(plan.UnsupportedAgents, []model.AgentID{model.AgentID("cursor")}) {
		t.Fatalf("Resolve() unsupported agents = %v", plan.UnsupportedAgents)
	}
}
