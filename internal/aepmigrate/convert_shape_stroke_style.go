package aepmigrate

import (
	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

func materializeShapeStroke(shapeLayer *aep.ShapeLayer, source profile.Layer) error {
	stroke, err := shapeLayer.RootGroup().AddStroke()
	if err != nil {
		return err
	}
	if err := materializeShapeStrokeBase(stroke, source); err != nil {
		return err
	}
	if err := materializeShapeStrokeDashes(stroke, source); err != nil {
		return err
	}
	if err := materializeShapeStrokeTaper(stroke, source); err != nil {
		return err
	}
	return materializeShapeStrokeWave(stroke, source)
}

func materializeShapeStrokeBase(stroke *aep.StrokeNode, source profile.Layer) error {
	if err := applyShapeARGBProperty(source.Properties, "ADBE Vector Stroke Color", stroke.SetColor); err != nil {
		return err
	}
	if err := applyShapeFloatProperty(source.Properties, "ADBE Vector Stroke Opacity", stroke.SetOpacity); err != nil {
		return err
	}
	if err := applyShapeFloatProperty(source.Properties, "ADBE Vector Stroke Width", stroke.SetWidth); err != nil {
		return err
	}
	if err := applyShapeFloatProperty(source.Properties, "ADBE Vector Stroke Line Cap", func(value float64) error {
		return stroke.SetLineCap(aep.StrokeLineCap(int(value)))
	}); err != nil {
		return err
	}
	if err := applyShapeFloatProperty(source.Properties, "ADBE Vector Stroke Line Join", func(value float64) error {
		return stroke.SetLineJoin(aep.StrokeLineJoin(int(value)))
	}); err != nil {
		return err
	}
	if err := applyShapeFloatProperty(source.Properties, "ADBE Vector Stroke Miter Limit", stroke.SetMiterLimit); err != nil {
		return err
	}
	if err := applyShapeFloatProperty(source.Properties, "ADBE Vector Blend Mode", func(value float64) error {
		return stroke.SetBlendMode(aep.ShapeBlendMode(int(value)))
	}); err != nil {
		return err
	}
	if err := applyShapeFloatProperty(source.Properties, "ADBE Vector Composite Order", func(value float64) error {
		return stroke.SetCompositeOrder(aep.ShapeCompositeOrder(int(value)))
	}); err != nil {
		return err
	}
	return nil
}
