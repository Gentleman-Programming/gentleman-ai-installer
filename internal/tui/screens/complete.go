package screens

import (
	"fmt"
	"strings"

	"github.com/gentleman-programming/gentle-ai/internal/tui/styles"
)

func RenderComplete(configuredAgents int, installedComponents int) string {
	var b strings.Builder

	b.WriteString(styles.SuccessStyle.Render("Done! Your AI agents are ready."))
	b.WriteString("\n\n")

	agentsCard := styles.StatCardStyle.Render(
		styles.HeadingStyle.Render("Configured agents") + "\n" +
			styles.SuccessStyle.Render(fmt.Sprintf("%d", configuredAgents)),
	)
	componentsCard := styles.StatCardStyle.Render(
		styles.HeadingStyle.Render("Installed components") + "\n" +
			styles.SuccessStyle.Render(fmt.Sprintf("%d", installedComponents)),
	)

	b.WriteString(agentsCard + "  " + componentsCard)
	b.WriteString("\n\n")

	b.WriteString(styles.HeadingStyle.Render("Next steps"))
	b.WriteString("\n")
	b.WriteString(styles.UnselectedStyle.Render("  1. Set your API keys"))
	b.WriteString("\n")
	b.WriteString(styles.UnselectedStyle.Render("  2. Run your selected agent"))
	b.WriteString("\n")
	b.WriteString(styles.UnselectedStyle.Render("  3. Try /sdd-new my-feature"))
	b.WriteString("\n\n")

	b.WriteString(styles.HelpStyle.Render("Press Enter to exit."))

	return b.String()
}
