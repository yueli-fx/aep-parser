package aepmigrate

import (
	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

func materializeShapeFill(shapeLayer *aep.ShapeLayer, source profile.Layer) error {
	fill, err := shapeLayer.RootGroup().AddFill()
	if err != nil {
		return err
	}
	if err := applyShapeARGBProperty(source.Properties, "ADBE Vector Fill Color", fill.SetColor); err != nil {
		return err
	}
	if err := applyShapeFloatProperty(source.Properties, "ADBE Vector Fill Opacity", fill.SetOpacity); err != nil {
		return err
	}
	if err := applyShapeFloatProperty(source.Properties, "ADBE Vector Blend Mode", func(value float64) error {
		return fill.SetBlendMode(aep.ShapeBlendMode(int(value)))
	}); err != nil {
		return err
	}
	if err := applyShapeFloatProperty(source.Properties, "ADBE Vector Composite Order", func(value float64) error {
		return fill.SetCompositeOrder(aep.ShapeCompositeOrder(int(value)))
	}); err != nil {
		return err
	}
	if err := applyShapeFloatProperty(source.Properties, "ADBE Vector Fill Rule", func(value float64) error {
		return fill.SetFillRule(aep.FillRule(int(value)))
	}); err != nil {
		return err
	}
	return nil
}
