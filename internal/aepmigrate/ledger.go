package aepmigrate

import "github.com/yueli-fx/aep-parser/internal/profile"

func classifyProfile(source, target VersionLabel, prof *profile.Profile) []Entry {
	return classifyProfileWithLedger(source, target, prof, DefaultCapabilityLedger())
}

func classifyProfileWithLedger(source, target VersionLabel, prof *profile.Profile, capabilityLedger VersionCapabilityLedger) []Entry {
	var entries []Entry
	if prof == nil {
		return []Entry{{
			Path:          "project",
			Class:         ClassUnknown,
			TargetVersion: target,
			Reason:        "Profile is unavailable.",
		}}
	}
	entries = append(entries, capabilityBlockerEntries(source, target, prof, capabilityLedger)...)
	if len(entries) == 0 {
		entries = append(entries, Entry{
			Path:          "project",
			Class:         ClassPreserved,
			TargetVersion: target,
			Reason:        "No version-specific blockers detected in the first assess slice.",
		})
	}
	return entries
}
