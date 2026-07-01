package aepmigrate

import "github.com/yueli-fx/aep-parser/internal/profile"

func isExplicitMatteRef(layer profile.Layer) bool {
	return layer.MatteRef != nil && layer.MatteRefKind == "explicit"
}

func supportsExplicitMatteTarget(target VersionLabel) bool {
	rule, ok := DefaultCapabilityLedger().RuleByID("layer-explicit-matte-source")
	return ok && rule.TargetSupport[target] == CapabilityPreserved
}
