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

func propertyVectorAtLeast(properties []profile.Property, matchName string, length int) ([]float64, bool) {
	for _, property := range properties {
		if property.MatchName != matchName {
			continue
		}
		return staticVectorAtLeast(property.StaticValue, length)
	}
	return nil, false
}
