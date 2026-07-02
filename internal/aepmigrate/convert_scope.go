package aepmigrate

import (
	"github.com/yueli-fx/aep-parser/internal/profile"
)

func convertAssessEntries(entries []Entry) []Entry {
	out := make([]Entry, 0, len(entries))
	for _, entry := range entries {
		if entry.Path == "project" && entry.Class == ClassPreserved && entry.CapabilityKey == "" {
			continue
		}
		out = append(out, entry)
	}
	return out
}

func convertScopeEntries(target VersionLabel, prof *profile.Profile) []Entry {
	if prof == nil {
		return nil
	}
	var entries []Entry
	footage := newConvertFootageIndex(prof)
	comps := newConvertCompIndex(prof)
	for _, comp := range prof.Comps {
		entries = append(entries, convertCompScopeEntry(target, comp))
		for _, layer := range comp.Layers {
			entries = append(entries, convertLayerScopeEntry(target, comp.Name, layer, footage, comps))
		}
	}
	return entries
}
