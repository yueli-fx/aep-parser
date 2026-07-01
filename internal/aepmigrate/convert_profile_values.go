package aepmigrate

import (
	"github.com/yueli-fx/aep-parser/internal/codec"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

func propertyVector(properties []profile.Property, matchName string, length int) ([]float64, bool) {
	for _, property := range properties {
		if property.MatchName != matchName {
			continue
		}
		return staticVector(property.StaticValue, length)
	}
	return nil, false
}

func propertyByMatchName(properties []profile.Property, matchName string) (profile.Property, bool) {
	for _, property := range properties {
		if property.MatchName == matchName {
			return property, true
		}
	}
	return profile.Property{}, false
}

func propertyFloat(properties []profile.Property, matchName string) (float64, bool) {
	for _, property := range properties {
		if property.MatchName != matchName {
			continue
		}
		return propertyFloatValue(property.StaticValue)
	}
	return 0, false
}

func propertyFloatValue(value any) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	default:
		return 0, false
	}
}

func propertyGradient(properties []profile.Property, matchName string) (*codec.Gradient, bool) {
	for _, property := range properties {
		if property.MatchName != matchName || property.Gradient == nil {
			continue
		}
		return cloneCodecGradient(property.Gradient), true
	}
	return nil, false
}

func cloneCodecGradient(g *codec.Gradient) *codec.Gradient {
	if g == nil {
		return nil
	}
	return &codec.Gradient{
		Version:    g.Version,
		ColorStops: append([]codec.GradientColorStop(nil), g.ColorStops...),
		AlphaStops: append([]codec.GradientAlphaStop(nil), g.AlphaStops...),
	}
}

func staticVector(value any, length int) ([]float64, bool) {
	var out []float64
	switch v := value.(type) {
	case []float64:
		out = append([]float64(nil), v...)
	case []any:
		out = make([]float64, 0, len(v))
		for _, item := range v {
			got, ok := item.(float64)
			if !ok {
				return nil, false
			}
			out = append(out, got)
		}
	case [2]float64:
		out = []float64{v[0], v[1]}
	case [3]float64:
		out = []float64{v[0], v[1], v[2]}
	case [4]float64:
		out = []float64{v[0], v[1], v[2], v[3]}
	default:
		return nil, false
	}
	if len(out) != length {
		return nil, false
	}
	return out, true
}

func profileARGBToRGBA(value []float64) [4]float64 {
	return [4]float64{
		colorByteToUnit(value[1]),
		colorByteToUnit(value[2]),
		colorByteToUnit(value[3]),
		colorByteToUnit(value[0]),
	}
}

func colorByteToUnit(value float64) float64 {
	if value > 1 {
		return value / 255
	}
	return value
}

func propertyVectorAtLeast(properties []profile.Property, matchName string, length int) ([]float64, bool) {
	for _, property := range properties {
		if property.MatchName != matchName {
			continue
		}
		return staticVectorAtLeast(property.StaticValue, length)
	}
	return nil, false
}

func staticVectorAtLeast(value any, length int) ([]float64, bool) {
	out, ok := staticVectorAny(value)
	if !ok || len(out) < length {
		return nil, false
	}
	return out, true
}

func staticVectorAny(value any) ([]float64, bool) {
	switch v := value.(type) {
	case []float64:
		return append([]float64(nil), v...), true
	case []any:
		out := make([]float64, 0, len(v))
		for _, item := range v {
			got, ok := item.(float64)
			if !ok {
				return nil, false
			}
			out = append(out, got)
		}
		return out, true
	case [2]float64:
		return []float64{v[0], v[1]}, true
	case [3]float64:
		return []float64{v[0], v[1], v[2]}, true
	case [4]float64:
		return []float64{v[0], v[1], v[2], v[3]}, true
	default:
		return nil, false
	}
}

func profileScaleToWriter(value float64) float64 {
	return value * 100
}

func profileOpacityToWriter(value float64) float64 {
	return value * 100
}

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
