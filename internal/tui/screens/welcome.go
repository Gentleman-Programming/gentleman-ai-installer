package screens

import (
	"strings"

	"github.com/gentleman-programming/gentle-ai/internal/tui/styles"
)

func WelcomeOptions(canUpdateAll bool) []string {
	options := []string{"Start installation", "Manage backups"}
	if canUpdateAll {
		options = append(options, "Update all")
	}
	return append(options, "Quit")
}

func RenderWelcome(cursor int, version string, statusBanner string, statusLevel string, updateBanner string, canUpdateAll bool) string {
	var b strings.Builder

	b.WriteString(styles.RenderLogo())
	b.WriteString("\n\n")
	b.WriteString(styles.SubtextStyle.Render(styles.Tagline(version)))
	b.WriteString("\n")

	if statusBanner != "" {
		switch statusLevel {
		case "success":
			b.WriteString(styles.SuccessStyle.Render(statusBanner))
		case "error":
			b.WriteString(styles.ErrorStyle.Render(statusBanner))
		default:
			b.WriteString(styles.WarningStyle.Render(statusBanner))
		}
		b.WriteString("\n")
	}

	if updateBanner != "" {
		b.WriteString(styles.WarningStyle.Render(updateBanner))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(styles.HeadingStyle.Render("Menu"))
	b.WriteString("\n\n")
	b.WriteString(renderOptions(WelcomeOptions(canUpdateAll), cursor))
	b.WriteString("\n")
	b.WriteString(styles.HelpStyle.Render("j/k: navigate • enter: select • q: quit"))

	return styles.FrameStyle.Render(b.String())
}
