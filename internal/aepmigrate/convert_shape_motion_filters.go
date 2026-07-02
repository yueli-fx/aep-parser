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

func materializeShapeRepeater(shapeLayer *aep.ShapeLayer, source profile.Layer) error {
	repeater, err := shapeLayer.RootGroup().AddRepeater()
	if err != nil {
		return err
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Repeater Copies"); ok {
		if err := repeater.SetCopies(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Repeater Offset"); ok {
		if err := repeater.SetOffset(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Repeater Order"); ok {
		if err := repeater.SetOrder(aep.RepeaterOrder(int(value))); err != nil {
			return err
		}
	}
	transform := repeater.Transform()
	if value, ok := propertyVector(source.Properties, "ADBE Vector Repeater Anchor", 2); ok {
		if err := transform.SetAnchor([2]float64{value[0], value[1]}); err != nil {
			return err
		}
	}
	if value, ok := propertyVector(source.Properties, "ADBE Vector Repeater Position", 2); ok {
		if err := transform.SetPosition([2]float64{value[0], value[1]}); err != nil {
			return err
		}
	}
	if value, ok := propertyVector(source.Properties, "ADBE Vector Repeater Scale", 2); ok {
		if err := transform.SetScale([2]float64{value[0], value[1]}); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Repeater Rotation"); ok {
		if err := transform.SetRotation(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Repeater Opacity 1"); ok {
		if err := transform.SetStartOpacity(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Repeater Opacity 2"); ok {
		if err := transform.SetEndOpacity(value); err != nil {
			return err
		}
	}
	return nil
}
