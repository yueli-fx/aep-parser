package aepmigrate

import (
	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

func materializeShapeRepeater(shapeLayer *aep.ShapeLayer, source profile.Layer) error {
	repeater, err := shapeLayer.RootGroup().AddRepeater()
	if err != nil {
		return err
	}
	if err := applyShapeFloatProperty(source.Properties, "ADBE Vector Repeater Copies", repeater.SetCopies); err != nil {
		return err
	}
	if err := applyShapeFloatProperty(source.Properties, "ADBE Vector Repeater Offset", repeater.SetOffset); err != nil {
		return err
	}
	if err := applyShapeFloatProperty(source.Properties, "ADBE Vector Repeater Order", func(value float64) error {
		return repeater.SetOrder(aep.RepeaterOrder(int(value)))
	}); err != nil {
		return err
	}
	transform := repeater.Transform()
	if err := applyShapeVector2Property(source.Properties, "ADBE Vector Repeater Anchor", transform.SetAnchor); err != nil {
		return err
	}
	if err := applyShapeVector2Property(source.Properties, "ADBE Vector Repeater Position", transform.SetPosition); err != nil {
		return err
	}
	if err := applyShapeVector2Property(source.Properties, "ADBE Vector Repeater Scale", transform.SetScale); err != nil {
		return err
	}
	if err := applyShapeFloatProperty(source.Properties, "ADBE Vector Repeater Rotation", transform.SetRotation); err != nil {
		return err
	}
	if err := applyShapeFloatProperty(source.Properties, "ADBE Vector Repeater Opacity 1", transform.SetStartOpacity); err != nil {
		return err
	}
	if err := applyShapeFloatProperty(source.Properties, "ADBE Vector Repeater Opacity 2", transform.SetEndOpacity); err != nil {
		return err
	}
	return nil
}
