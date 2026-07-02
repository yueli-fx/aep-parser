package aepmigrate

import "github.com/yueli-fx/aep-parser/internal/profile"

func convertCompScopeEntry(target VersionLabel, comp profile.Composition) Entry {
	return Entry{
		Path:          "comps[" + comp.Name + "]",
		Class:         ClassRetargeted,
		TargetVersion: target,
		Reason:        "Composition and stable composition settings are recreated through the target AE project template.",
	}
}

func convertLayerScopeEntry(target VersionLabel, compName string, layer profile.Layer, footage convertFootageIndex, comps convertCompIndex) Entry {
	path := "comps[" + compName + "].layers[" + layer.Name + "]"
	if isSupportedDefaultNullLayer(layer, footage) {
		return convertRetargetedScopeEntry(target, path, "Default null layer is recreated through the target AE project template.")
	}
	if isSupportedDefaultSolidLayer(layer, footage) {
		return convertRetargetedScopeEntry(target, path, "Default solid layer is recreated through the target AE project template.")
	}
	if isSupportedDefaultAdjustmentLayer(layer, footage) {
		return convertRetargetedScopeEntry(target, path, "Default adjustment layer is recreated through the target AE project template.")
	}
	if isSupportedDefaultCameraLayer(layer) {
		return convertRetargetedScopeEntry(target, path, "Default camera layer is recreated through the target AE project template.")
	}
	if isSupportedDefaultLightLayer(layer) {
		return convertRetargetedScopeEntry(target, path, "Default light layer is recreated through the target AE project template.")
	}
	if isSupportedDefaultTextLayer(layer) {
		return convertRetargetedScopeEntry(target, path, "Default text layer is recreated through the target AE project template.")
	}
	if isSupportedDefaultShapeLayer(layer) {
		return convertRetargetedScopeEntry(target, path, "Default empty shape layer is recreated through the target AE project template.")
	}
	if isSupportedRectGraphicShapeLayer(layer) {
		return convertRetargetedScopeEntry(target, path, "Single parametric graphic shape layer is recreated from the stable profile shape properties.")
	}
	if isSupportedDefaultPrecompLayer(layer, comps) {
		return convertRetargetedScopeEntry(target, path, "Default precomp layer is recreated through the target AE project template.")
	}
	return Entry{
		Path:          path,
		Class:         ClassBlocked,
		TargetVersion: target,
		Reason:        "This convert slice only reconstructs no-layer comps, default null layers, default solid layers, default adjustment layers, default camera layers, default light layers, default text layers, default empty shape layers, single parametric graphic/filter shape layers, and default precomp layers; refusing output to avoid silent layer loss.",
	}
}

func convertRetargetedScopeEntry(target VersionLabel, path, reason string) Entry {
	return Entry{
		Path:          path,
		Class:         ClassRetargeted,
		TargetVersion: target,
		Reason:        reason,
	}
}
