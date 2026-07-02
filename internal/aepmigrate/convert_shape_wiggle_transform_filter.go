package aepmigrate

import (
	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

func materializeShapeWiggleTransform(shapeLayer *aep.ShapeLayer, source profile.Layer) error {
	wiggleTransform, err := shapeLayer.RootGroup().AddWiggleTransform()
	if err != nil {
		return err
	}
	transform := wiggleTransform.Transform()
	if value, ok := propertyVector(source.Properties, "ADBE Vector Wiggler Anchor", 2); ok {
		if err := transform.SetAnchor([2]float64{value[0], value[1]}); err != nil {
			return err
		}
	}
	if value, ok := propertyVector(source.Properties, "ADBE Vector Wiggler Position", 2); ok {
		if err := transform.SetPosition([2]float64{value[0], value[1]}); err != nil {
			return err
		}
	}
	if value, ok := propertyVector(source.Properties, "ADBE Vector Wiggler Scale", 2); ok {
		if err := transform.SetScale([2]float64{value[0], value[1]}); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Wiggler Rotation"); ok {
		if err := transform.SetRotation(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Xform Temporal Freq"); ok {
		if err := wiggleTransform.SetWigglesPerSecond(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Random Seed"); ok {
		if err := wiggleTransform.SetRandomSeed(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Correlation"); ok {
		if err := wiggleTransform.SetCorrelation(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Temporal Phase"); ok {
		if err := wiggleTransform.SetTemporalPhase(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Spatial Phase"); ok {
		if err := wiggleTransform.SetSpatialPhase(value); err != nil {
			return err
		}
	}
	return nil
}
