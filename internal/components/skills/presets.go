package skills

type Preset struct {
	ID     string
	Skills []string
}

const (
	PresetFullStack = "full-stack"
	PresetMinimal   = "minimal"
)

func Presets() []Preset {
	return []Preset{
		{
			ID: PresetFullStack,
			Skills: []string{
				"typescript",
				"claude-developer-platform",
				"react-19",
				"nextjs-15",
				"tailwind-4",
				"zustand-5",
				"zod-4",
				"ai-sdk-5",
				"playwright",
				"pytest",
				"go-testing",
			},
		},
		{
			ID: PresetMinimal,
			Skills: []string{
				"typescript",
				"claude-developer-platform",
			},
		},
	}
}
