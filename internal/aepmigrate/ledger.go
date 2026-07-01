package aepmigrate

import (
	"fmt"

	"github.com/yueli-fx/aep-parser/internal/profile"
)

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
	explicitMatteRule, hasExplicitMatteRule := capabilityLedger.RuleByID("layer-explicit-matte-source")
	for _, comp := range prof.Comps {
		for _, layer := range comp.Layers {
			if hasExplicitMatteRule && explicitMatteRule.AppliesToSource(source) && isExplicitMatteRef(layer) && explicitMatteRule.TargetSupport[target] == CapabilityBlocked {
				entries = append(entries, Entry{
					Path:          fmt.Sprintf("comps[%q].layers[%q].matte_ref", comp.Name, layer.Name),
					Class:         ClassBlocked,
					TargetVersion: target,
					Reason:        explicitMatteRule.BlockedReason,
					CapabilityKey: explicitMatteRule.CapabilityKey,
				})
			}
		}
	}
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
