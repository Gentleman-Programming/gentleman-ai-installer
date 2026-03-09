package update

// UpdateStatus represents the outcome of a single tool version check.
type UpdateStatus string

const (
	UpToDate        UpdateStatus = "up-to-date"
	UpdateAvailable UpdateStatus = "update-available"
	NotInstalled    UpdateStatus = "not-installed"
	VersionUnknown  UpdateStatus = "version-unknown"
	CheckFailed     UpdateStatus = "check-failed"
)

// ToolInfo describes a managed tool that can be checked for updates.
type ToolInfo struct {
	Name          string   // human-readable name (e.g., "gentle-ai")
	Owner         string   // GitHub repository owner
	Repo          string   // GitHub repository name
	DetectCmd     []string // command to detect installed version; nil = use build var
	VersionPrefix string   // prefix to strip from version output (e.g., "v")
}

// UpdateResult holds the result of checking a single tool for updates.
type UpdateResult struct {
	Tool             ToolInfo
	InstalledVersion string
	LatestVersion    string
	Status           UpdateStatus
	ReleaseURL       string
	UpdateHint       string
	Err              error
}
