package catalog

import "github.com/gentleman-programming/gentle-ai/internal/model"

type Skill struct {
	ID       model.SkillID
	Name     string
	Category string
	Priority string
}

var mvpSkills = []Skill{
	// SDD skills
	{ID: model.SkillSDDInit, Name: "sdd-init", Category: "sdd", Priority: "p0"},

	{ID: model.SkillSDDApply, Name: "sdd-apply", Category: "sdd", Priority: "p0"},
	{ID: model.SkillSDDVerify, Name: "sdd-verify", Category: "sdd", Priority: "p0"},
	{ID: model.SkillSDDExplore, Name: "sdd-explore", Category: "sdd", Priority: "p0"},
	{ID: model.SkillSDDPropose, Name: "sdd-propose", Category: "sdd", Priority: "p0"},
	{ID: model.SkillSDDSpec, Name: "sdd-spec", Category: "sdd", Priority: "p0"},
	{ID: model.SkillSDDDesign, Name: "sdd-design", Category: "sdd", Priority: "p0"},
	{ID: model.SkillSDDTasks, Name: "sdd-tasks", Category: "sdd", Priority: "p0"},
	{ID: model.SkillSDDArchive, Name: "sdd-archive", Category: "sdd", Priority: "p0"},
	// Framework/coding skills
	{ID: model.SkillTypeScript, Name: "typescript", Category: "coding", Priority: "p0"},
	{ID: model.SkillReact19, Name: "react-19", Category: "coding", Priority: "p0"},
	{ID: model.SkillNextJS15, Name: "nextjs-15", Category: "coding", Priority: "p0"},
	{ID: model.SkillTailwind4, Name: "tailwind-4", Category: "coding", Priority: "p0"},
	{ID: model.SkillZustand5, Name: "zustand-5", Category: "coding", Priority: "p1"},
	{ID: model.SkillZod4, Name: "zod-4", Category: "coding", Priority: "p1"},
	{ID: model.SkillAISDK5, Name: "ai-sdk-5", Category: "coding", Priority: "p1"},
	{ID: model.SkillPlaywright, Name: "playwright", Category: "testing", Priority: "p1"},
	{ID: model.SkillPytest, Name: "pytest", Category: "testing", Priority: "p1"},
	{ID: model.SkillDjangoDRF, Name: "django-drf", Category: "coding", Priority: "p1"},
	{ID: model.SkillGoTesting, Name: "go-testing", Category: "testing", Priority: "p1"},
}

func MVPSkills() []Skill {
	skills := make([]Skill, len(mvpSkills))
	copy(skills, mvpSkills)
	return skills
}
