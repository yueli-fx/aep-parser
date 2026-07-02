package aepmigrate

import (
	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

func materializeShapeRoundCorners(shapeLayer *aep.ShapeLayer, source profile.Layer) error {
	roundCorners, err := shapeLayer.RootGroup().AddRoundCorners()
	if err != nil {
		return err
	}
	if err := applyShapeFloatProperty(source.Properties, "ADBE Vector RoundCorner Radius", roundCorners.SetRadius); err != nil {
		return err
	}
	return nil
}

func materializeShapeOffsetPaths(shapeLayer *aep.ShapeLayer, source profile.Layer) error {
	offsetPaths, err := shapeLayer.RootGroup().AddOffsetPaths()
	if err != nil {
		return err
	}
	if err := applyShapeFloatProperty(source.Properties, "ADBE Vector Offset Amount", offsetPaths.SetAmount); err != nil {
		return err
	}
	if err := applyShapeFloatProperty(source.Properties, "ADBE Vector Offset Line Join", func(value float64) error {
		return offsetPaths.SetLineJoin(aep.StrokeLineJoin(int(value)))
	}); err != nil {
		return err
	}
	if err := applyShapeFloatProperty(source.Properties, "ADBE Vector Offset Miter Limit", offsetPaths.SetMiterLimit); err != nil {
		return err
	}
	if err := applyShapeFloatProperty(source.Properties, "ADBE Vector Offset Copies", offsetPaths.SetCopies); err != nil {
		return err
	}
	if err := applyShapeFloatProperty(source.Properties, "ADBE Vector Offset Copy Offset", offsetPaths.SetCopyOffset); err != nil {
		return err
	}
	return nil
}

func materializeShapeTrim(shapeLayer *aep.ShapeLayer, source profile.Layer) error {
	trim, err := shapeLayer.RootGroup().AddTrim()
	if err != nil {
		return err
	}
	if err := applyShapeFloatProperty(source.Properties, "ADBE Vector Trim Start", trim.SetStart); err != nil {
		return err
	}
	if err := applyShapeFloatProperty(source.Properties, "ADBE Vector Trim End", trim.SetEnd); err != nil {
		return err
	}
	if err := applyShapeFloatProperty(source.Properties, "ADBE Vector Trim Offset", trim.SetOffset); err != nil {
		return err
	}
	if err := applyShapeFloatProperty(source.Properties, "ADBE Vector Trim Type", func(value float64) error {
		return trim.SetType(aep.TrimType(int(value)))
	}); err != nil {
		return err
	}
	return nil
}
