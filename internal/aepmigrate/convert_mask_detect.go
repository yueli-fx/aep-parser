package aepmigrate

import "github.com/yueli-fx/aep-parser/internal/profile"

func hasProfileMasks(prof *profile.Profile) bool {
	if prof == nil {
		return false
	}
	for _, comp := range prof.Comps {
		for _, layer := range comp.Layers {
			if len(layer.Masks) != 0 {
				return true
			}
		}
	}
	return false
}

func isSupportedMaskSurface(masks []profile.Mask) bool {
	for _, mask := range masks {
		if len(mask.Vertices) == 0 {
			return false
		}
		if _, err := convertMaskMode(mask.Mode); mask.Mode != "" && err != nil {
			return false
		}
		if _, err := convertMaskMotionBlur(mask.MotionBlur); mask.MotionBlur != "" && err != nil {
			return false
		}
		if _, err := convertMaskFeatherFalloff(mask.FeatherFalloff); mask.FeatherFalloff != "" && err != nil {
			return false
		}
		if len(mask.Color) != 0 {
			if _, ok := profileRGBColor(mask.Color); !ok {
				return false
			}
		}
		if len(mask.Feather) != 0 && len(mask.Feather) != 2 {
			return false
		}
		if len(mask.PathKeyframes) == 1 {
			return false
		}
		for _, keyframe := range mask.PathKeyframes {
			if len(keyframe.Vertices) == 0 {
				return false
			}
		}
	}
	return true
}
