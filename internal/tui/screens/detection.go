package screens

import (
	"fmt"
	"sort"
	"strings"

	"github.com/gentleman-programming/gentle-ai/internal/system"
	"github.com/gentleman-programming/gentle-ai/internal/tui/styles"
)

func DetectionOptions() []string {
	return []string{"Continue", "Back"}
}

func RenderDetection(result system.DetectionResult, cursor int) string {
	var b strings.Builder

	b.WriteString(styles.TitleStyle.Render("System Detection"))
	b.WriteString("\n\n")

	osCard := styles.StatCardStyle.Render(
		styles.HeadingStyle.Render("OS") + "\n" +
			styles.UnselectedStyle.Render(fmt.Sprintf("%s (%s)", result.System.OS, result.System.Arch)),
	)
	shellCard := styles.StatCardStyle.Render(
		styles.HeadingStyle.Render("Shell") + "\n" +
			styles.UnselectedStyle.Render(result.System.Shell),
	)

	supportedText := styles.ErrorStyle.Render("No")
	if result.System.Supported {
		supportedText = styles.SuccessStyle.Render("Yes")
	}
	supportedCard := styles.StatCardStyle.Render(
		styles.HeadingStyle.Render("Supported") + "\n" + supportedText,
	)

	b.WriteString(osCard + "  " + shellCard + "  " + supportedCard)
	b.WriteString("\n\n")

	if len(result.Tools) > 0 {
		b.WriteString(styles.HeadingStyle.Render("Tools"))
		b.WriteString("\n")
		keys := make([]string, 0, len(result.Tools))
		for key := range result.Tools {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			status := result.Tools[key]
			indicator := styles.ErrorStyle.Render("not found")
			if status.Installed {
				indicator = styles.SuccessStyle.Render("found")
			}
			b.WriteString(fmt.Sprintf("  %s: %s\n", styles.UnselectedStyle.Render(key), indicator))
		}
		b.WriteString("\n")
	}

	if len(result.Configs) > 0 {
		b.WriteString(styles.HeadingStyle.Render("Detected Configs"))
		b.WriteString("\n")
		for _, config := range result.Configs {
			indicator := styles.ErrorStyle.Render("missing")
			if config.Exists {
				indicator = styles.SuccessStyle.Render("present")
			}
			b.WriteString(fmt.Sprintf("  %s: %s\n", styles.UnselectedStyle.Render(config.Agent), indicator))
		}
		b.WriteString("\n")
	}

	b.WriteString(renderOptions(DetectionOptions(), cursor))
	b.WriteString("\n")
	b.WriteString(styles.HelpStyle.Render("j/k: navigate • enter: select • esc: back"))

	return b.String()
}
