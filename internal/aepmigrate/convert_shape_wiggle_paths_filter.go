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
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Roughen Size"); ok {
		if err := wigglePaths.SetSize(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Roughen Detail"); ok {
		if err := wigglePaths.SetDetail(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Temporal Freq"); ok {
		if err := wigglePaths.SetWigglesPerSecond(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Random Seed"); ok {
		if err := wigglePaths.SetRandomSeed(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Roughen Points"); ok {
		if err := wigglePaths.SetPoints(aep.RoughenPoints(int(value))); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Correlation"); ok {
		if err := wigglePaths.SetCorrelation(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Temporal Phase"); ok {
		if err := wigglePaths.SetTemporalPhase(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Spatial Phase"); ok {
		if err := wigglePaths.SetSpatialPhase(value); err != nil {
			return err
		}
	}
	return nil
}
