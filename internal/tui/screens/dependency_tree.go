package screens

import (
	"fmt"
	"strings"

	"github.com/gentleman-programming/gentle-ai/internal/model"
	"github.com/gentleman-programming/gentle-ai/internal/planner"
	"github.com/gentleman-programming/gentle-ai/internal/tui/styles"
)

func DependencyTreeOptions() []string {
	return []string{"Continue", "Back"}
}

func RenderDependencyTree(plan planner.ResolvedPlan, cursor int) string {
	var b strings.Builder

	b.WriteString(styles.TitleStyle.Render("Dependency Tree"))
	b.WriteString("\n\n")

	if len(plan.OrderedComponents) == 0 {
		b.WriteString(styles.WarningStyle.Render("No components selected yet."))
		b.WriteString("\n\n")
	} else {
		b.WriteString(styles.HeadingStyle.Render("Install order"))
		b.WriteString("\n")

		autoSet := make(map[model.ComponentID]struct{}, len(plan.AddedDependencies))
		for _, auto := range plan.AddedDependencies {
			autoSet[auto] = struct{}{}
		}

		for idx, component := range plan.OrderedComponents {
			num := styles.SubtextStyle.Render(fmt.Sprintf("%d.", idx+1))
			name := styles.UnselectedStyle.Render(string(component))
			note := styles.SubtextStyle.Render("selected")
			if _, isAuto := autoSet[component]; isAuto {
				note = styles.WarningStyle.Render("auto-dependency")
			}
			b.WriteString(fmt.Sprintf("  %s %s %s\n", num, name, note))
		}
		b.WriteString("\n")
	}

	b.WriteString(renderOptions(DependencyTreeOptions(), cursor))
	b.WriteString("\n")
	b.WriteString(styles.HelpStyle.Render("j/k: navigate • enter: select • esc: back"))

	return b.String()
}
