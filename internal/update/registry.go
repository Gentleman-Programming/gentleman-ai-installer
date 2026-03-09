package update

// Tools is the static registry of managed tools that can be checked for updates.
var Tools = []ToolInfo{
	{
		Name:          "gentle-ai",
		Owner:         "Gentleman-Programming",
		Repo:          "gentle-ai",
		DetectCmd:     nil, // version comes from build-time ldflags (app.Version)
		VersionPrefix: "v",
	},
	{
		Name:          "engram",
		Owner:         "Gentleman-Programming",
		Repo:          "engram",
		DetectCmd:     []string{"engram", "version"},
		VersionPrefix: "v",
	},
	{
		Name:          "gga",
		Owner:         "Gentleman-Programming",
		Repo:          "gentleman-guardian-angel",
		DetectCmd:     []string{"gga", "--version"},
		VersionPrefix: "v",
	},
}
