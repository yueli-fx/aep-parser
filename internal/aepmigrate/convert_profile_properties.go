package aepmigrate

import "github.com/yueli-fx/aep-parser/internal/profile"

func hasTransformProperties(layer profile.Layer) bool {
	return hasProperty(layer, "ADBE Anchor Point") ||
		hasProperty(layer, "ADBE Position") ||
		hasProperty(layer, "ADBE Scale") ||
		hasProperty(layer, "ADBE Rotate Z") ||
		hasProperty(layer, "ADBE Opacity")
}

func hasProperty(layer profile.Layer, matchName string) bool {
	for _, property := range layer.Properties {
		if property.MatchName == matchName {
			return true
		}
	}
	return false
}

func hasStaticPropertyVector(layer profile.Layer, matchName string, want []float64) bool {
	for _, property := range layer.Properties {
		if property.MatchName != matchName {
			continue
		}
		return vectorEquals(property.StaticValue, want)
	}
	return false
}

func vectorEquals(value any, want []float64) bool {
	switch v := value.(type) {
	case []float64:
		if len(v) != len(want) {
			return false
		}
		for i := range want {
			if v[i] != want[i] {
				return false
			}
		}
		return true
	case []any:
		if len(v) != len(want) {
			return false
		}
		for i := range want {
			got, ok := v[i].(float64)
			if !ok || got != want[i] {
				return false
			}
		}
		return true
	case [2]float64:
		return len(want) == 2 && v[0] == want[0] && v[1] == want[1]
	case [3]float64:
		return len(want) == 3 && v[0] == want[0] && v[1] == want[1] && v[2] == want[2]
	default:
		return false
	}
}
