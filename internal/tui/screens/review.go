package screens

import (
	"strings"

	"github.com/gentleman-programming/gentle-ai/internal/model"
	"github.com/gentleman-programming/gentle-ai/internal/planner"
	"github.com/gentleman-programming/gentle-ai/internal/tui/styles"
)

func ReviewOptions() []string {
	return []string{"Install", "Back"}
}

func RenderReview(payload planner.ReviewPayload, cursor int) string {
	var b strings.Builder

	b.WriteString(styles.TitleStyle.Render("Review and Confirm"))
	b.WriteString("\n\n")

	agentsCard := styles.StatCardStyle.Render(
		styles.HeadingStyle.Render("Agents") + "\n" +
			styles.UnselectedStyle.Render(joinIDs(payload.Agents)),
	)
	personaCard := styles.StatCardStyle.Render(
		styles.HeadingStyle.Render("Persona") + "\n" +
			styles.UnselectedStyle.Render(string(payload.Persona)),
	)
	presetCard := styles.StatCardStyle.Render(
		styles.HeadingStyle.Render("Preset") + "\n" +
			styles.UnselectedStyle.Render(string(payload.Preset)),
	)

	b.WriteString(agentsCard + "  " + personaCard + "  " + presetCard)
	b.WriteString("\n\n")

	if len(payload.Components) > 0 {
		autoSet := make(map[model.ComponentID]struct{}, len(payload.AddedDependencies))
		for _, dep := range payload.AddedDependencies {
			autoSet[dep] = struct{}{}
		}

		b.WriteString(styles.HeadingStyle.Render("Components"))
		b.WriteString("\n")
		for _, comp := range payload.Components {
			badge := styles.SubtextStyle.Render("selected")
			if _, isAuto := autoSet[comp.ID]; isAuto {
				badge = styles.WarningStyle.Render("auto-dependency")
			}
			b.WriteString("  " + styles.UnselectedStyle.Render(string(comp.ID)) + " " + badge + "\n")
		}
		b.WriteString("\n")
	}

	if len(payload.UnsupportedAgents) > 0 {
		b.WriteString(styles.WarningStyle.Render("Unsupported agents: " + joinIDs(payload.UnsupportedAgents)))
		b.WriteString("\n\n")
	}

	b.WriteString(renderOptions(ReviewOptions(), cursor))
	b.WriteString("\n")
	b.WriteString(styles.HelpStyle.Render("enter: install • esc: back"))

	return b.String()
}

func joinIDs[T ~string](values []T) string {
	if len(values) == 0 {
		return "none"
	}

	parts := make([]string, 0, len(values))
	for _, value := range values {
		parts = append(parts, string(value))
	}

	return strings.Join(parts, ", ")
}
