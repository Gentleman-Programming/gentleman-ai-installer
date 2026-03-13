package screens

import (
	"fmt"
	"strings"

	"github.com/gentleman-programming/gentle-ai/internal/tui/styles"
)

// SkillDiscoverySubState represents the current sub-state of the discovery screen.
type SkillDiscoverySubState int

const (
	SkillDiscoveryLoading       SkillDiscoverySubState = iota
	SkillDiscoveryBrowsing
	SkillDiscoverySearchPrompt
	SkillDiscoverySearchResults
	SkillDiscoveryInstalling
	SkillDiscoveryDone
)

// SkillDiscoveryItem is a plain-data representation of a single skill.
// The screens package does NOT import skillssh — it uses this local type.
type SkillDiscoveryItem struct {
	Name     string
	SkillID  string // ID from skills.sh API — used for installation
	Source   string
	Installs int
}

// SkillDiscoveryViewState is the plain-data projection passed from the model
// to the renderer. All state lives in the model; this struct is built on every
// View() call.
type SkillDiscoveryViewState struct {
	SubState     SkillDiscoverySubState
	Skills       []SkillDiscoveryItem // all filtered skills
	CurrentIndex int                  // which skill is being shown while browsing
	Selected     []SkillDiscoveryItem // skills the user said yes to
	Installing   int                  // index of skill currently being installed
	Error        string               // non-fatal error message (empty if none)
	Done         bool
	Failures     map[int]string // index in Selected → error message

	// Search prompt / results sub-state fields.
	SearchQuery         string
	SearchResults       []SkillDiscoveryItem
	SearchSelected      map[int]bool
	SearchResultCursor  int // highlighted row in SearchResults
}

// RenderSkillDiscovery renders the skill discovery screen.
//
// Sub-states:
//   - Loading       — spinner + "Searching for skills..."
//   - Browsing      — show current skill (name, source, installs), [↑] add [→] skip [q] quit
//   - SearchPrompt  — text input to search for more skills
//   - SearchResults — shows top 3 results from manual search
//   - Installing    — "Installing X of Y: {name}..." progress
//   - Done          — summary of installed skills (or "No skills installed")
func RenderSkillDiscovery(vs SkillDiscoveryViewState) string {
	var b strings.Builder

	b.WriteString(styles.TitleStyle.Render("Skill Discovery"))
	b.WriteString("\n\n")

	if vs.Error != "" {
		b.WriteString(styles.ErrorStyle.Render("⚠  " + vs.Error))
		b.WriteString("\n\n")
	}

	switch vs.SubState {
	case SkillDiscoveryLoading:
		renderSkillDiscoveryLoading(&b)
	case SkillDiscoveryBrowsing:
		renderSkillDiscoveryBrowsing(&b, vs)
	case SkillDiscoverySearchPrompt:
		renderSkillDiscoverySearchPrompt(&b, vs)
	case SkillDiscoverySearchResults:
		renderSkillDiscoverySearchResults(&b, vs)
	case SkillDiscoveryInstalling:
		renderSkillDiscoveryInstalling(&b, vs)
	case SkillDiscoveryDone:
		renderSkillDiscoveryDone(&b, vs)
	}

	return b.String()
}

func renderSkillDiscoveryLoading(b *strings.Builder) {
	b.WriteString(styles.SubtextStyle.Render("Searching for skills..."))
	b.WriteString("\n")
}

func renderSkillDiscoveryBrowsing(b *strings.Builder, vs SkillDiscoveryViewState) {
	if len(vs.Skills) == 0 {
		b.WriteString(styles.WarningStyle.Render("No skills found for the detected stack."))
		b.WriteString("\n\n")
		b.WriteString(styles.HelpStyle.Render("q: quit"))
		return
	}

	total := len(vs.Skills)
	idx := vs.CurrentIndex
	if idx < 0 || idx >= total {
		idx = 0
	}

	skill := vs.Skills[idx]

	b.WriteString(styles.SubtextStyle.Render(fmt.Sprintf("Skill %d of %d", idx+1, total)))
	b.WriteString("\n\n")

	b.WriteString(styles.PanelStyle.Render(
		styles.HeadingStyle.Render(skill.Name) + "\n" +
			styles.SubtextStyle.Render("Source:   "+skill.Source) + "\n" +
			styles.SubtextStyle.Render(fmt.Sprintf("Installs: %d", skill.Installs)),
	))
	b.WriteString("\n\n")

	if len(vs.Selected) > 0 {
		b.WriteString(styles.SuccessStyle.Render(fmt.Sprintf("Selected: %d skill(s)", len(vs.Selected))))
		b.WriteString("\n\n")
	}

	b.WriteString(styles.HelpStyle.Render("↑: add  →: skip  q: finish"))
}

