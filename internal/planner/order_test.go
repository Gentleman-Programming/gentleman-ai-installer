package planner

import (
	"errors"
	"reflect"
	"testing"

	"github.com/gentleman-programming/gentle-ai/internal/model"
)

func TestTopologicalSortOrdersDependenciesFirst(t *testing.T) {
	deps := map[model.ComponentID][]model.ComponentID{
		model.ComponentSkills:   {model.ComponentSDD},
		model.ComponentSDD:      {model.ComponentEngram},
		model.ComponentEngram:   nil,
		model.ComponentPersona:  nil,
		model.ComponentContext7: nil,
	}

	ordered, err := TopologicalSort(deps)
	if err != nil {
		t.Fatalf("TopologicalSort() returned error: %v", err)
	}

	if !reflect.DeepEqual(ordered, []model.ComponentID{
		model.ComponentContext7,
		model.ComponentEngram,
		model.ComponentPersona,
		model.ComponentSDD,
		model.ComponentSkills,
	}) {
		t.Fatalf("TopologicalSort() order = %v", ordered)
	}
}

func TestTopologicalSortDetectsCycles(t *testing.T) {
	deps := map[model.ComponentID][]model.ComponentID{
		model.ComponentEngram: {model.ComponentSDD},
		model.ComponentSDD:    {model.ComponentEngram},
	}

	_, err := TopologicalSort(deps)
	if err == nil {
		t.Fatalf("TopologicalSort() expected cycle error")
	}

	if !errors.Is(err, ErrDependencyCycle) {
		t.Fatalf("TopologicalSort() error = %v, want ErrDependencyCycle", err)
	}
}
