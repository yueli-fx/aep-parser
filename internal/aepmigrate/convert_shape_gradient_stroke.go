package aepmigrate

import (
	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

func materializeShapeGradientStroke(shapeLayer *aep.ShapeLayer, source profile.Layer) error {
	stroke, err := shapeLayer.RootGroup().AddGradientStroke()
	if err != nil {
		return err
	}
	if err := applyShapeFloatProperty(source.Properties, "ADBE Vector Grad Type", func(value float64) error {
		return stroke.SetGradientType(aep.GradientType(int(value)))
	}); err != nil {
		return err
	}
	if err := applyShapeVector2Property(source.Properties, "ADBE Vector Grad Start Pt", stroke.SetStartPoint); err != nil {
		return err
	}
	if err := applyShapeVector2Property(source.Properties, "ADBE Vector Grad End Pt", stroke.SetEndPoint); err != nil {
		return err
	}
	if err := applyShapeFloatProperty(source.Properties, "ADBE Vector Grad HiLite Length", stroke.SetHighlightLength); err != nil {
		return err
	}
	if err := applyShapeFloatProperty(source.Properties, "ADBE Vector Grad HiLite Angle", stroke.SetHighlightAngle); err != nil {
		return err
	}
	if err := applyShapeGradientStops(source.Properties, stroke); err != nil {
		return err
	}
	if err := applyShapeFloatProperty(source.Properties, "ADBE Vector Stroke Width", stroke.SetStrokeWidth); err != nil {
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
	return nil
}
