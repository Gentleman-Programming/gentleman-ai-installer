package model

// ModelAssignment represents a provider/model pair assigned to an SDD phase sub-agent.
type ModelAssignment struct {
	ProviderID string // e.g., "anthropic"
	ModelID    string // e.g., "claude-sonnet-4-20250514"
}

// FullID returns the provider-qualified model identifier (e.g., "anthropic/claude-sonnet-4-20250514").
func (m ModelAssignment) FullID() string {
	return m.ProviderID + "/" + m.ModelID
}
