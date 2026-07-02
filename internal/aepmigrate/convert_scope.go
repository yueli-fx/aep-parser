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
		entries = append(entries, Entry{
			Path:          "comps[" + comp.Name + "]",
			Class:         ClassRetargeted,
			TargetVersion: target,
			Reason:        "Composition and stable composition settings are recreated through the target AE project template.",
		})
		for _, layer := range comp.Layers {
			if isSupportedDefaultNullLayer(layer, footage) {
				entries = append(entries, Entry{
					Path:          "comps[" + comp.Name + "].layers[" + layer.Name + "]",
					Class:         ClassRetargeted,
					TargetVersion: target,
					Reason:        "Default null layer is recreated through the target AE project template.",
				})
				continue
			}
			if isSupportedDefaultSolidLayer(layer, footage) {
				entries = append(entries, Entry{
					Path:          "comps[" + comp.Name + "].layers[" + layer.Name + "]",
					Class:         ClassRetargeted,
					TargetVersion: target,
					Reason:        "Default solid layer is recreated through the target AE project template.",
				})
				continue
			}
			if isSupportedDefaultAdjustmentLayer(layer, footage) {
				entries = append(entries, Entry{
					Path:          "comps[" + comp.Name + "].layers[" + layer.Name + "]",
					Class:         ClassRetargeted,
					TargetVersion: target,
					Reason:        "Default adjustment layer is recreated through the target AE project template.",
				})
				continue
			}
			if isSupportedDefaultCameraLayer(layer) {
				entries = append(entries, Entry{
					Path:          "comps[" + comp.Name + "].layers[" + layer.Name + "]",
					Class:         ClassRetargeted,
					TargetVersion: target,
					Reason:        "Default camera layer is recreated through the target AE project template.",
				})
				continue
			}
			if isSupportedDefaultLightLayer(layer) {
				entries = append(entries, Entry{
					Path:          "comps[" + comp.Name + "].layers[" + layer.Name + "]",
					Class:         ClassRetargeted,
					TargetVersion: target,
					Reason:        "Default light layer is recreated through the target AE project template.",
				})
				continue
			}
			if isSupportedDefaultTextLayer(layer) {
				entries = append(entries, Entry{
					Path:          "comps[" + comp.Name + "].layers[" + layer.Name + "]",
					Class:         ClassRetargeted,
					TargetVersion: target,
					Reason:        "Default text layer is recreated through the target AE project template.",
				})
				continue
			}
			if isSupportedDefaultShapeLayer(layer) {
				entries = append(entries, Entry{
					Path:          "comps[" + comp.Name + "].layers[" + layer.Name + "]",
					Class:         ClassRetargeted,
					TargetVersion: target,
					Reason:        "Default empty shape layer is recreated through the target AE project template.",
				})
				continue
			}
			if isSupportedRectGraphicShapeLayer(layer) {
				entries = append(entries, Entry{
					Path:          "comps[" + comp.Name + "].layers[" + layer.Name + "]",
					Class:         ClassRetargeted,
					TargetVersion: target,
					Reason:        "Single parametric graphic shape layer is recreated from the stable profile shape properties.",
				})
				continue
			}
			if isSupportedDefaultPrecompLayer(layer, comps) {
				entries = append(entries, Entry{
					Path:          "comps[" + comp.Name + "].layers[" + layer.Name + "]",
					Class:         ClassRetargeted,
					TargetVersion: target,
					Reason:        "Default precomp layer is recreated through the target AE project template.",
				})
				continue
			}
			entries = append(entries, Entry{
				Path:          "comps[" + comp.Name + "].layers[" + layer.Name + "]",
				Class:         ClassBlocked,
				TargetVersion: target,
				Reason:        "This convert slice only reconstructs no-layer comps, default null layers, default solid layers, default adjustment layers, default camera layers, default light layers, default text layers, default empty shape layers, single parametric graphic/filter shape layers, and default precomp layers; refusing output to avoid silent layer loss.",
			})
		}
	}
	return entries
}
