package catalog

import "github.com/gentleman-programming/gentle-ai/internal/model"

type Skill struct {
	ID       model.SkillID
	Name     string
	Category string
	Priority string
}

var mvpSkills = []Skill{
	{ID: model.SkillSDDInit, Name: "sdd-init", Category: "sdd", Priority: "p0"},
	{ID: model.SkillSDDNew, Name: "sdd-new", Category: "sdd", Priority: "p0"},
	{ID: model.SkillSDDApply, Name: "sdd-apply", Category: "sdd", Priority: "p0"},
	{ID: model.SkillSDDVerify, Name: "sdd-verify", Category: "sdd", Priority: "p0"},
	{ID: model.SkillTypeScript, Name: "typescript", Category: "coding", Priority: "p0"},
}

func MVPSkills() []Skill {
	skills := make([]Skill, len(mvpSkills))
	copy(skills, mvpSkills)
	return skills
}
