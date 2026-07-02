package aepmigrate

import (
	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

func materializeShapeStrokeWave(stroke *aep.StrokeNode, source profile.Layer) error {
	if err := applyShapeFloatProperty(source.Properties, "ADBE Vector Taper Wave Amount", stroke.Wave().SetAmount); err != nil {
		return err
	}
	if err := applyShapeFloatProperty(source.Properties, "ADBE Vector Taper Wavelength", stroke.Wave().SetWavelength); err != nil {
		return err
	}
	if err := applyShapeFloatProperty(source.Properties, "ADBE Vector Taper Wave Phase", stroke.Wave().SetPhase); err != nil {
		return err
	}
	return nil
}
