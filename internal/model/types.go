package model

type AgentID string

const (
	AgentClaudeCode AgentID = "claude-code"
	AgentOpenCode   AgentID = "opencode"
)

type ComponentID string

const (
	ComponentEngram     ComponentID = "engram"
	ComponentSDD        ComponentID = "sdd"
	ComponentSkills     ComponentID = "skills"
	ComponentContext7   ComponentID = "context7"
	ComponentPersona    ComponentID = "persona"
	ComponentPermission ComponentID = "permissions"
)

type SkillID string

const (
	SkillSDDInit    SkillID = "sdd-init"
	SkillSDDNew     SkillID = "sdd-new"
	SkillSDDApply   SkillID = "sdd-apply"
	SkillSDDVerify  SkillID = "sdd-verify"
	SkillTypeScript SkillID = "typescript"
)

type PersonaID string

const (
	PersonaGentleman PersonaID = "gentleman"
	PersonaNeutral   PersonaID = "neutral"
	PersonaCustom    PersonaID = "custom"
)

type PresetID string

const (
	PresetFullGentleman PresetID = "full-gentleman"
	PresetEcosystemOnly PresetID = "ecosystem-only"
	PresetMinimal       PresetID = "minimal"
	PresetCustom        PresetID = "custom"
)
