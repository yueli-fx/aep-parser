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
	if err := applyShapeVector2Property(source.Properties, "ADBE Vector Wiggler Anchor", transform.SetAnchor); err != nil {
		return err
	}
	if err := applyShapeVector2Property(source.Properties, "ADBE Vector Wiggler Position", transform.SetPosition); err != nil {
		return err
	}
	if err := applyShapeVector2Property(source.Properties, "ADBE Vector Wiggler Scale", transform.SetScale); err != nil {
		return err
	}
	if err := applyShapeFloatProperty(source.Properties, "ADBE Vector Wiggler Rotation", transform.SetRotation); err != nil {
		return err
	}
	if err := applyShapeFloatProperty(source.Properties, "ADBE Vector Xform Temporal Freq", wiggleTransform.SetWigglesPerSecond); err != nil {
		return err
	}
	if err := applyShapeFloatProperty(source.Properties, "ADBE Vector Random Seed", wiggleTransform.SetRandomSeed); err != nil {
		return err
	}
	if err := applyShapeFloatProperty(source.Properties, "ADBE Vector Correlation", wiggleTransform.SetCorrelation); err != nil {
		return err
	}
	if err := applyShapeFloatProperty(source.Properties, "ADBE Vector Temporal Phase", wiggleTransform.SetTemporalPhase); err != nil {
		return err
	}
	if err := applyShapeFloatProperty(source.Properties, "ADBE Vector Spatial Phase", wiggleTransform.SetSpatialPhase); err != nil {
		return err
	}
	return nil
}