func renderSkillDiscoverySearchPrompt(b *strings.Builder, vs SkillDiscoveryViewState) {
	b.WriteString(styles.HeadingStyle.Render("Search for more skills"))
	b.WriteString("\n\n")

	if len(vs.Selected) > 0 {
		b.WriteString(styles.SuccessStyle.Render(fmt.Sprintf("%d skill(s) ready to install", len(vs.Selected))))
		b.WriteString("\n\n")
	}

	b.WriteString(styles.SubtextStyle.Render("> " + vs.SearchQuery + "█"))
	b.WriteString("\n\n")

	if vs.SearchQuery == "" {
		b.WriteString(styles.HelpStyle.Render("type to search  ↵: install all  esc: install all"))
	} else {
		b.WriteString(styles.HelpStyle.Render("↵: search  esc: install all"))
	}
}

func renderSkillDiscoverySearchResults(b *strings.Builder, vs SkillDiscoveryViewState) {
	b.WriteString(styles.HeadingStyle.Render("Search Results"))
	b.WriteString("\n\n")

	if len(vs.SearchResults) == 0 {
		b.WriteString(styles.WarningStyle.Render("No results found for your query."))
		b.WriteString("\n\n")
		b.WriteString(styles.HelpStyle.Render("esc: back to search"))
		return
	}

	for i, skill := range vs.SearchResults {
		cursor := "  "
		if i == vs.SearchResultCursor {
			cursor = styles.SuccessStyle.Render("▸ ")
		}
		line := fmt.Sprintf("%s%-32s %s  %s installs",
			cursor,
			skill.Name,
			skill.Source,
			formatInstalls(skill.Installs),
		)
		b.WriteString(styles.SubtextStyle.Render(line))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(styles.HelpStyle.Render("↑/↓: navigate  ↵: add  esc: back without adding"))
}

// formatInstalls formats an install count with comma separators.
func formatInstalls(n int) string {
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}
	var result []byte
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, byte(c))
	}
	return string(result)
}

func renderSkillDiscoveryInstalling(b *strings.Builder, vs SkillDiscoveryViewState) {
	total := len(vs.Selected)
	if total == 0 {
		b.WriteString(styles.SubtextStyle.Render("Nothing to install."))
		return
	}

	current := vs.Installing
	if current < 0 {
		current = 0
	}
	if current >= total {
		current = total - 1
	}

	skill := vs.Selected[current]

	b.WriteString(styles.SubtextStyle.Render(
		fmt.Sprintf("Installing %d of %d: %s...", current+1, total, skill.Name),
	))
	b.WriteString("\n\n")

	// Simple progress bar.
	const barWidth = 30
	filled := 0
	if total > 0 {
		filled = (current * barWidth) / total
	}
	bar := styles.ProgressFilled.Render(strings.Repeat("█", filled)) +
		styles.ProgressEmpty.Render(strings.Repeat("░", barWidth-filled))
	pct := 0
	if total > 0 {
		pct = (current * 100) / total
	}
	b.WriteString(bar + " " + styles.PercentStyle.Render(fmt.Sprintf("%d%%", pct)))
	b.WriteString("\n")
}

func renderSkillDiscoveryDone(b *strings.Builder, vs SkillDiscoveryViewState) {
	if len(vs.Selected) == 0 {
		b.WriteString(styles.SubtextStyle.Render("No skills installed."))
		b.WriteString("\n\n")
		b.WriteString(styles.HelpStyle.Render("enter: continue"))
		return
	}

	successCount := len(vs.Selected) - len(vs.Failures)
	if successCount > 0 {
		b.WriteString(styles.SuccessStyle.Render(fmt.Sprintf("Installed %d skill(s):", successCount)))
		b.WriteString("\n\n")
		for i, skill := range vs.Selected {
			if _, failed := vs.Failures[i]; !failed {
				b.WriteString(styles.SuccessStyle.Render("  ✓ " + skill.Name))
				b.WriteString("\n")
			}
		}
		b.WriteString("\n")
	}

	if len(vs.Failures) > 0 {
		b.WriteString(styles.ErrorStyle.Render(fmt.Sprintf("Failed %d skill(s):", len(vs.Failures))))
		b.WriteString("\n\n")
		for i, errMsg := range vs.Failures {
			if i < len(vs.Selected) {
				b.WriteString(styles.ErrorStyle.Render("  ✗ " + vs.Selected[i].Name + ": " + errMsg))
				b.WriteString("\n")
			}
		}
		b.WriteString("\n")
	}

	b.WriteString(styles.HelpStyle.Render("enter: continue"))
}
