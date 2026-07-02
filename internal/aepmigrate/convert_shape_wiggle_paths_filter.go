package aepmigrate

import (
	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

func materializeShapeWigglePaths(shapeLayer *aep.ShapeLayer, source profile.Layer) error {
	wigglePaths, err := shapeLayer.RootGroup().AddWigglePaths()
	if err != nil {
		return err
	}
	if err := applyShapeFloatProperty(source.Properties, "ADBE Vector Roughen Size", wigglePaths.SetSize); err != nil {
		return err
	}
	if err := applyShapeFloatProperty(source.Properties, "ADBE Vector Roughen Detail", wigglePaths.SetDetail); err != nil {
		return err
	}
	if err := applyShapeFloatProperty(source.Properties, "ADBE Vector Temporal Freq", wigglePaths.SetWigglesPerSecond); err != nil {
		return err
	}
	if err := applyShapeFloatProperty(source.Properties, "ADBE Vector Random Seed", wigglePaths.SetRandomSeed); err != nil {
		return err
	}
	if err := applyShapeFloatProperty(source.Properties, "ADBE Vector Roughen Points", func(value float64) error {
		return wigglePaths.SetPoints(aep.RoughenPoints(int(value)))
	}); err != nil {
		return err
	}
	if err := applyShapeFloatProperty(source.Properties, "ADBE Vector Correlation", wigglePaths.SetCorrelation); err != nil {
		return err
	}
	if err := applyShapeFloatProperty(source.Properties, "ADBE Vector Temporal Phase", wigglePaths.SetTemporalPhase); err != nil {
		return err
	}
	if err := applyShapeFloatProperty(source.Properties, "ADBE Vector Spatial Phase", wigglePaths.SetSpatialPhase); err != nil {
		return err
	}
	return nil
}
