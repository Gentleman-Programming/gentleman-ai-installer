package planner

import (
	"errors"
	"fmt"
	"slices"

	"github.com/gentleman-programming/gentle-ai/internal/model"
)

var ErrDependencyCycle = errors.New("dependency cycle detected")

func TopologicalSort(dependencies map[model.ComponentID][]model.ComponentID) ([]model.ComponentID, error) {
	nodes := make(map[model.ComponentID]struct{}, len(dependencies))
	inDegree := make(map[model.ComponentID]int, len(dependencies))
	children := make(map[model.ComponentID][]model.ComponentID, len(dependencies))

	for component, deps := range dependencies {
		nodes[component] = struct{}{}
		if _, ok := inDegree[component]; !ok {
			inDegree[component] = 0
		}

		for _, dep := range deps {
			nodes[dep] = struct{}{}
			inDegree[component]++
			children[dep] = append(children[dep], component)
			if _, ok := inDegree[dep]; !ok {
				inDegree[dep] = 0
			}
		}
	}

	queue := make([]model.ComponentID, 0, len(nodes))
	for node := range nodes {
		if inDegree[node] == 0 {
			queue = append(queue, node)
		}
	}
	slices.Sort(queue)

	ordered := make([]model.ComponentID, 0, len(nodes))
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		ordered = append(ordered, node)

		slices.Sort(children[node])
		for _, child := range children[node] {
			inDegree[child]--
			if inDegree[child] == 0 {
				queue = append(queue, child)
				slices.Sort(queue)
			}
		}
	}

	if len(ordered) != len(nodes) {
		return nil, fmt.Errorf("%w: unresolved graph", ErrDependencyCycle)
	}

	return ordered, nil
}
