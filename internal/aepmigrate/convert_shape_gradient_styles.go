package aepmigrate

import (
	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

func materializeShapeGradientFill(shapeLayer *aep.ShapeLayer, source profile.Layer) error {
	fill, err := shapeLayer.RootGroup().AddGradientFill()
	if err != nil {
		return err
	}
	if err := applyShapeFloatProperty(source.Properties, "ADBE Vector Grad Type", func(value float64) error {
		return fill.SetGradientType(aep.GradientType(int(value)))
	}); err != nil {
		return err
	}
	if err := applyShapeVector2Property(source.Properties, "ADBE Vector Grad Start Pt", fill.SetStartPoint); err != nil {
		return err
	}
	if err := applyShapeVector2Property(source.Properties, "ADBE Vector Grad End Pt", fill.SetEndPoint); err != nil {
		return err
	}
	if err := applyShapeFloatProperty(source.Properties, "ADBE Vector Grad HiLite Length", fill.SetHighlightLength); err != nil {
		return err
	}
	if err := applyShapeFloatProperty(source.Properties, "ADBE Vector Grad HiLite Angle", fill.SetHighlightAngle); err != nil {
		return err
	}
	if err := applyShapeGradientStops(source.Properties, fill); err != nil {
		return err
	}
	return nil
}
