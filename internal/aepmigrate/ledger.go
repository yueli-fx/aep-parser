package aepmigrate

import (
	"fmt"

	"github.com/yueli-fx/aep-parser/internal/profile"
)

func classifyProfile(source, target VersionLabel, prof *profile.Profile) []Entry {
	var entries []Entry
	if prof == nil {
		return []Entry{{
			Path:          "project",
			Class:         ClassUnknown,
			TargetVersion: target,
			Reason:        "Profile is unavailable.",
		}}
	}
	for _, comp := range prof.Comps {
		for _, layer := range comp.Layers {
			if source == VersionAE2025 && isExplicitMatteRef(layer) && !supportsExplicitMatteTarget(target) {
				entries = append(entries, Entry{
					Path:          fmt.Sprintf("comps[%q].layers[%q].matte_ref", comp.Name, layer.Name),
					Class:         ClassBlocked,
					TargetVersion: target,
					Reason:        "Explicit matte source requires AE2025 in the current writer contract.",
					CapabilityKey: "layer.set_track_matte_source",
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
