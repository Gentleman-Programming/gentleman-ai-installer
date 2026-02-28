package model

type AgentID string

const (
	AgentClaudeCode    AgentID = "claude-code"
	AgentOpenCode      AgentID = "opencode"
	AgentGeminiCLI     AgentID = "gemini-cli"
	AgentCursor        AgentID = "cursor"
	AgentVSCodeCopilot AgentID = "vscode-copilot"
)

// SupportTier indicates how fully an agent supports the Gentleman AI ecosystem.
type SupportTier string

const (
	// TierFull supports sub-agent delegation, skills, MCP, and all features (Claude Code, OpenCode).
	TierFull SupportTier = "full"
	// TierGood supports inline skills, MCP, system prompts, but no sub-agent delegation (Gemini CLI).
	TierGood SupportTier = "good"
	// TierPartial supports system instructions and MCP, limited skill support (Cursor, VS Code Copilot).
	TierPartial SupportTier = "partial"
	// TierMinimal supports only persona via project rules.
	TierMinimal SupportTier = "minimal"
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

// SystemPromptStrategy defines how an agent's system prompt file is managed.
type SystemPromptStrategy int

const (
	// StrategyMarkdownSections uses <!-- gentle-ai:ID --> markers to inject sections
	// into an existing file without clobbering user content (Claude Code CLAUDE.md).
	StrategyMarkdownSections SystemPromptStrategy = iota
	// StrategyFileReplace replaces the entire system prompt file (OpenCode AGENTS.md).
	StrategyFileReplace
	// StrategyAppendToFile appends content to an existing system prompt file (Gemini CLI, Cursor).
	StrategyAppendToFile
)

// MCPStrategy defines how MCP server configs are written for an agent.
type MCPStrategy int

const (
	// StrategySeparateMCPFiles writes one JSON file per server in a dedicated directory
	// (e.g., ~/.claude/mcp/context7.json).
	StrategySeparateMCPFiles MCPStrategy = iota
	// StrategyMergeIntoSettings merges mcpServers into a settings.json file
	// (e.g., OpenCode, Gemini CLI).
	StrategyMergeIntoSettings
	// StrategyMCPConfigFile writes to a dedicated mcp.json config file (e.g., Cursor ~/.cursor/mcp.json).
	StrategyMCPConfigFile
)

type PresetID string

const (
	PresetFullGentleman PresetID = "full-gentleman"
	PresetEcosystemOnly PresetID = "ecosystem-only"
	PresetMinimal       PresetID = "minimal"
	PresetCustom        PresetID = "custom"
)
