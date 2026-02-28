package planner

import "github.com/gentleman-programming/gentle-ai/internal/model"

type Graph struct {
	dependencies map[model.ComponentID][]model.ComponentID
}

func NewGraph(dependencies map[model.ComponentID][]model.ComponentID) Graph {
	normalized := make(map[model.ComponentID][]model.ComponentID, len(dependencies))
	for component, deps := range dependencies {
		copyDeps := make([]model.ComponentID, len(deps))
		copy(copyDeps, deps)
		normalized[component] = copyDeps
	}

	return Graph{dependencies: normalized}
}

func (g Graph) Has(component model.ComponentID) bool {
	_, ok := g.dependencies[component]
	return ok
}

func (g Graph) DependenciesOf(component model.ComponentID) []model.ComponentID {
	deps, ok := g.dependencies[component]
	if !ok {
		return nil
	}

	copyDeps := make([]model.ComponentID, len(deps))
	copy(copyDeps, deps)
	return copyDeps
}

func MVPGraph() Graph {
	return NewGraph(map[model.ComponentID][]model.ComponentID{
		model.ComponentEngram:     nil,
		model.ComponentSDD:        {model.ComponentEngram},
		model.ComponentSkills:     {model.ComponentSDD},
		model.ComponentContext7:   nil,
		model.ComponentPersona:    nil,
		model.ComponentPermission: nil,
		model.ComponentGGA:        nil,
		model.ComponentTheme:      nil,
	})
}
