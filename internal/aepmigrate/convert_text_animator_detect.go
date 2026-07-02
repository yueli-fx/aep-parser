package aepmigrate

import "github.com/yueli-fx/aep-parser/internal/profile"

func hasProfileTextAnimators(prof *profile.Profile) bool {
	if prof == nil {
		return false
	}
	for _, comp := range prof.Comps {
		for _, layer := range comp.Layers {
			if hasLayerTextAnimators(layer) {
				return true
			}
		}
	}
	return false
}

func hasLayerTextAnimators(layer profile.Layer) bool {
	if layer.Type != "text" {
		return false
	}
	for _, property := range layer.Properties {
		if isTextAnimatorValueProperty(property.MatchName) {
			return true
		}
	}
	return false
}

func isTextAnimatorValueProperty(matchName string) bool {
	switch matchName {
	case "ADBE Text Opacity",
		"ADBE Text Position 3D",
		"ADBE Text Scale 3D",
		"ADBE Text Rotation",
		"ADBE Text Fill Color",
		"ADBE Text Stroke Color",
		"ADBE Text Tracking Amount",
		"ADBE Text Character Offset",
		"ADBE Text Fill Opacity",
		"ADBE Text Stroke Opacity",
		"ADBE Text Stroke Width",
		"ADBE Text Skew",
		"ADBE Text Rotation X",
		"ADBE Text Rotation Y":
		return true
	default:
		return false
	}
}
