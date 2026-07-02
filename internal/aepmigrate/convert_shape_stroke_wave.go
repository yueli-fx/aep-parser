package aepmigrate

import (
	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

func materializeShapeStrokeWave(stroke *aep.StrokeNode, source profile.Layer) error {
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Taper Wave Amount"); ok {
		if err := stroke.Wave().SetAmount(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Taper Wavelength"); ok {
		if err := stroke.Wave().SetWavelength(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Taper Wave Phase"); ok {
		if err := stroke.Wave().SetPhase(value); err != nil {
			return err
		}
	}
	return nil
}
