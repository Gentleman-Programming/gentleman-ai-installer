package screens

import (
	"strings"

	"github.com/gentleman-programming/gentle-ai/internal/model"
	"github.com/gentleman-programming/gentle-ai/internal/tui/styles"
)

func PresetOptions() []model.PresetID {
	return []model.PresetID{
		model.PresetFullGentleman,
		model.PresetEcosystemOnly,
		model.PresetMinimal,
		model.PresetCustom,
	}
}

func RenderPreset(selected model.PresetID, cursor int) string {
	var b strings.Builder

	b.WriteString(styles.TitleStyle.Render("Select Ecosystem Preset"))
	b.WriteString("\n\n")

	for idx, preset := range PresetOptions() {
		isSelected := preset == selected
		focused := idx == cursor
		b.WriteString(renderRadio(string(preset), isSelected, focused))
	}

	b.WriteString("\n")
	b.WriteString(renderOptions([]string{"Back"}, cursor-len(PresetOptions())))
	b.WriteString("\n")
	b.WriteString(styles.HelpStyle.Render("j/k: navigate • enter: select • esc: back"))

	return b.String()
}
