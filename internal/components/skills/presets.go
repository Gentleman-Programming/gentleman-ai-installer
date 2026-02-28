package skills

import "github.com/gentleman-programming/gentle-ai/internal/model"

// sddSkills are the SDD orchestrator skills — always included.
var sddSkills = []model.SkillID{
	model.SkillSDDInit,
	model.SkillSDDExplore,
	model.SkillSDDPropose,
	model.SkillSDDSpec,
	model.SkillSDDDesign,
	model.SkillSDDTasks,
	model.SkillSDDApply,
	model.SkillSDDVerify,
	model.SkillSDDArchive,
}

// frameworkSkills are common coding skills for the "recommended" tier.
var frameworkSkills = []model.SkillID{
	model.SkillTypeScript,
	model.SkillReact19,
	model.SkillNextJS15,
	model.SkillTailwind4,
	model.SkillZustand5,
	model.SkillZod4,
}

// extraSkills round out the "full" preset with all remaining skills.
var extraSkills = []model.SkillID{
	model.SkillAISDK5,
	model.SkillPlaywright,
	model.SkillPytest,
	model.SkillDjangoDRF,
	model.SkillGoTesting,
}

// SkillsForPreset returns which skills should be installed for a given preset.
//
//   - "minimal" / PresetMinimal:       SDD skills only
//   - "ecosystem-only" / PresetEcosystemOnly: SDD + common framework skills
//   - "full-gentleman" / PresetFullGentleman: all available skills
//   - "custom" / PresetCustom:         empty (caller should provide explicit list)
func SkillsForPreset(preset model.PresetID) []model.SkillID {
	switch preset {
	case model.PresetMinimal:
		return copySkills(sddSkills)
	case model.PresetEcosystemOnly:
		return copySkills(append(sddSkills, frameworkSkills...))
	case model.PresetFullGentleman:
		all := make([]model.SkillID, 0, len(sddSkills)+len(frameworkSkills)+len(extraSkills))
		all = append(all, sddSkills...)
		all = append(all, frameworkSkills...)
		all = append(all, extraSkills...)
		return all
	case model.PresetCustom:
		return nil
	default:
		// Unknown preset — default to full.
		all := make([]model.SkillID, 0, len(sddSkills)+len(frameworkSkills)+len(extraSkills))
		all = append(all, sddSkills...)
		all = append(all, frameworkSkills...)
		all = append(all, extraSkills...)
		return all
	}
}

// AllSkillIDs returns every known skill ID.
func AllSkillIDs() []model.SkillID {
	all := make([]model.SkillID, 0, len(sddSkills)+len(frameworkSkills)+len(extraSkills))
	all = append(all, sddSkills...)
	all = append(all, frameworkSkills...)
	all = append(all, extraSkills...)
	return all
}

func copySkills(src []model.SkillID) []model.SkillID {
	dst := make([]model.SkillID, len(src))
	copy(dst, src)
	return dst
}
