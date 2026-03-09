package skillssh

import (
	"sort"
	"strings"
)

// Filter keeps only skills from TrustedSources with a valid SkillID,
// sorted by Installs descending. Always returns a non-nil slice.
func Filter(skills []SearchSkill) []SearchSkill {
	trusted := make([]SearchSkill, 0, len(skills))

	for _, s := range skills {
		if isTrusted(s.Source) && s.Installs >= MinInstalls && isInstallable(s.SkillID) {
			trusted = append(trusted, s)
		}
	}

	sort.Slice(trusted, func(i, j int) bool {
		return trusted[i].Installs > trusted[j].Installs
	})

	return trusted
}

// isTrusted reports whether the given source is in TrustedSources.
func isTrusted(source string) bool {
	for _, t := range TrustedSources {
		if t == source {
			return true
		}
	}
	return false
}

// isInstallable reports whether the skillId can be safely passed to
// `npx skills add source@skillId`. SkillIDs containing `:` break the
// underlying git clone (git treats `:` as a refspec separator).
func isInstallable(skillID string) bool {
	return skillID != "" && !strings.Contains(skillID, ":")
}
