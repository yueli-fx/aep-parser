package aepmigrate

import (
	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

func hasProfileEffects(prof *profile.Profile) bool {
	if prof == nil {
		return false
	}
	for _, comp := range prof.Comps {
		for _, layer := range comp.Layers {
			if len(layer.Effects) != 0 {
				return true
			}
		}
	}
	return false
}

func isSupportedStaticEffectSurface(layer profile.Layer) bool {
	for _, effect := range layer.Effects {
		if !isSupportedEffectMatchName(effect.MatchName) {
			return false
		}
		for _, param := range effect.Params {
			if param.LayerRef != nil {
				if len(param.Keyframes) != 0 || effectParamHasExpression(param) {
					return false
				}
				continue
			}
			if len(param.Keyframes) != 0 {
				if !isSupportedEffectParamKeyframes(param.Keyframes) {
					return false
				}
				continue
			}
			if param.StaticValue != nil {
				_, ok := effectParamStaticValue(param.StaticValue)
				if !ok {
					return false
				}
				continue
			}
			if effectParamHasExpression(param) || param.Changed {
				return false
			}
		}
	}
	return true
}

func effectParamHasExpression(param profile.Property) bool {
	return param.Expression != "" || param.ExpressionEnabled != nil
}

func isSupportedEffectParamKeyframes(keyframes []profile.Keyframe) bool {
	if len(keyframes) < 2 {
		return false
	}
	if _, ok := effectParamNumber(keyframes[0].Value); ok {
		_, err := effectParamScalarKeyframes(keyframes)
		return err == nil
	}
	_, err := effectParamVectorKeyframes(keyframes)
	return err == nil
}

func isSupportedEffectMatchName(matchName string) bool {
	for _, supported := range aep.SupportedEffects() {
		if supported == matchName {
			return true
		}
	}
	return false
}
