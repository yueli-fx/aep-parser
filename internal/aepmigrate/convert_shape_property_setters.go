package aepmigrate

import (
	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

type shapeGradientStops interface {
	SetColorStops([]aep.GradientColorStop) error
	SetAlphaStops([]aep.GradientAlphaStop) error
}

func applyShapeFloatProperty(properties []profile.Property, matchName string, set func(float64) error) error {
	value, ok := propertyFloat(properties, matchName)
	if !ok {
		return nil
	}
	return set(value)
}

func applyShapeVector2Property(properties []profile.Property, matchName string, set func([2]float64) error) error {
	value, ok := propertyVector(properties, matchName, 2)
	if !ok {
		return nil
	}
	return set([2]float64{value[0], value[1]})
}

func applyShapeARGBProperty(properties []profile.Property, matchName string, set func([4]float64) error) error {
	value, ok := propertyVector(properties, matchName, 4)
	if !ok {
		return nil
	}
	return set(profileARGBToRGBA(value))
}

func applyShapeGradientStops(properties []profile.Property, target shapeGradientStops) error {
	gradient, ok := propertyGradient(properties, "ADBE Vector Grad Colors")
	if !ok {
		return nil
	}
	if len(gradient.ColorStops) > 0 {
		if err := target.SetColorStops(gradient.ColorStops); err != nil {
			return err
		}
	}
	if len(gradient.AlphaStops) > 0 {
		if err := target.SetAlphaStops(gradient.AlphaStops); err != nil {
			return err
		}
	}
	return nil
}
