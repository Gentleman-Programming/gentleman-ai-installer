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
	ComponentGGA        ComponentID = "gga"
	ComponentTheme      ComponentID = "theme"
)

type SkillID string

const (
	SkillSDDInit    SkillID = "sdd-init"
	SkillSDDNew     SkillID = "sdd-new"
	SkillSDDApply   SkillID = "sdd-apply"
	SkillSDDVerify  SkillID = "sdd-verify"
	SkillSDDExplore SkillID = "sdd-explore"
	SkillSDDPropose SkillID = "sdd-propose"
	SkillSDDSpec    SkillID = "sdd-spec"
	SkillSDDDesign  SkillID = "sdd-design"
	SkillSDDTasks   SkillID = "sdd-tasks"
	SkillSDDArchive SkillID = "sdd-archive"
	SkillTypeScript SkillID = "typescript"
	SkillReact19    SkillID = "react-19"
	SkillNextJS15   SkillID = "nextjs-15"
	SkillTailwind4  SkillID = "tailwind-4"
	SkillZustand5   SkillID = "zustand-5"
	SkillZod4       SkillID = "zod-4"
	SkillAISDK5     SkillID = "ai-sdk-5"
	SkillPlaywright SkillID = "playwright"
	SkillPytest     SkillID = "pytest"
	SkillDjangoDRF  SkillID = "django-drf"
	SkillGoTesting  SkillID = "go-testing"
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
