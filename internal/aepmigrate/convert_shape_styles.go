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
	if value, ok := propertyVector(source.Properties, "ADBE Vector Fill Color", 4); ok {
		if err := fill.SetColor(profileARGBToRGBA(value)); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Fill Opacity"); ok {
		if err := fill.SetOpacity(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Blend Mode"); ok {
		if err := fill.SetBlendMode(aep.ShapeBlendMode(int(value))); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Composite Order"); ok {
		if err := fill.SetCompositeOrder(aep.ShapeCompositeOrder(int(value))); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Fill Rule"); ok {
		if err := fill.SetFillRule(aep.FillRule(int(value))); err != nil {
			return err
		}
	}
	return nil
}

func materializeShapeStroke(shapeLayer *aep.ShapeLayer, source profile.Layer) error {
	stroke, err := shapeLayer.RootGroup().AddStroke()
	if err != nil {
		return err
	}
	if value, ok := propertyVector(source.Properties, "ADBE Vector Stroke Color", 4); ok {
		if err := stroke.SetColor(profileARGBToRGBA(value)); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Stroke Opacity"); ok {
		if err := stroke.SetOpacity(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Stroke Width"); ok {
		if err := stroke.SetWidth(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Stroke Line Cap"); ok {
		if err := stroke.SetLineCap(aep.StrokeLineCap(int(value))); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Stroke Line Join"); ok {
		if err := stroke.SetLineJoin(aep.StrokeLineJoin(int(value))); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Stroke Miter Limit"); ok {
		if err := stroke.SetMiterLimit(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Blend Mode"); ok {
		if err := stroke.SetBlendMode(aep.ShapeBlendMode(int(value))); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Composite Order"); ok {
		if err := stroke.SetCompositeOrder(aep.ShapeCompositeOrder(int(value))); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Stroke Dash 1"); ok {
		if err := stroke.Dashes().SetDash(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Stroke Gap 1"); ok {
		if err := stroke.Dashes().SetGap(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Taper Start Length"); ok {
		if err := stroke.Taper().SetStartLength(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Taper End Length"); ok {
		if err := stroke.Taper().SetEndLength(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Taper Start Width"); ok {
		if err := stroke.Taper().SetStartWidth(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Taper End Width"); ok {
		if err := stroke.Taper().SetEndWidth(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Taper Start Ease"); ok {
		if err := stroke.Taper().SetStartEase(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Taper End Ease"); ok {
		if err := stroke.Taper().SetEndEase(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Taper Wave Amount"); ok {
		if err := stroke.Wave().SetAmount(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Taper Wavelength"); ok {
		if err := stroke.Wave().SetWavelength(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Taper Wave Phase"); ok {
		if err := stroke.Wave().SetPhase(value); err != nil {
			return err
		}
	}
	return nil
}
