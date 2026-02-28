package screens

import (
	"fmt"
	"strings"

	"github.com/gentleman-programming/gentle-ai/internal/backup"
	"github.com/gentleman-programming/gentle-ai/internal/tui/styles"
)

func RenderBackups(backups []backup.Manifest, cursor int) string {
	var b strings.Builder

	b.WriteString(styles.TitleStyle.Render("Backup Management"))
	b.WriteString("\n\n")

	if len(backups) == 0 {
		b.WriteString(styles.WarningStyle.Render("No backups found yet."))
		b.WriteString("\n\n")
		b.WriteString(renderOptions([]string{"Back"}, 0))
		return b.String()
	}

	for idx, snapshot := range backups {
		label := fmt.Sprintf("%s (%s)", snapshot.ID, snapshot.CreatedAt.Format("2006-01-02 15:04"))
		focused := idx == cursor
		if focused {
			b.WriteString(styles.SelectedStyle.Render(styles.Cursor + label))
		} else {
			b.WriteString(styles.UnselectedStyle.Render("  " + label))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(renderOptions([]string{"Back"}, cursor-len(backups)))
	b.WriteString("\n")
	b.WriteString(styles.HelpStyle.Render("j/k: navigate • enter: select • esc: back"))

	return b.String()
}
